package goshp

import (
	"encoding/binary"
	"math"
	"os"
)

type Writer struct {
	shp          *os.File
	shx          *os.File
	dbf          *os.File
	GeometryType ShapeType
	num          int32
	bbox         Box
	fields       []Field
}

func Create(filename string, t ShapeType) (*Writer, error) {
	filename = filename[0 : len(filename)-3]
	shp, err := os.Create(filename + "shp")
	if err != nil {
		return nil, err
	}
	shx, err := os.Create(filename + "shx")
	if err != nil {
		return nil, err
	}
	dbf, err := os.Create(filename + "dbf")
	if err != nil {
		return nil, err
	}
	shp.Seek(100, os.SEEK_SET)
	shx.Seek(100, os.SEEK_SET)
	w := &Writer{shp: shp, shx: shx, dbf: dbf, GeometryType: t}
	return w, nil
}

func (w *Writer) Write(shape Shape) {
	// increate bbox
	if w.num == 0 {
		w.bbox = shape.BBox()
	} else {
		w.bbox.Extend(shape.BBox())
	}

	w.num += 1
	binary.Write(w.shp, binary.BigEndian, w.num)
	w.shp.Seek(4, os.SEEK_CUR)
	start, _ := w.shp.Seek(0, os.SEEK_CUR)
	binary.Write(w.shp, binary.LittleEndian, w.GeometryType)
	shape.write(w.shp)
	finish, _ := w.shp.Seek(0, os.SEEK_CUR)
	length := int32(math.Floor((float64(finish) - float64(start)) / 2.0))
	w.shp.Seek(start-4, os.SEEK_SET)
	binary.Write(w.shp, binary.BigEndian, length)
	w.shp.Seek(finish, os.SEEK_SET)

	// write shx
	binary.Write(w.shx, binary.BigEndian, int32((start-8)/2))
	binary.Write(w.shx, binary.BigEndian, length)
}

func (w *Writer) Close() {
	w.writeHeader(w.shx)
	w.writeHeader(w.shp)
	w.writeDbfHeader(w.dbf)
	w.shp.Close()
	w.shx.Close()
	w.dbf.Close()
}

func (w *Writer) writeHeader(file *os.File) {
	filelength, _ := file.Seek(0, os.SEEK_END)
	if filelength == 0 {
		filelength = 100
	}
	file.Seek(0, os.SEEK_SET)
	// file code
	binary.Write(file, binary.BigEndian, []int32{9994, 0, 0, 0, 0, 0})
	// file length
	binary.Write(file, binary.BigEndian, int32(filelength/2))
	// version and shape type
	binary.Write(file, binary.LittleEndian, []int32{1000, int32(w.GeometryType)})
	// bounding box
	binary.Write(file, binary.LittleEndian, w.bbox)
	// elevation, measure
	binary.Write(file, binary.LittleEndian, []float64{0.0, 0.0, 0.0, 0.0})
}
