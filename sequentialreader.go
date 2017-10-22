package shp

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"strings"
)

// SequentialReader is the interface that allows reading shapes and attributes one after another. It also embeds io.Closer.
type SequentialReader interface {
	// Close() frees the resources allocated by the SequentialReader.
	io.Closer

	// Next() tries to advance the reading by one shape and one attribute row
	// and returns true if the read operation could be performed without any
	// error.
	Next() bool

	// Shape returns the index and the last read shape. If the SequentialReader
	// encountered any errors, nil is returned for the Shape.
	Shape() (int, Shape)

	// Attribute returns the value of the n-th attribute in the current row. If
	// the SequentialReader encountered any errors, the empty string is
	// returned.
	Attribute(n int) string

	// Fields returns the fields of the database. If the SequentialReader
	// encountered any errors, nil is returned.
	Fields() []Field

	// Err returns the last non-EOF error encountered.
	Err() error
}

// Attributes returns all attributes of the shape that sr was last advanced to.
func Attributes(sr SequentialReader) []string {
	if sr.Err() != nil {
		return nil
	}
	s := make([]string, len(sr.Fields()))
	for i := range s {
		s[i] = sr.Attribute(i)
	}
	return s
}

// AttributeCount returns the number of fields of the database.
func AttributeCount(sr SequentialReader) int {
	return len(sr.Fields())
}

// seqReader implements SequentialReader based on external io.ReadCloser
// instances
type seqReader struct {
	shp, dbf io.ReadCloser
	err      error

	geometryType ShapeType
	bbox         Box

	shape      Shape
	num        int32
	filelength int64

	dbfFields       []Field
	dbfNumRecords   int32
	dbfHeaderLength int16
	dbfRecordLength int16
	dbfRow          []byte
}

// Read and parse headers in the Shapefile. This will fill out GeometryType,
// filelength and bbox.
func (sr *seqReader) readHeaders() {
	// contrary to Reader.readHeaders we cannot seek with the ReadCloser, so we
	// need to trust the filelength in the header

	er := &errReader{Reader: sr.shp}
	// shp headers
	io.CopyN(ioutil.Discard, er, 24)
	var l int32
	binary.Read(er, binary.BigEndian, &l)
	sr.filelength = int64(l) * 2
	io.CopyN(ioutil.Discard, er, 4)
	binary.Read(er, binary.LittleEndian, &sr.geometryType)
	sr.bbox.MinX = readFloat64(er)
	sr.bbox.MinY = readFloat64(er)
	sr.bbox.MaxX = readFloat64(er)
	sr.bbox.MaxY = readFloat64(er)
	io.CopyN(ioutil.Discard, er, 32) // skip four float64: Zmin, Zmax, Mmin, Max
	if er.e != nil {
		sr.err = fmt.Errorf("Error when reading SHP header: %v", er.e)
		return
	}

	// dbf header
	er = &errReader{Reader: sr.dbf}
	if sr.dbf == nil {
		return
	}
	io.CopyN(ioutil.Discard, er, 4)
	binary.Read(er, binary.LittleEndian, &sr.dbfNumRecords)
	binary.Read(er, binary.LittleEndian, &sr.dbfHeaderLength)
	binary.Read(er, binary.LittleEndian, &sr.dbfRecordLength)
	io.CopyN(ioutil.Discard, er, 20) // skip padding
	numFields := int(math.Floor(float64(sr.dbfHeaderLength-33) / 32.0))
	sr.dbfFields = make([]Field, numFields)
	binary.Read(er, binary.LittleEndian, &sr.dbfFields)
	buf := make([]byte, 1)
	er.Read(buf[:])
	if er.e != nil {
		sr.err = fmt.Errorf("Error when reading DBF header: %v", er.e)
		return
	}
	if buf[0] != 0x0d {
		sr.err = fmt.Errorf("Field descriptor array terminator not found")
		return
	}
	sr.dbfRow = make([]byte, sr.dbfRecordLength)
}

// Next implements a method of interface SequentialReader for seqReader.
func (sr *seqReader) Next() bool {
	if sr.err != nil {
		return false
	}
	var num, size int32
	var shapetype ShapeType

	// read shape
	er := &errReader{Reader: sr.shp}
	binary.Read(er, binary.BigEndian, &num)
	binary.Read(er, binary.BigEndian, &size)
	binary.Read(er, binary.LittleEndian, &shapetype)

	if er.e != nil {
		if er.e != io.EOF {
			sr.err = fmt.Errorf("Error when reading shapefile header: %v", er.e)
		} else {
			sr.err = io.EOF
		}
		return false
	}
	sr.num = num
	var err error
	sr.shape, err = newShape(shapetype)
	if err != nil {
		sr.err = fmt.Errorf("Error decoding shape type: %v", err)
		return false
	}
	sr.shape.read(er)
	switch {
	case er.e == io.EOF:
		// io.EOF means end-of-file was reached gracefully after all
		// shape-internal reads succeeded, so it's not a reason stop
		// iterating over all shapes.
		er.e = nil
	case er.e != nil:
		sr.err = fmt.Errorf("Error while reading next shape: %v", er.e)
		return false
	}
	skipBytes := int64(size)*2 + 8 - er.n
	_, ce := io.CopyN(ioutil.Discard, er, skipBytes)
	if er.e != nil {
		sr.err = er.e
		return false
	}
	if ce != nil {
		sr.err = fmt.Errorf("Error when discarding bytes on sequential read: %v", ce)
		return false
	}
	if _, err := io.ReadFull(sr.dbf, sr.dbfRow); err != nil {
		sr.err = fmt.Errorf("Error when reading DBF row: %v", err)
		return false
	}
	if sr.dbfRow[0] != 0x20 && sr.dbfRow[0] != 0x2a {
		sr.err = fmt.Errorf("Attribute row %d starts with incorrect deletion indicator", num)
	}
	return sr.err == nil
}

// Shape implements a method of interface SequentialReader for seqReader.
func (sr *seqReader) Shape() (int, Shape) {
	return int(sr.num) - 1, sr.shape
}

// Attribute implements a method of interface SequentialReader for seqReader.
func (sr *seqReader) Attribute(n int) string {
	if sr.err != nil {
		return ""
	}
	start := 1
	f := 0
	for ; f < n; f++ {
		start += int(sr.dbfFields[f].Size)
	}
	s := string(sr.dbfRow[start : start+int(sr.dbfFields[f].Size)])
	return strings.Trim(s, " ")
}

// Err returns the first non-EOF error that was encountered.
func (sr *seqReader) Err() error {
	if sr.err == io.EOF {
		return nil
	}
	return sr.err
}

// Close closes the seqReader and free all the allocated resources.
func (sr *seqReader) Close() error {
	if err := sr.shp.Close(); err != nil {
		return err
	}
	if err := sr.dbf.Close(); err != nil {
		return err
	}
	return nil
}

// Fields returns a slice of the fields that are present in the DBF table.
func (sr *seqReader) Fields() []Field {
	return sr.dbfFields
}

// SequentialReaderFromExt returns a new SequentialReader that interprets shp
// as a source of shapes whose attributes can be retrieved from dbf.
func SequentialReaderFromExt(shp, dbf io.ReadCloser) SequentialReader {
	sr := &seqReader{shp: shp, dbf: dbf}
	sr.readHeaders()
	return sr
}
