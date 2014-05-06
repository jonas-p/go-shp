package goshp

import (
	"encoding/binary"
	"os"
)

type ShapeType int32

const (
	NULL        ShapeType = 0
	POINT                 = 1
	POLYLINE              = 3
	POLYGON               = 5
	MULTIPOINT            = 9
	POINTZ                = 11
	POLYLINEZ             = 13
	POLYGONZ              = 15
	MULTIPOINTZ           = 18
	POINTM                = 21
	POLYLINEM             = 23
	POLYGONM              = 25
	MULTIPOINTM           = 28
	MULTIPATCH            = 31
)

type Box struct {
	MinX, MinY, MaxX, MaxY float64
}

func (b *Box) Extend(box Box) {
	if box.MinX < b.MinX {
		b.MinX = box.MinX
	}
	if box.MinY < b.MinY {
		b.MinY = box.MinY
	}
	if box.MaxX > b.MaxX {
		b.MaxX = box.MaxX
	}
	if box.MaxY > b.MaxY {
		b.MaxY = box.MaxY
	}
}

type Shape interface {
	BBox() Box

	read(*os.File)
	write(*os.File)
}

type Null struct {
}

func (n *Null) BBox() Box {
	return Box{0.0, 0.0, 0.0, 0.0}
}

func (n *Null) read(file *os.File) {
	binary.Read(file, binary.LittleEndian, n)
}

func (n *Null) write(file *os.File) {
	binary.Write(file, binary.LittleEndian, n)
}

type Point struct {
	X, Y float64
}

func (p *Point) BBox() Box {
	return Box{p.X, p.Y, p.X, p.Y}
}

func (p *Point) read(file *os.File) {
	binary.Read(file, binary.LittleEndian, p)
}

func (p *Point) write(file *os.File) {
	binary.Write(file, binary.LittleEndian, p)
}

type PolyLine struct {
	Box
	NumParts  int32
	NumPoints int32
	Parts     []int32
	Points    []Point
}

// TODO: Rewrite to calculate actual values
func (p PolyLine) BBox() Box {
	return p.Box
}

func (p *PolyLine) read(file *os.File) {
	binary.Read(file, binary.LittleEndian, &p.Box)
	binary.Read(file, binary.LittleEndian, &p.NumParts)
	binary.Read(file, binary.LittleEndian, &p.NumPoints)
	p.Parts = make([]int32, p.NumParts)
	p.Points = make([]Point, p.NumPoints)
	binary.Read(file, binary.LittleEndian, &p.Parts)
	binary.Read(file, binary.LittleEndian, &p.Points)
}

func (p *PolyLine) write(file *os.File) {
	binary.Write(file, binary.LittleEndian, p.Box)
	binary.Write(file, binary.LittleEndian, p.NumParts)
	binary.Write(file, binary.LittleEndian, p.NumPoints)
	binary.Write(file, binary.LittleEndian, p.Parts)
	binary.Write(file, binary.LittleEndian, p.Points)
}

// The Polygon structure is identical to the PolyLine structure
type Polygon PolyLine

func (p Polygon) BBox() Box {
	return p.Box
}

func (p *Polygon) read(file *os.File) {
	binary.Read(file, binary.LittleEndian, &p.Box)
	binary.Read(file, binary.LittleEndian, &p.NumParts)
	binary.Read(file, binary.LittleEndian, &p.NumPoints)
	p.Parts = make([]int32, p.NumParts)
	p.Points = make([]Point, p.NumPoints)
	binary.Read(file, binary.LittleEndian, &p.Parts)
	binary.Read(file, binary.LittleEndian, &p.Points)
}

func (p *Polygon) write(file *os.File) {
	binary.Write(file, binary.LittleEndian, p.Box)
	binary.Write(file, binary.LittleEndian, p.NumParts)
	binary.Write(file, binary.LittleEndian, p.NumPoints)
	binary.Write(file, binary.LittleEndian, p.Parts)
	binary.Write(file, binary.LittleEndian, p.Points)
}
