package goshp

import (
	"encoding/binary"
	"log"
	"math"
	"os"
	"strings"
)

type File struct {
	filename     string
	shp          *os.File
	filelength   int64
	GeometryType ShapeType

	dbf             *os.File
	dbfFields       []Field
	dbfNumRecords   int32
	dbfHeaderLength int16
	dbfRecordLength int16
}

// Opens a Shapefile for reading.
func Open(filename string) (*File, error) {
	filename = filename[0 : len(filename)-3]
	shp, err := os.Open(filename + "shp")
	if err != nil {
		return nil, err
	}
	s := &File{filename: filename, shp: shp}
	s.readHeaders()
	return s, nil
}

// Read and parse headers in the Shapefile. This will
// fill out GeometryType and filelength.
func (f *File) readHeaders() {
	// don't trust the the filelength in the header
	f.filelength, _ = f.shp.Seek(0, os.SEEK_END)

	var filelength int32
	f.shp.Seek(24, 0)
	// file length
	binary.Read(f.shp, binary.BigEndian, &filelength)
	f.shp.Seek(32, 0)
	binary.Read(f.shp, binary.LittleEndian, &f.GeometryType)
	f.shp.Seek(100, 0)
}

// Returns true if the file cursor has passed the end
// of the file.
func (f *File) EOF() (ok bool) {
	n, _ := f.shp.Seek(0, os.SEEK_CUR)
	if n >= f.filelength {
		ok = true
	}
	return
}

// Closes the Shapefile
func (f *File) Close() {
	f.shp.Close()
	if f.dbf != nil {
		f.dbf.Close()
	}
}

// Read and returns the next shape in the Shapefile as
// a Shape interface which can be type asserted to the
// correct type.
func (f *File) ReadShape() (shape Shape, err error) {
	var size int32
	var num int32
	var shapetype ShapeType
	binary.Read(f.shp, binary.BigEndian, &num)
	binary.Read(f.shp, binary.BigEndian, &size)
	cur, _ := f.shp.Seek(0, os.SEEK_CUR)
	binary.Read(f.shp, binary.LittleEndian, &shapetype)

	switch shapetype {
	case NULL:
		shape = new(Null)
	case POINT:
		shape = new(Point)
	case POLYLINE:
		shape = new(PolyLine)
	case POLYGON:
		shape = new(Polygon)
	default:
		log.Fatal("Unsupported shape type")
	}
	shape.read(f.shp)

	_, err = f.shp.Seek(int64(size)*2+cur, 0)
	return shape, err
}

// Opens DBF file using f.filename + "dbf". This method
// will parse the header and fill out all dbf* values int
// the f object.
func (f *File) openDbf() {
	if f.dbf != nil {
		return
	}

	var err error
	f.dbf, err = os.Open(f.filename + "dbf")
	if err != nil {
		log.Fatal("Failed to open DBF", err)
	}

	// read header
	f.dbf.Seek(4, os.SEEK_SET)
	binary.Read(f.dbf, binary.LittleEndian, &f.dbfNumRecords)
	binary.Read(f.dbf, binary.LittleEndian, &f.dbfHeaderLength)
	binary.Read(f.dbf, binary.LittleEndian, &f.dbfRecordLength)

	f.dbf.Seek(20, os.SEEK_CUR) // skip padding
	numFields := int(math.Floor(float64(f.dbfHeaderLength-33) / 32.0))
	f.dbfFields = make([]Field, numFields)
	binary.Read(f.dbf, binary.LittleEndian, &f.dbfFields)
}

// Returns number of records in the DBF table
func (f *File) AttributeCount() int {
	f.openDbf() // make sure we have a dbf file to read from
	return int(f.dbfNumRecords)
}

// Read attribute from DBF at row and field
func (f *File) ReadAttribute(row int, field int) string {
	f.openDbf() // make sure we have a dbf file to read from
	seekTo := 1 + int64(f.dbfHeaderLength) + (int64(row) * int64(f.dbfRecordLength))
	for n := 0; n < field; n++ {
		seekTo += int64(f.dbfFields[n].Size)
	}
	f.dbf.Seek(seekTo, os.SEEK_SET)
	buf := make([]byte, f.dbfFields[field].Size)
	f.dbf.Read(buf)
	return strings.Trim(string(buf[:]), " ")
}
