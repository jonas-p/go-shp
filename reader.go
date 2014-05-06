package goshp

import (
	"encoding/binary"
	"os"
)

type File struct {
	file         *os.File
	filelength   int64
	GeometryType ShapeType
}

func Open(filename string) (*File, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	s := &File{file: file}
	s.readHeaders()
	return s, nil
}

func (f *File) readHeaders() {
	// don't trust the the filelength in the header
	f.filelength, _ = f.file.Seek(0, os.SEEK_END)

	var filelength int32
	f.file.Seek(24, 0)
	// file length
	binary.Read(f.file, binary.BigEndian, &filelength)
	f.file.Seek(32, 0)
	binary.Read(f.file, binary.LittleEndian, &f.GeometryType)
	f.file.Seek(100, 0)
}

func (f *File) nextObject() (current int32, position int64) {
	var record [2]int32
	cur, _ := f.file.Seek(0, os.SEEK_CUR)
	binary.Read(f.file, binary.BigEndian, &record)
	current = record[0]
	position = int64(record[1])*2 + cur + 8
	return
}

func (f *File) EOF() (ok bool) {
	n, _ := f.file.Seek(0, os.SEEK_CUR)
	if n >= f.filelength {
		ok = true
	}
	return
}

func (f *File) Close() {
	f.file.Close()
}

func (f *File) ReadShape() (shape Shape, err error) {
	var size int32
	var num int32
	var shapetype ShapeType
	binary.Read(f.file, binary.BigEndian, &num)
	binary.Read(f.file, binary.BigEndian, &size)
	cur, _ := f.file.Seek(0, os.SEEK_CUR)
	binary.Read(f.file, binary.LittleEndian, &shapetype)

	switch shapetype {
	case NULL:
		shape = new(Null)
	case POINT:
		shape = new(Point)
	case POLYLINE:
		shape = new(PolyLine)
	case POLYGON:
		shape = new(Polygon)
	}
	shape.read(f.file)

	_, err = f.file.Seek(int64(size)*2+cur, 0)
	return shape, err
}
