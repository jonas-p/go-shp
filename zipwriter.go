package shp

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultBufferSize = 64 * 1024

// WGS84ProjWKT is projection WKT data for WGS84 coordinate system
const WGS84ProjWKT = `GEOGCS["WGS 84", DATUM["World Geodetic System 1984", SPHEROID["WGS 84", 6378137, 298.257223563, AUTHORITY["EPSG", "7030"]], AUTHORITY["EPSG", "6326"]], PRIMEM["Greenwich", 0, AUTHORITY["EPSG", "8901"]], UNIT["degree", 0.0174532925199433, AUTHORITY["EPSG", "9102"]], AUTHORITY["EPSG", "4326"]]`
const NAD27ProjWKT = `GEOGCS["NAD27", DATUM["North_American_Datum_1927", SPHEROID["Clarke 1866", 6378206.4, 294.9786982138982, AUTHORITY["EPSG", "7008"]], AUTHORITY["EPSG", "6267"]], PRIMEM["Greenwich", 0, AUTHORITY["EPSG","8901"]], UNIT["degree", 0.01745329251994328, AUTHORITY["EPSG", "9122"]], AUTHORITY["EPSG", "4267"]]`

// A byteRWSBuffer is a variable-sized buffer of bytes with Read, Write and Seek methods.
type byteRWSBuffer struct {
	buf []byte
	off int
	sz  int
}

// Write appends the contents of p to the buffer.  Return value
// n is the length of p, err is always null
func (b *byteRWSBuffer) Write(p []byte) (n int, err error) {
	b.grow(b.off + len(p))
	n = copy(b.buf[b.off:], p)
	b.off += n
	return
}

// Read reads the next len(p) bytes from the buffer or until
// the buffer is drained
func (b *byteRWSBuffer) Read(p []byte) (n int, err error) {
	if b.off >= len(b.buf) {
		return 0, io.EOF
	}
	n = copy(p, b.buf[b.off:])
	b.off += n
	return
}

// Seek sets the offset for the next Read or Write to offset, interpreted
// according to whence
func (b *byteRWSBuffer) Seek(offset int64, whence int) (ret int64, err error) {
	switch whence {
	case io.SeekStart:
		ret = 0
	case io.SeekCurrent:
		ret = int64(b.off)
	case io.SeekEnd:
		ret = int64(b.sz)
	default:
		return int64(b.off), io.EOF
	}
	ret += offset
	if ret < 0 {
		return int64(b.off), io.EOF
	}
	b.off = int(ret)
	b.grow(b.off)
	return
}

// Len returns number of bytes allowed to read from slice
func (b *byteRWSBuffer) Len() int {
	return b.sz
}

// Bytes returns the slice of length b.Len() holding
// buffer's underlying byte slice
func (b *byteRWSBuffer) Bytes() []byte {
	return b.buf[:b.sz]
}

// Reset resets buffer to zero length
func (b *byteRWSBuffer) Reset() {
	curCap := cap(b.buf)
	b.buf = nil
	b.buf = make([]byte, curCap)
	b.sz = 0
	b.off = 0
}

// internal
// Grow buffer to guarantee space for n more bytes.
func (b *byteRWSBuffer) grow(n int) {
	if n > cap(b.buf) {
		//buf := make([]byte, n, 2*cap(b.buf)+n)
		buf := make([]byte, 2*cap(b.buf)+n)
		copy(buf, b.buf[0:b.sz])
		b.buf = buf
	}
	if n > b.sz {
		b.sz = n
	}
}

func newByteBuffer(n int) *byteRWSBuffer {
	return &byteRWSBuffer{
		buf: make([]byte, n),
	}
}

// ZipWriter is the type that is used to write a new shapefile.
// It holds all data in memory. ZipWriter is optimized for memory usage
// and archives data on the fly using archive/zip
// It support incremental DBF writes
type ZipWriter struct {
	filename     string
	shp          *byteRWSBuffer
	shx          *byteRWSBuffer
	GeometryType ShapeType
	num          int32
	bbox         Box

	dbf             io.Writer
	dbfRecordsNum   int32
	dbfFields       []Field
	dbfHeaderLength int16
	dbfRecordLength int16

	zbuf *byteRWSBuffer
	zipw *zip.Writer

	headerSent bool
}

// CreateZip creates ZipWriter instance and the first error that was encountered.
// This also creates a corresponding SHX file. It is important to use Close()
// when done because that method writes all the headers for each file (SHP, SHX
// and DBF).
// If filename does not end on ".shp" already, it will be treated as the basename
// for the file and the ".shp" extension will be appended to that name.
func CreateZip(filename string, t ShapeType, n int) (*ZipWriter, error) {
	if strings.HasSuffix(strings.ToLower(filename), ".shp") {
		filename = filename[0 : len(filename)-4]
	}

	shp := newByteBuffer(defaultBufferSize)
	shx := newByteBuffer(defaultBufferSize)

	shp.Seek(100, io.SeekStart)
	shx.Seek(100, io.SeekStart)

	zbuf := newByteBuffer(defaultBufferSize)
	zw := zip.NewWriter(zbuf)
	dbf, err := zw.Create(filename + ".dbf")
	if err != nil {
		return nil, err
	}

	w := &ZipWriter{
		filename:      filename,
		shp:           shp,
		shx:           shx,
		dbf:           dbf,
		zbuf:          zbuf,
		zipw:          zw,
		GeometryType:  t,
		dbfRecordsNum: int32(n),
	}
	return w, nil
}

// UseShapeType does "lazy" shape type initialization when
// type was not known initially
func (w *ZipWriter) UseShapeType(t ShapeType) error {
	if w.GeometryType != NULL {
		return errors.New("ShapeType already initialized")
	}
	w.GeometryType = t
	return nil
}

// GetShapeType returns writer's shape type
func (w *ZipWriter) GetShapeType() ShapeType {
	return w.GeometryType
}

// Write shape to the Shapefile. This also creates
// a record in the SHX file.
// Returns the index of the written object
func (w *ZipWriter) Write(shape Shape) int32 {
	// increase bbox
	if w.num == 0 {
		w.bbox = shape.BBox()
	} else {
		w.bbox.Extend(shape.BBox())
	}

	w.num++
	binary.Write(w.shp, binary.BigEndian, w.num)
	w.shp.Seek(4, io.SeekCurrent)
	start, _ := w.shp.Seek(0, io.SeekCurrent)
	binary.Write(w.shp, binary.LittleEndian, w.GeometryType)
	shape.write(w.shp)
	finish, _ := w.shp.Seek(0, io.SeekCurrent)
	length := int32(math.Floor((float64(finish) - float64(start)) / 2.0))
	w.shp.Seek(start-4, io.SeekStart)
	binary.Write(w.shp, binary.BigEndian, length)
	w.shp.Seek(finish, io.SeekStart)

	// write shx
	binary.Write(w.shx, binary.BigEndian, int32((start-8)/2))
	binary.Write(w.shx, binary.BigEndian, length)

	return w.num - 1
}

// Close closes the Writer. This must be used at the end of
// the transaction because it writes the correct headers
// to the SHP/SHX and adds all files to ZIP archive
func (w *ZipWriter) Close() error {
	w.writeHeader(w.shx)
	w.writeHeader(w.shp)

	// flush SHP data
	shp, err := w.zipw.Create(w.filename + ".shp")
	if err != nil {
		return err
	}
	shp.Write(w.shp.Bytes())

	// flush SHX data
	shx, err := w.zipw.Create(w.filename + ".shx")
	if err != nil {
		return err
	}
	shx.Write(w.shx.Bytes())

	// flush PRJ data
	prj, err := w.zipw.Create(w.filename + ".prj")
	if err != nil {
		return err
	}
	prj.Write([]byte(WGS84ProjWKT))
	return w.zipw.Close()
}

// writeHeader wrires SHP/SHX headers to ws.
func (w *ZipWriter) writeHeader(ws io.WriteSeeker) {
	filelength, _ := ws.Seek(0, io.SeekEnd)
	if filelength == 0 {
		filelength = 100
	}
	ws.Seek(0, io.SeekStart)
	// file code
	binary.Write(ws, binary.BigEndian, []int32{9994, 0, 0, 0, 0, 0})
	// file length
	binary.Write(ws, binary.BigEndian, int32(filelength/2))
	// version and shape type
	binary.Write(ws, binary.LittleEndian, []int32{1000, int32(w.GeometryType)})
	// bounding box
	binary.Write(ws, binary.LittleEndian, w.bbox)
	// elevation, measure
	binary.Write(ws, binary.LittleEndian, []float64{0.0, 0.0, 0.0, 0.0})
}

// writeDbfHeader writes a DBF header to ws.
func (w *ZipWriter) writeDbfHeader(ws io.Writer) {
	//ws.Seek(0, 0)
	// version, year (YEAR-1990), month, day
	binary.Write(ws, binary.LittleEndian, []byte{3, 24, 5, 3})
	// number of records
	binary.Write(ws, binary.LittleEndian, w.dbfRecordsNum)
	// header length, record length
	binary.Write(ws, binary.LittleEndian, []int16{w.dbfHeaderLength, w.dbfRecordLength})
	// padding
	binary.Write(ws, binary.LittleEndian, make([]byte, 20))

	for _, field := range w.dbfFields {
		binary.Write(ws, binary.LittleEndian, field)
	}

	// end with return
	ws.Write([]byte("\r"))

	w.headerSent = true
}

// SetFields sets DBF fields. This should be used prior to writing any attributes.
func (w *ZipWriter) SetFields(fields []Field) error {
	if w.headerSent {
		return errors.New("Header already written")
	}

	w.dbfFields = fields

	// calculate record length
	w.dbfRecordLength = int16(1)
	for _, field := range w.dbfFields {
		w.dbfRecordLength += int16(field.Size)
	}

	// header lengh
	w.dbfHeaderLength = int16(len(w.dbfFields)*32 + 33)

	return nil
}

// BBox returns the bounding box of the Writer.
func (w *ZipWriter) BBox() Box {
	return w.bbox
}

// WriteRecord appends new row to DBF
func (w *ZipWriter) WriteRecord(values []interface{}) error {
	if w.dbf == nil {
		return errors.New("Initialize DBF by using SetFields first")
	}
	if len(values) != len(w.dbfFields) {
		return fmt.Errorf("values count (%d) doesn't match DBF fields count (%d)", len(values), len(w.dbfFields))
	}
	if !w.headerSent {
		w.writeDbfHeader(w.dbf)
	}
	w.dbf.Write([]byte("\x20")) // Record starts with space

	var buf []byte
	var err error
	for i, value := range values {
		sz := int(w.dbfFields[i].Size)
		switch v := value.(type) {
		case int64:
			buf = w.dbfString(strconv.FormatInt(v, 10), sz)
		case float64:
			precision := w.dbfFields[i].Precision
			buf = w.dbfString(strconv.FormatFloat(v, 'f', int(precision), 64), sz)
		case string:
			if w.dbfFields[i].Fieldtype == 'D' {
				buf = w.dbfDate(v, sz)
			} else {
				buf = w.dbfString(v, sz)
			}
		case bool:
			if v {
				buf = w.dbfString("Yes", sz)
			} else {
				buf = w.dbfString("No", sz)
			}
		default:
			return fmt.Errorf("Unsupported value type: %T", v)
		}

		if len(buf) > sz {
			return fmt.Errorf("Unable to write field %v: %q exceeds field length %v", i, buf, sz)
		}

		err = binary.Write(w.dbf, binary.LittleEndian, buf)
		if err != nil {
			break
		}
	}
	return err
}

func (w *ZipWriter) dbfString(str string, size int) []byte {
	var buf []byte
	if len(str) < size { // pad with spaces
		buf = append([]byte(str), bytes.Repeat([]byte{32}, size-len(str))...)
	} else { // truncate
		buf = []byte(str[:size])
	}
	return buf
}

func (w *ZipWriter) dbfDate(str string, size int) []byte {
	var strDate string
	if len(str) == 10 { // 2006-01-02
		t, err := time.Parse("2006-01-02", str)
		if err == nil {
			strDate = t.Format("20060102")
		}
	} else { // DateTime
		t, err := time.Parse(time.RFC3339, str)
		if err == nil {
			strDate = t.Format("20060102")
		}
	}

	return w.dbfString(strDate, size)
}

// Save closes zip writer and writes resulting shapefile archive to disk
func (w *ZipWriter) Save() error {
	err := w.Close()
	if err != nil {
		return err
	}
	f, err := os.Create(w.filename + ".ZIP")
	if err != nil {
		return err
	}
	_, err = f.Write(w.Bytes())
	if err != nil {
		return nil
	}
	return f.Close()
}

// Bytes returnes current content of zip buffer
func (w *ZipWriter) Bytes() []byte {
	w.zipw.Flush()
	return w.zbuf.Bytes()
}

// Reset resets zip buffer
func (w *ZipWriter) Reset() {
	w.zbuf.Reset()
}
