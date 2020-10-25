package shp

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strings"
)

const magic int32 = 0x0000270a

// Reader provides a interface for reading Shapefiles. Calls
// to the Next method will iterate through the objects in the
// Shapefile. After a call to Next the object will be available
// through the Shape method.
type Reader struct {
	GeometryType ShapeType
	bbox         Box
	err          error

	//shpFs      io.ReadCloser
	shp   io.Reader
	shape Shape
	num   int32
	//filename   string
	filelength int64

	dbf             io.Reader //readSeekCloser
	dbfSeek         io.ReadSeeker
	dbfFields       []Field
	dbfNumRecords   int32
	dbfHeaderLength int16
	dbfRecordLength int16
}

type readSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

// ProviderConfigurator is the Reader configurator
type ProviderConfigurator func(*Reader) error

// New Creates a Reader from streams
func New(shp io.Reader, conf ...ProviderConfigurator) (*Reader, error) {
	if shp == nil {
		return nil, fmt.Errorf("missing shp reader")
	}

	s := &Reader{shp: shp}

	for _, cnf := range conf {
		if err := cnf(s); err != nil {
			return nil, err
		}
	}

	return s, s.readHeaders()
}

// WithDBF appends a io.Reader as DBF source
func WithDBF(dbf io.Reader) ProviderConfigurator {
	return func(r *Reader) error {
		if dbf == nil {
			return fmt.Errorf("missing reader")
		}

		if r.dbfSeek != nil {
			return fmt.Errorf("you can only provide one DBF source")
		}

		r.dbf = dbf
		return nil
	}
}

// WithSeekableDBF appends a io.ReadSeeker as DBF source
func WithSeekableDBF(dbf io.ReadSeeker) ProviderConfigurator {
	return func(r *Reader) error {
		if dbf == nil {
			return fmt.Errorf("missing readseeker")
		}

		if r.dbf != nil {
			return fmt.Errorf("you can only provide one DBF source")
		}

		r.dbfSeek = dbf
		return nil
	}
}

// BBox returns the bounding box of the shapefile.
func (r *Reader) BBox() Box {
	return r.bbox
}

// Read and parse headers in the Shapefile. This will
// fill out GeometryType, filelength and bbox.
func (r *Reader) readHeaders() error {
	er := &errReader{Reader: r.shp}
	var (
		m          int32
		filelength int32
	)

	// Read magic
	err := binary.Read(er, binary.BigEndian, &m)
	if err != nil {
		return err
	}

	if magic != m {
		return fmt.Errorf("wrong magic, expected %04x, got %04x", magic, m)
	}

	// skip next 5 uint32
	_, _ = r.shp.Read(make([]byte, 20))

	err = binary.Read(er, binary.BigEndian, &filelength)
	if err != nil {
		return err
	}
	r.filelength = int64(filelength)

	// skip version (int32)
	_, _ = r.shp.Read(make([]byte, 4))

	// Read Type
	err = binary.Read(er, binary.LittleEndian, &r.GeometryType)
	if err != nil {
		return err
	}

	r.bbox.MinX = readFloat64(er)
	r.bbox.MinY = readFloat64(er)
	r.bbox.MaxX = readFloat64(er)
	r.bbox.MaxY = readFloat64(er)

	// skip next 32 bytes
	_, _ = r.shp.Read(make([]byte, 32))
	return er.e
}

func readFloat64(r io.Reader) float64 {
	var bits uint64
	binary.Read(r, binary.LittleEndian, &bits)
	return math.Float64frombits(bits)
}

// Close closes the Shapefile.
func (r *Reader) Close() error {
	/*if r.err == nil && r.shpFs != nil {
		r.err = r.shpFs.Close()
		if r.dbf != nil {
			r.dbf.Close()
		}
	}
	return r.err*/
	return nil
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
func newShape(shapetype ShapeType) (Shape, error) {
	switch shapetype {
	case NULL:
		return new(Null), nil
	case POINT:
		return new(Point), nil
	case POLYLINE:
		return new(PolyLine), nil
	case POLYGON:
		return new(Polygon), nil
	case MULTIPOINT:
		return new(MultiPoint), nil
	case POINTZ:
		return new(PointZ), nil
	case POLYLINEZ:
		return new(PolyLineZ), nil
	case POLYGONZ:
		return new(PolygonZ), nil
	case MULTIPOINTZ:
		return new(MultiPointZ), nil
	case POINTM:
		return new(PointM), nil
	case POLYLINEM:
		return new(PolyLineM), nil
	case POLYGONM:
		return new(PolygonM), nil
	case MULTIPOINTM:
		return new(MultiPointM), nil
	case MULTIPATCH:
		return new(MultiPatch), nil
	default:
		return nil, fmt.Errorf("Unsupported shape type: %v", shapetype)
	}
}

// Next reads in the next Shape in the Shapefile, which
// will then be available through the Shape method. It
// returns false when the reader has reached the end of the
// file or encounters an error.
func (r *Reader) Next() bool {
	var size int32
	var shapetype ShapeType
	er := &errReader{Reader: r.shp}

	_ = binary.Read(er, binary.BigEndian, &r.num)
	_ = binary.Read(er, binary.BigEndian, &size) // size counts the 16-bit words
	_ = binary.Read(er, binary.LittleEndian, &shapetype)

	if er.e != nil {
		if er.e != io.EOF {
			r.err = fmt.Errorf("Error when reading metadata of next shape: %v", er.e)
		} else {
			r.err = io.EOF
		}
		return false
	}

	var err error
	r.shape, err = newShape(shapetype)
	if err != nil {
		r.err = fmt.Errorf("Error decoding shape type: %v", err)
		return false
	}
	count := r.shape.read(er)
	if er.e != nil {
		r.err = fmt.Errorf("Error while reading next shape: %v", er.e)
		return false
	}

	offset := 2*size - int32(count) - 4
	if offset < 0 {
		r.err = fmt.Errorf("too many bytes were read")
		return false
	}

	// move to next object
	if offset > 0 {
		_, err = r.shp.Read(make([]byte, offset))
		if err != nil && err != io.EOF {
			r.err = err
		}
		if err != nil {
			return false
		}
	}

	return true
}

// Opens DBF file using r.filename + "dbf". This method
// will parse the header and fill out all dbf* values int
// the f object.
func (r *Reader) openDbf() (err error) {
	if r.dbf == nil && r.dbfSeek == nil {
		return fmt.Errorf("missing DBF")
	}

	dbf := r.dbf
	if dbf == nil {
		dbf = r.dbfSeek
	}

	// skip next 4 bytes
	_, _ = r.shp.Read(make([]byte, 4))

	// read header
	binary.Read(dbf, binary.LittleEndian, &r.dbfNumRecords)
	binary.Read(dbf, binary.LittleEndian, &r.dbfHeaderLength)
	binary.Read(dbf, binary.LittleEndian, &r.dbfRecordLength)

	// skip padding
	_, _ = r.shp.Read(make([]byte, 20))

	numFields := int(math.Floor(float64(r.dbfHeaderLength-33) / 32.0))
	r.dbfFields = make([]Field, numFields)
	binary.Read(dbf, binary.LittleEndian, &r.dbfFields)
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
	if r.dbfSeek == nil {
		return ""
	}

	r.openDbf() // make sure we have a dbf file to read from
	seekTo := 1 + int64(r.dbfHeaderLength) + (int64(row) * int64(r.dbfRecordLength))
	for n := 0; n < field; n++ {
		seekTo += int64(r.dbfFields[n].Size)
	}
	r.dbfSeek.Seek(seekTo, io.SeekStart)
	buf := make([]byte, r.dbfFields[field].Size)
	r.dbfSeek.Read(buf)
	return strings.Trim(string(buf[:]), " ")
}
