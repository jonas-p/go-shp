package shp

import (
	"encoding/binary"
	"io"
)

type ShapeType int32

const (
	NULL        ShapeType = 0
	POINT                 = 1
	POLYLINE              = 3
	POLYGON               = 5
	MULTIPOINT            = 8
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

// Box structure made up from four coordinates. This type
// is used to represent bounding boxes
type Box struct {
	MinX, MinY, MaxX, MaxY float64
}

// Extend extends the box with coordinates from the provided
// box. This method calls Box.ExtendWithPoint twice with
// {MinX, MinY} and {MaxX, MaxY}
func (b *Box) Extend(box Box) {
	b.ExtendWithPoint(Point{box.MinX, box.MinY})
	b.ExtendWithPoint(Point{box.MaxX, box.MaxX})
}

// ExtendWithPoint extends box with coordinates from point
// if they are outside the range of the current box.
func (b *Box) ExtendWithPoint(p Point) {
	if p.X < b.MinX {
		b.MinX = p.X
	}
	if p.Y < b.MinY {
		b.MinY = p.Y
	}
	if p.X > b.MaxX {
		b.MaxX = p.X
	}
	if p.Y > b.MaxY {
		b.MaxY = p.Y
	}
}

// BBoxFromPoints returns the bounding box calculated
// from points.
func BBoxFromPoints(points []Point) (box Box) {
	for k, p := range points {
		if k == 0 {
			box = Box{p.X, p.Y, p.X, p.Y}
		} else {
			if p.X < box.MinX {
				box.MinX = p.X
			}
			if p.Y < box.MinY {
				box.MinY = p.Y
			}
			if p.X > box.MaxX {
				box.MaxX = p.X
			}
			if p.Y > box.MaxY {
				box.MaxY = p.Y
			}
		}
	}
	return
}

// Shape interface
type Shape interface {
	BBox() Box

	read(io.Reader)
	write(io.Writer)
}

// Shapefile NULL type
type Null struct {
}

// Returns the bounding box of the Null feature
func (n Null) BBox() Box {
	return Box{0.0, 0.0, 0.0, 0.0}
}

func (n *Null) read(file io.Reader) {
	binary.Read(file, binary.LittleEndian, n)
}

func (n *Null) write(file io.Writer) {
	binary.Write(file, binary.LittleEndian, n)
}

// Shapefile Point type
type Point struct {
	X, Y float64
}

// Returns the bounding box of the Point feature
func (p Point) BBox() Box {
	return Box{p.X, p.Y, p.X, p.Y}
}

func (p *Point) read(file io.Reader) {
	binary.Read(file, binary.LittleEndian, p)
}

func (p *Point) write(file io.Writer) {
	binary.Write(file, binary.LittleEndian, p)
}

func flatten(points [][]Point) []Point {
	n, i := 0, 0
	for _, v := range points {
		n += len(v)
	}
	r := make([]Point, n)
	for _, v := range points {
		for _, p := range v {
			r[i] = p
			i += 1
		}
	}
	return r
}

// Shapefile PolyLine type
type PolyLine struct {
	Box
	NumParts  int32
	NumPoints int32
	Parts     []int32
	Points    []Point
}

// NewPolyLine returns a pointer a new PolyLine created
// with the provided points. The inner slice should be
// the points that the parent part consists of.
func NewPolyLine(parts [][]Point) *PolyLine {
	points := flatten(parts)

	p := &PolyLine{}
	p.NumParts = int32(len(parts))
	p.NumPoints = int32(len(points))
	p.Parts = make([]int32, len(parts))
	p.Points = points
	p.Box = p.BBox()

	return p
}

// Returns the bounding box of the PolyLine feature
func (p PolyLine) BBox() Box {
	return BBoxFromPoints(p.Points)
}

func (p *PolyLine) read(file io.Reader) {
	binary.Read(file, binary.LittleEndian, &p.Box)
	binary.Read(file, binary.LittleEndian, &p.NumParts)
	binary.Read(file, binary.LittleEndian, &p.NumPoints)
	p.Parts = make([]int32, p.NumParts)
	p.Points = make([]Point, p.NumPoints)
	binary.Read(file, binary.LittleEndian, &p.Parts)
	binary.Read(file, binary.LittleEndian, &p.Points)
}

func (p *PolyLine) write(file io.Writer) {
	binary.Write(file, binary.LittleEndian, p.Box)
	binary.Write(file, binary.LittleEndian, p.NumParts)
	binary.Write(file, binary.LittleEndian, p.NumPoints)
	binary.Write(file, binary.LittleEndian, p.Parts)
	binary.Write(file, binary.LittleEndian, p.Points)
}

// Shapefile Polygon type
// The Polygon structure is identical to the PolyLine structure
type Polygon PolyLine

// Returns the bounding box of the Polygon feature
func (p Polygon) BBox() Box {
	return BBoxFromPoints(p.Points)
}

func (p *Polygon) read(file io.Reader) {
	binary.Read(file, binary.LittleEndian, &p.Box)
	binary.Read(file, binary.LittleEndian, &p.NumParts)
	binary.Read(file, binary.LittleEndian, &p.NumPoints)
	p.Parts = make([]int32, p.NumParts)
	p.Points = make([]Point, p.NumPoints)
	binary.Read(file, binary.LittleEndian, &p.Parts)
	binary.Read(file, binary.LittleEndian, &p.Points)
}

func (p *Polygon) write(file io.Writer) {
	binary.Write(file, binary.LittleEndian, p.Box)
	binary.Write(file, binary.LittleEndian, p.NumParts)
	binary.Write(file, binary.LittleEndian, p.NumPoints)
	binary.Write(file, binary.LittleEndian, p.Parts)
	binary.Write(file, binary.LittleEndian, p.Points)
}

// Shapefile MultiPoint type
type MultiPoint struct {
	Box       Box
	NumPoints int32
	Points    []Point
}

// Returns the bounding box of the MultiPoint feature
func (p MultiPoint) BBox() Box {
	return BBoxFromPoints(p.Points)
}

func (p *MultiPoint) read(file io.Reader) {
	binary.Read(file, binary.LittleEndian, &p.Box)
	binary.Read(file, binary.LittleEndian, &p.NumPoints)
	p.Points = make([]Point, p.NumPoints)
	binary.Read(file, binary.LittleEndian, &p.Points)
}

func (p *MultiPoint) write(file io.Writer) {
	binary.Write(file, binary.LittleEndian, p.Box)
	binary.Write(file, binary.LittleEndian, p.NumPoints)
	binary.Write(file, binary.LittleEndian, p.Points)
}

// Shapefile PointZ type
type PointZ struct {
	X float64
	Y float64
	Z float64
	M float64
}

// Returns the bounding box of the PointZ feature
func (p PointZ) BBox() Box {
	return Box{p.X, p.Y, p.X, p.Y}
}

func (p *PointZ) read(file io.Reader) {
	binary.Read(file, binary.LittleEndian, p)
}

func (p *PointZ) write(file io.Writer) {
	binary.Write(file, binary.LittleEndian, p)
}

// Shapefile PolyLineZ type
type PolyLineZ struct {
	Box       Box
	NumParts  int32
	NumPoints int32
	Parts     []int32
	Points    []Point
	ZRange    [2]float64
	ZArray    []float64
	MRange    [2]float64
	MArray    []float64
}

// Returns the bounding box of the PolyLineZ feature
func (p PolyLineZ) BBox() Box {
	return BBoxFromPoints(p.Points)
}

func (p *PolyLineZ) read(file io.Reader) {
	binary.Read(file, binary.LittleEndian, &p.Box)
	binary.Read(file, binary.LittleEndian, &p.NumParts)
	binary.Read(file, binary.LittleEndian, &p.NumPoints)
	p.Parts = make([]int32, p.NumParts)
	p.Points = make([]Point, p.NumPoints)
	p.ZArray = make([]float64, p.NumPoints)
	p.MArray = make([]float64, p.NumPoints)
	binary.Read(file, binary.LittleEndian, &p.Parts)
	binary.Read(file, binary.LittleEndian, &p.Points)
	binary.Read(file, binary.LittleEndian, &p.ZRange)
	binary.Read(file, binary.LittleEndian, &p.ZArray)
	binary.Read(file, binary.LittleEndian, &p.MRange)
	binary.Read(file, binary.LittleEndian, &p.MArray)
}

func (p *PolyLineZ) write(file io.Writer) {
	binary.Write(file, binary.LittleEndian, p.Box)
	binary.Write(file, binary.LittleEndian, p.NumParts)
	binary.Write(file, binary.LittleEndian, p.NumPoints)
	binary.Write(file, binary.LittleEndian, p.Parts)
	binary.Write(file, binary.LittleEndian, p.Points)
	binary.Write(file, binary.LittleEndian, p.ZRange)
	binary.Write(file, binary.LittleEndian, p.ZArray)
	binary.Write(file, binary.LittleEndian, p.MRange)
	binary.Write(file, binary.LittleEndian, p.MArray)
}

// Shapefile PolygonZ type
// The PolygonZ structure is identical to the PolyLineZ structure
type PolygonZ PolyLineZ

// Returns the bounding box of the PolygonZ feature
func (p PolygonZ) BBox() Box {
	return BBoxFromPoints(p.Points)
}

func (p *PolygonZ) read(file io.Reader) {
	binary.Read(file, binary.LittleEndian, &p.Box)
	binary.Read(file, binary.LittleEndian, &p.NumParts)
	binary.Read(file, binary.LittleEndian, &p.NumPoints)
	p.Parts = make([]int32, p.NumParts)
	p.Points = make([]Point, p.NumPoints)
	p.ZArray = make([]float64, p.NumPoints)
	p.MArray = make([]float64, p.NumPoints)
	binary.Read(file, binary.LittleEndian, &p.Parts)
	binary.Read(file, binary.LittleEndian, &p.Points)
	binary.Read(file, binary.LittleEndian, &p.ZRange)
	binary.Read(file, binary.LittleEndian, &p.ZArray)
	binary.Read(file, binary.LittleEndian, &p.MRange)
	binary.Read(file, binary.LittleEndian, &p.MArray)
}

func (p *PolygonZ) write(file io.Writer) {
	binary.Write(file, binary.LittleEndian, p.Box)
	binary.Write(file, binary.LittleEndian, p.NumParts)
	binary.Write(file, binary.LittleEndian, p.NumPoints)
	binary.Write(file, binary.LittleEndian, p.Parts)
	binary.Write(file, binary.LittleEndian, p.Points)
	binary.Write(file, binary.LittleEndian, p.ZRange)
	binary.Write(file, binary.LittleEndian, p.ZArray)
	binary.Write(file, binary.LittleEndian, p.MRange)
	binary.Write(file, binary.LittleEndian, p.MArray)
}

// Shapefile MultiPointZ type
type MultiPointZ struct {
	Box       Box
	NumPoints int32
	Points    []Point
	ZRange    [2]float64
	ZArray    []float64
	MRange    [2]float64
	MArray    []float64
}

// Returns the bounding box of the MultiPointZ feature
func (p MultiPointZ) BBox() Box {
	return BBoxFromPoints(p.Points)
}

func (p *MultiPointZ) read(file io.Reader) {
	binary.Read(file, binary.LittleEndian, &p.Box)
	binary.Read(file, binary.LittleEndian, &p.NumPoints)
	p.Points = make([]Point, p.NumPoints)
	p.ZArray = make([]float64, p.NumPoints)
	p.MArray = make([]float64, p.NumPoints)
	binary.Read(file, binary.LittleEndian, &p.Points)
	binary.Read(file, binary.LittleEndian, &p.ZRange)
	binary.Read(file, binary.LittleEndian, &p.ZArray)
	binary.Read(file, binary.LittleEndian, &p.MRange)
	binary.Read(file, binary.LittleEndian, &p.MArray)
}

func (p *MultiPointZ) write(file io.Writer) {
	binary.Write(file, binary.LittleEndian, p.Box)
	binary.Write(file, binary.LittleEndian, p.NumPoints)
	binary.Write(file, binary.LittleEndian, p.Points)
	binary.Write(file, binary.LittleEndian, p.ZRange)
	binary.Write(file, binary.LittleEndian, p.ZArray)
	binary.Write(file, binary.LittleEndian, p.MRange)
	binary.Write(file, binary.LittleEndian, p.MArray)
}

// Shapefile PointM type
type PointM struct {
	X float64
	Y float64
	M float64
}

// Returns the bounding box of the PointM feature
func (p PointM) BBox() Box {
	return Box{p.X, p.Y, p.X, p.Y}
}

func (p *PointM) read(file io.Reader) {
	binary.Read(file, binary.LittleEndian, p)
}

func (p *PointM) write(file io.Writer) {
	binary.Write(file, binary.LittleEndian, p)
}

// Shapefile PolyLineM type
type PolyLineM struct {
	Box       Box
	NumParts  int32
	NumPoints int32
	Parts     []int32
	Points    []Point
	MRange    [2]float64
	MArray    []float64
}

// Returns the bounding box of the PolyLineM feature
func (p PolyLineM) BBox() Box {
	return BBoxFromPoints(p.Points)
}

func (p *PolyLineM) read(file io.Reader) {
	binary.Read(file, binary.LittleEndian, &p.Box)
	binary.Read(file, binary.LittleEndian, &p.NumParts)
	binary.Read(file, binary.LittleEndian, &p.NumPoints)
	p.Parts = make([]int32, p.NumParts)
	p.Points = make([]Point, p.NumPoints)
	p.MArray = make([]float64, p.NumPoints)
	binary.Read(file, binary.LittleEndian, &p.Parts)
	binary.Read(file, binary.LittleEndian, &p.Points)
	binary.Read(file, binary.LittleEndian, &p.MRange)
	binary.Read(file, binary.LittleEndian, &p.MArray)
}

func (p *PolyLineM) write(file io.Writer) {
	binary.Write(file, binary.LittleEndian, p.Box)
	binary.Write(file, binary.LittleEndian, p.NumParts)
	binary.Write(file, binary.LittleEndian, p.NumPoints)
	binary.Write(file, binary.LittleEndian, p.Parts)
	binary.Write(file, binary.LittleEndian, p.Points)
	binary.Write(file, binary.LittleEndian, p.MRange)
	binary.Write(file, binary.LittleEndian, p.MArray)
}

// Shapefile PolygonM type
// The PolygonZ structure is identical to the PolyLineZ structure
type PolygonM PolyLineZ

// Returns the bounding box of the PolygonM feature
func (p PolygonM) BBox() Box {
	return BBoxFromPoints(p.Points)
}

func (p *PolygonM) read(file io.Reader) {
	binary.Read(file, binary.LittleEndian, &p.Box)
	binary.Read(file, binary.LittleEndian, &p.NumParts)
	binary.Read(file, binary.LittleEndian, &p.NumPoints)
	p.Parts = make([]int32, p.NumParts)
	p.Points = make([]Point, p.NumPoints)
	p.MArray = make([]float64, p.NumPoints)
	binary.Read(file, binary.LittleEndian, &p.Parts)
	binary.Read(file, binary.LittleEndian, &p.Points)
	binary.Read(file, binary.LittleEndian, &p.MRange)
	binary.Read(file, binary.LittleEndian, &p.MArray)
}

func (p *PolygonM) write(file io.Writer) {
	binary.Write(file, binary.LittleEndian, p.Box)
	binary.Write(file, binary.LittleEndian, p.NumParts)
	binary.Write(file, binary.LittleEndian, p.NumPoints)
	binary.Write(file, binary.LittleEndian, p.Parts)
	binary.Write(file, binary.LittleEndian, p.Points)
	binary.Write(file, binary.LittleEndian, p.MRange)
	binary.Write(file, binary.LittleEndian, p.MArray)
}

// Shapefile MultiPointM type
type MultiPointM struct {
	Box       Box
	NumPoints int32
	Points    []Point
	MRange    [2]float64
	MArray    []float64
}

// Returns the bounding box of the MultiPointM feature
func (p MultiPointM) BBox() Box {
	return BBoxFromPoints(p.Points)
}

func (p *MultiPointM) read(file io.Reader) {
	binary.Read(file, binary.LittleEndian, &p.Box)
	binary.Read(file, binary.LittleEndian, &p.NumPoints)
	p.Points = make([]Point, p.NumPoints)
	p.MArray = make([]float64, p.NumPoints)
	binary.Read(file, binary.LittleEndian, &p.Points)
	binary.Read(file, binary.LittleEndian, &p.MRange)
	binary.Read(file, binary.LittleEndian, &p.MArray)
}

func (p *MultiPointM) write(file io.Writer) {
	binary.Write(file, binary.LittleEndian, p.Box)
	binary.Write(file, binary.LittleEndian, p.NumPoints)
	binary.Write(file, binary.LittleEndian, p.Points)
	binary.Write(file, binary.LittleEndian, p.MRange)
	binary.Write(file, binary.LittleEndian, p.MArray)
}

// Shapefile MultiPatch type
type MultiPatch struct {
	Box       Box
	NumParts  int32
	NumPoints int32
	Parts     []int32
	PartTypes []int32
	Points    []Point
	ZRange    [2]float64
	ZArray    []float64
	MRange    [2]float64
	MArray    []float64
}

// Returns the bounding box of the MultiPatch feature
func (p MultiPatch) BBox() Box {
	return BBoxFromPoints(p.Points)
}

func (p *MultiPatch) read(file io.Reader) {
	binary.Read(file, binary.LittleEndian, &p.Box)
	binary.Read(file, binary.LittleEndian, &p.NumParts)
	binary.Read(file, binary.LittleEndian, &p.NumPoints)
	p.Parts = make([]int32, p.NumParts)
	p.PartTypes = make([]int32, p.NumParts)
	p.Points = make([]Point, p.NumPoints)
	p.ZArray = make([]float64, p.NumPoints)
	p.MArray = make([]float64, p.NumPoints)
	binary.Read(file, binary.LittleEndian, &p.Parts)
	binary.Read(file, binary.LittleEndian, &p.PartTypes)
	binary.Read(file, binary.LittleEndian, &p.Points)
	binary.Read(file, binary.LittleEndian, &p.ZRange)
	binary.Read(file, binary.LittleEndian, &p.ZArray)
	binary.Read(file, binary.LittleEndian, &p.MRange)
	binary.Read(file, binary.LittleEndian, &p.MArray)
}

func (p *MultiPatch) write(file io.Writer) {
	binary.Write(file, binary.LittleEndian, p.Box)
	binary.Write(file, binary.LittleEndian, p.NumParts)
	binary.Write(file, binary.LittleEndian, p.NumPoints)
	binary.Write(file, binary.LittleEndian, p.Parts)
	binary.Write(file, binary.LittleEndian, p.PartTypes)
	binary.Write(file, binary.LittleEndian, p.Points)
	binary.Write(file, binary.LittleEndian, p.ZRange)
	binary.Write(file, binary.LittleEndian, p.ZArray)
	binary.Write(file, binary.LittleEndian, p.MRange)
	binary.Write(file, binary.LittleEndian, p.MArray)
}

// Field representation of a field object in the DBF file
type Field struct {
	Name      [11]byte
	Fieldtype byte
	Addr      [4]byte // not used
	Size      uint8
	Precision uint8
	Padding   [14]byte
}

// Returns a string representation of the Field. Currently
// this only returns field name.
func (f Field) String() string {
	return string(f.Name[:])
}

// Returns a StringField that can be used in SetFields to
// initialize the DBF file.
func StringField(name string, length uint8) Field {
	// TODO: Error checking
	field := Field{Fieldtype: 'C', Size: length}
	copy(field.Name[:], []byte(name))
	return field
}

// Returns a NumberField that can be used in SetFields to
// initialize the DBF file.
func NumberField(name string, length uint8) Field {
	field := Field{Fieldtype: 'N', Size: length}
	copy(field.Name[:], []byte(name))
	return field
}

// Returns a FloatField that can be used in SetFields to
// initialize the DBF file. Used to store floating points
// with precision in the DBF.
func FloatField(name string, length uint8, precision uint8) Field {
	field := Field{Fieldtype: 'F', Size: length, Precision: precision}
	copy(field.Name[:], []byte(name))
	return field
}

// Returns a DateField that can be used in SetFields to
// initialize the DBF file. Used to store Date strings
// formatted as YYYYMMDD. Data wise this is the same as
// a StringField with length 8.
func DateField(name string) Field {
	field := Field{Fieldtype: 'D', Size: 8}
	copy(field.Name[:], []byte(name))
	return field
}
