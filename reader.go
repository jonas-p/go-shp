package shp

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strings"
)

// Reader provides a interface for reading Shapefiles. Calls
// to the Next method will iterate through the objects in the
// Shapefile. After a call to Next the object will be available
// through the Shape method.
type Reader struct {
	GeometryType ShapeType
	bbox         Box
	err          error

	shp        *os.File
	shape      Shape
	num        int32
	filename   string
	filelength int64

	dbf             *os.File
	dbfFields       []Field
	dbfNumRecords   int32
	dbfHeaderLength int16
	dbfRecordLength int16
}

// Open opens a Shapefile for reading.
func Open(filename string) (*Reader, error) {
	filename = filename[0 : len(filename)-3]
	shp, err := os.Open(filename + "shp")
	if err != nil {
		return nil, err
	}
	s := &Reader{filename: filename, shp: shp}
	s.readHeaders()
	return s, nil
}

func (r *Reader) BBox() Box {
	return r.bbox
}

// Read and parse headers in the Shapefile. This will
// fill out GeometryType, filelength and bbox.
func (r *Reader) readHeaders() {
	// don't trust the the filelength in the header
	r.filelength, _ = r.shp.Seek(0, os.SEEK_END)

	var filelength int32
	r.shp.Seek(24, 0)
	// file length
	binary.Read(r.shp, binary.BigEndian, &filelength)
	r.shp.Seek(32, 0)
	binary.Read(r.shp, binary.LittleEndian, &r.GeometryType)
	r.bbox.MinX = readFloat64(r.shp)
	r.bbox.MinY = readFloat64(r.shp)
	r.bbox.MaxX = readFloat64(r.shp)
	r.bbox.MaxY = readFloat64(r.shp)
	r.shp.Seek(100, 0)
}

func readFloat64(r io.Reader) float64 {
	var bits uint64
	binary.Read(r, binary.LittleEndian, &bits)
	return math.Float64frombits(bits)
}

// Close closes the Shapefile.
func (r *Reader) Close() error {
	if r.err == nil {
		r.err = r.shp.Close()
		if r.dbf != nil {
			r.dbf.Close()
		}
	}
	return r.err
}

// Shape returns the most recent feature that was read by
// a call to Next. It returns two values, the int is the
// object index starting from zero in the shapefile which
// can be used as row in ReadAttribute, and the Shape is the object.
func (r *Reader) Shape() (int, Shape) {
	return int(r.num) - 1, r.shape
}

// Attribute returns value of the n-th attribute of the most recent feature
// that was read by a call to Next.
func (r *Reader) Attribute(n int) string {
	return r.ReadAttribute(int(r.num)-1, n)
}

// newShape creates a new shape with a given type.
func newShape(shapetype ShapeType) Shape {
	switch shapetype {
	case NULL:
		return new(Null)
	case POINT:
		return new(Point)
	case POLYLINE:
		return new(PolyLine)
	case POLYGON:
		return new(Polygon)
	case MULTIPOINT:
		return new(MultiPoint)
	case POINTZ:
		return new(PointZ)
	case POLYLINEZ:
		return new(PolyLineZ)
	case POLYGONZ:
		return new(PolygonZ)
	case MULTIPOINTZ:
		return new(MultiPointZ)
	case POINTM:
		return new(PointM)
	case POLYLINEM:
		return new(PolyLineM)
	case POLYGONM:
		return new(PolygonM)
	case MULTIPOINTM:
		return new(MultiPointM)
	case MULTIPATCH:
		return new(MultiPatch)
	default:
		log.Fatal("Unsupported shape type:", shapetype)
		return nil
	}
}

// Next reads in the next Shape in the Shapefile, which
// will then be available through the Shape method. It
// returns false when the reader has reached the end of the
// file.
func (r *Reader) Next() bool {
	cur, _ := r.shp.Seek(0, os.SEEK_CUR)
	if cur >= r.filelength {
		return false
	}

	var size int32
	var shapetype ShapeType
	er := &errReader{Reader: r.shp}
	binary.Read(er, binary.BigEndian, &r.num)
	binary.Read(er, binary.BigEndian, &size)
	binary.Read(er, binary.LittleEndian, &shapetype)
	if er.e != nil {
		r.err = fmt.Errorf("Error when reading metadata of next shape: %v", er.e)
		return false
	}

	r.shape = newShape(shapetype)
	r.shape.read(er)
	if er.e != nil {
		r.err = fmt.Errorf("Error while reading next shape: %v", er.e)
		return false
	}

	// move to next object
	r.shp.Seek(int64(size)*2+cur+8, 0)
	return true
}

// Opens DBF file using r.filename + "dbf". This method
// will parse the header and fill out all dbf* values int
// the f object.
func (r *Reader) openDbf() (err error) {
	if r.dbf != nil {
		return
	}

	r.dbf, err = os.Open(r.filename + "dbf")
	if err != nil {
		return
	}

	// read header
	r.dbf.Seek(4, os.SEEK_SET)
	binary.Read(r.dbf, binary.LittleEndian, &r.dbfNumRecords)
	binary.Read(r.dbf, binary.LittleEndian, &r.dbfHeaderLength)
	binary.Read(r.dbf, binary.LittleEndian, &r.dbfRecordLength)

	r.dbf.Seek(20, os.SEEK_CUR) // skip padding
	numFields := int(math.Floor(float64(r.dbfHeaderLength-33) / 32.0))
	r.dbfFields = make([]Field, numFields)
	binary.Read(r.dbf, binary.LittleEndian, &r.dbfFields)

	return
}

// Fields returns a slice of Fields that are present in the
// DBF table.
func (r *Reader) Fields() []Field {
	r.openDbf() // make sure we have dbf file to read from
	return r.dbfFields
}

// Err returns the last non-EOF error encountered.
func (r *Reader) Err() error {
	if r.err == io.EOF {
		return nil
	}
	return r.err
}

// AttributeCount returns number of records in the DBF table.
func (r *Reader) AttributeCount() int {
	r.openDbf() // make sure we have a dbf file to read from
	return int(r.dbfNumRecords)
}

// ReadAttribute returns the attribute value at row for field in
// the DBF table as a string. Both values starts at 0.
func (r *Reader) ReadAttribute(row int, field int) string {
	r.openDbf() // make sure we have a dbf file to read from
	seekTo := 1 + int64(r.dbfHeaderLength) + (int64(row) * int64(r.dbfRecordLength))
	for n := 0; n < field; n++ {
		seekTo += int64(r.dbfFields[n].Size)
	}
	r.dbf.Seek(seekTo, os.SEEK_SET)
	buf := make([]byte, r.dbfFields[field].Size)
	r.dbf.Read(buf)
	return strings.Trim(string(buf[:]), " ")
}
