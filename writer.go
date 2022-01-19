package shp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Writer is the type that is used to write a new shapefile.
type Writer struct {
	filename     string
	shp          writeSeekCloser
	shx          writeSeekCloser
	GeometryType ShapeType
	num          int32
	bbox         Box

	dbf             writeSeekCloser
	dbfFields       []Field
	dbfHeaderLength int16
	dbfRecordLength int16
}

type writeSeekCloser interface {
	io.Writer
	io.Seeker
	io.Closer
}

type CreateOptions struct {
	Projection    string
	WithoutDotPrj bool
}

// Create returns a point to new Writer and the first error that was
// encountered. In case an error occurred the returned Writer point will be nil
// This also creates a corresponding SHX file. It is important to use Close()
// when done because that method writes all the headers for each file (SHP, SHX
// and DBF).
// If filename does not end on ".shp" already, it will be treated as the basename
// for the file and the ".shp" extension will be appended to that name.
func Create(filename string, t ShapeType, option *CreateOptions) (*Writer, error) {
	if strings.HasSuffix(strings.ToLower(filename), ".shp") {
		filename = filename[0 : len(filename)-4]
	}
	shp, err := os.Create(filename + ".shp")
	if err != nil {
		return nil, err
	}
	shx, err := os.Create(filename + ".shx")
	if err != nil {
		return nil, err
	}

	if !(option != nil && option.WithoutDotPrj) {
		prj, err := os.Create(filename + ".prj")
		if err != nil {
			return nil, err
		}
		defer prj.Close()

		projection := `GEOGCS["GCS_WGS_1984",DATUM["D_WGS_1984",SPHEROID["WGS_1984",6378137,298.257223563]],PRIMEM["Greenwich",0],UNIT["Degree",0.017453292519943295]]`
		if option != nil && option.Projection != "" {
			projection = option.Projection
		}
		_, err = prj.WriteString(projection)
		if err != nil {
			return nil, err
		}
	}

	shp.Seek(100, io.SeekStart)
	shx.Seek(100, io.SeekStart)
	w := &Writer{
		filename:     filename,
		shp:          shp,
		shx:          shx,
		GeometryType: t,
	}
	return w, nil
}

// Append returns a Writer pointer that will append to the given shapefile and
// the first error that was encounted during creation of that Writer. The
// shapefile must have a valid index file.
func Append(filename string) (*Writer, error) {
	shp, err := os.OpenFile(filename, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	ext := filepath.Ext(filename)
	basename := filename[:len(filename)-len(ext)]
	w := &Writer{
		filename: basename,
		shp:      shp,
	}
	_, err = shp.Seek(32, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("cannot seek to SHP geometry type: %v", err)
	}
	err = binary.Read(shp, binary.LittleEndian, &w.GeometryType)
	if err != nil {
		return nil, fmt.Errorf("cannot read geometry type: %v", err)
	}
	er := &errReader{Reader: shp}
	w.bbox.MinX = readFloat64(er)
	w.bbox.MinY = readFloat64(er)
	w.bbox.MaxX = readFloat64(er)
	w.bbox.MaxY = readFloat64(er)
	if er.e != nil {
		return nil, fmt.Errorf("cannot read bounding box: %v", er.e)
	}

	shx, err := os.OpenFile(basename+".shx", os.O_RDWR, 0666)
	if os.IsNotExist(err) {
		// TODO allow index file to not exist, in that case just
		// read through all the shapes and create it on the fly
	}
	if err != nil {
		return nil, fmt.Errorf("cannot open shapefile index: %v", err)
	}
	_, err = shx.Seek(-8, io.SeekEnd)
	if err != nil {
		return nil, fmt.Errorf("cannot seek to last shape index: %v", err)
	}
	var offset int32
	err = binary.Read(shx, binary.BigEndian, &offset)
	if err != nil {
		return nil, fmt.Errorf("cannot read last shape index: %v", err)
	}
	offset = offset * 2
	_, err = shp.Seek(int64(offset), io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("cannot seek to last shape: %v", err)
	}
	err = binary.Read(shp, binary.BigEndian, &w.num)
	if err != nil {
		return nil, fmt.Errorf("cannot read number of last shape: %v", err)
	}
	_, err = shp.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, fmt.Errorf("cannot seek to SHP end: %v", err)
	}
	_, err = shx.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, fmt.Errorf("cannot seek to SHX end: %v", err)
	}
	w.shx = shx

	dbf, err := os.Open(basename + ".dbf")
	if os.IsNotExist(err) {
		return w, nil // it's okay if the DBF does not exist
	}
	if err != nil {
		return nil, fmt.Errorf("cannot open DBF: %v", err)
	}

	_, err = dbf.Seek(8, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("cannot seek in DBF: %v", err)
	}
	err = binary.Read(dbf, binary.LittleEndian, &w.dbfHeaderLength)
	if err != nil {
		return nil, fmt.Errorf("cannot read header length from DBF: %v", err)
	}
	err = binary.Read(dbf, binary.LittleEndian, &w.dbfRecordLength)
	if err != nil {
		return nil, fmt.Errorf("cannot read record length from DBF: %v", err)
	}

	_, err = dbf.Seek(20, io.SeekCurrent) // skip padding
	if err != nil {
		return nil, fmt.Errorf("cannot seek in DBF: %v", err)
	}
	numFields := int(math.Floor(float64(w.dbfHeaderLength-33) / 32.0))
	w.dbfFields = make([]Field, numFields)
	err = binary.Read(dbf, binary.LittleEndian, &w.dbfFields)
	if err != nil {
		return nil, fmt.Errorf("cannot read number of fields from DBF: %v", err)
	}
	_, err = dbf.Seek(0, io.SeekEnd) // skip padding
	if err != nil {
		return nil, fmt.Errorf("cannot seek to DBF end: %v", err)
	}
	w.dbf = dbf

	return w, nil
}

// Write shape to the Shapefile. This also creates
// a record in the SHX file and DBF file (if it is
// initialized). Returns the index of the written object
// which can be used in WriteAttribute.
func (w *Writer) Write(shape Shape) int {
	// increate bbox
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

	// write empty record to dbf
	if w.dbf != nil {
		w.writeEmptyRecord()
	}

	return int(w.num - 1)
}

// Close closes the Writer. This must be used at the end of
// the transaction because it writes the correct headers
// to the SHP/SHX and DBF files before closing.
func (w *Writer) Close() {
	w.writeHeader(w.shx)
	w.writeHeader(w.shp)
	w.shp.Close()
	w.shx.Close()

	if w.dbf == nil {
		w.SetFields([]Field{})
	}
	w.writeDbfHeader(w.dbf)
	w.dbf.Close()
}

// writeHeader wrires SHP/SHX headers to ws.
func (w *Writer) writeHeader(ws io.WriteSeeker) {
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
func (w *Writer) writeDbfHeader(ws io.WriteSeeker) {
	ws.Seek(0, 0)
	// version, year (YEAR-1990), month, day
	binary.Write(ws, binary.LittleEndian, []byte{3, 24, 5, 3})
	// number of records
	binary.Write(ws, binary.LittleEndian, w.num)
	// header length, record length
	binary.Write(ws, binary.LittleEndian, []int16{w.dbfHeaderLength, w.dbfRecordLength})
	// padding
	binary.Write(ws, binary.LittleEndian, make([]byte, 20))

	for _, field := range w.dbfFields {
		binary.Write(ws, binary.LittleEndian, field)
	}

	// end with return
	ws.Write([]byte("\r"))
}

// SetFields sets field values in the DBF. This initializes the DBF file and
// should be used prior to writing any attributes.
func (w *Writer) SetFields(fields []Field) error {
	if w.dbf != nil {
		return errors.New("Cannot set fields in existing dbf")
	}

	var err error
	w.dbf, err = os.Create(w.filename + ".dbf")
	if err != nil {
		return fmt.Errorf("Failed to open %s.dbf: %v", w.filename, err)
	}
	w.dbfFields = fields

	// calculate record length
	w.dbfRecordLength = int16(1)
	for _, field := range w.dbfFields {
		w.dbfRecordLength += int16(field.Size)
	}

	// header lengh
	w.dbfHeaderLength = int16(len(w.dbfFields)*32 + 33)

	// fill header space with empty bytes for now
	buf := make([]byte, w.dbfHeaderLength)
	binary.Write(w.dbf, binary.LittleEndian, buf)

	// write empty records
	for n := int32(0); n < w.num; n++ {
		w.writeEmptyRecord()
	}
	return nil
}

// Writes an empty record to the end of the DBF. This
// works by seeking to the end of the file and writing
// dbfRecordLength number of bytes. The first byte is a
// space that indicates a new record.
func (w *Writer) writeEmptyRecord() {
	w.dbf.Seek(0, io.SeekEnd)
	buf := make([]byte, w.dbfRecordLength)
	buf[0] = ' '
	binary.Write(w.dbf, binary.LittleEndian, buf)
}

// WriteAttribute writes value for field into the given row in the DBF. Row
// number should be the same as the order the Shape was written to the
// Shapefile. The field value corresponds to the field in the slice used in
// SetFields.
func (w *Writer) WriteAttribute(row int, field int, value interface{}) error {
	var buf []byte
	switch v := value.(type) {
	case int:
		buf = []byte(strconv.Itoa(v))
	case int64:
		buf = []byte(strconv.FormatInt(v, 10))
	case uint:
		buf = []byte(strconv.FormatUint(uint64(v), 10))
	case uint64:
		buf = []byte(strconv.FormatUint(v, 10))
	case float64:
		precision := w.dbfFields[field].Precision
		buf = []byte(strconv.FormatFloat(v, 'f', int(precision), 64))
	case string:
		buf = []byte(v)
	default:
		return fmt.Errorf("Unsupported value type: %T", v)
	}

	if w.dbf == nil {
		return errors.New("Initialize DBF by using SetFields first")
	}
	if sz := int(w.dbfFields[field].Size); len(buf) > sz {
		return fmt.Errorf("Unable to write field %v: %q exceeds field length %v", field, buf, sz)
	}

	seekTo := 1 + int64(w.dbfHeaderLength) + (int64(row) * int64(w.dbfRecordLength))
	for n := 0; n < field; n++ {
		seekTo += int64(w.dbfFields[n].Size)
	}
	w.dbf.Seek(seekTo, io.SeekStart)
	return binary.Write(w.dbf, binary.LittleEndian, buf)
}

// BBox returns the bounding box of the Writer.
func (w *Writer) BBox() Box {
	return w.bbox
}

func (w *Writer) WriteRecord(r Record) error {
	row := w.Write(r.Shape())
	for n, attr := range r.Attributes() {
		err := w.WriteAttribute(row, n, attr.Value())
		if err != nil {
			return err
		}
	}

	return nil
}
