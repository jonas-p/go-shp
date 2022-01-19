package shp

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"testing"
)

var filenamePrefix = "test_files/write_"

func removeShapefile(filename string) {
	os.Remove(filename + ".shp")
	os.Remove(filename + ".shx")
	os.Remove(filename + ".dbf")
}

func pointsToFloats(points []Point) [][]float64 {
	floats := make([][]float64, len(points))
	for k, v := range points {
		floats[k] = make([]float64, 2)
		floats[k][0] = v.X
		floats[k][1] = v.Y
	}
	return floats
}

func TestAppend(t *testing.T) {
	filename := filenamePrefix + "point"
	defer removeShapefile(filename)
	points := [][]float64{
		{0.0, 0.0},
		{5.0, 5.0},
		{10.0, 10.0},
	}

	shape, err := Create(filename+".shp", POINT, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range points {
		shape.Write(&Point{p[0], p[1]})
	}
	wantNum := shape.num
	shape.Close()

	newPoints := [][]float64{
		{15.0, 15.0},
		{20.0, 20.0},
		{25.0, 25.0},
	}
	shape, err = Append(filename + ".shp")
	if err != nil {
		t.Fatal(err)
	}
	if shape.GeometryType != POINT {
		t.Fatalf("wanted geo type %d, got %d", POINT, shape.GeometryType)
	}
	if shape.num != wantNum {
		t.Fatalf("wrong 'num', wanted type %d, got %d", wantNum, shape.num)
	}

	for _, p := range newPoints {
		shape.Write(&Point{p[0], p[1]})
	}

	points = append(points, newPoints...)

	shapes := getShapesFromFile(filename, t)
	if len(shapes) != len(points) {
		t.Error("Number of shapes read was wrong")
	}
	testPoint(t, points, shapes)
}

func TestWritePoint(t *testing.T) {
	filename := filenamePrefix + "point"
	defer removeShapefile(filename)

	points := [][]float64{
		{0.0, 0.0},
		{5.0, 5.0},
		{10.0, 10.0},
	}

	shape, err := Create(filename+".shp", POINT, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range points {
		shape.Write(&Point{p[0], p[1]})
	}
	shape.Close()

	shapes := getShapesFromFile(filename, t)
	if len(shapes) != len(points) {
		t.Error("Number of shapes read was wrong")
	}
	testPoint(t, points, shapes)
}

func TestWritePolyLine(t *testing.T) {
	filename := filenamePrefix + "polyline"
	defer removeShapefile(filename)

	points := [][]Point{
		{Point{0.0, 0.0}, Point{5.0, 5.0}},
		{Point{10.0, 10.0}, Point{15.0, 15.0}},
	}

	shape, err := Create(filename+".shp", POLYLINE, nil)
	if err != nil {
		t.Log(shape, err)
	}

	l := NewPolyLine(points)

	lWant := &PolyLine{
		Box:       Box{MinX: 0, MinY: 0, MaxX: 15, MaxY: 15},
		NumParts:  2,
		NumPoints: 4,
		Parts:     []int32{0, 2},
		Points: []Point{{X: 0, Y: 0},
			{X: 5, Y: 5},
			{X: 10, Y: 10},
			{X: 15, Y: 15},
		},
	}
	if !reflect.DeepEqual(l, lWant) {
		t.Errorf("incorrect NewLine: have: %+v; want: %+v", l, lWant)
	}

	shape.Write(l)
	shape.Close()

	shapes := getShapesFromFile(filename, t)
	if len(shapes) != 1 {
		t.Error("Number of shapes read was wrong")
	}
	testPolyLine(t, pointsToFloats(flatten(points)), shapes)
}

type seekTracker struct {
	io.Writer
	offset int64
}

func (s *seekTracker) Seek(offset int64, whence int) (int64, error) {
	s.offset = offset
	return s.offset, nil
}

func (s *seekTracker) Close() error {
	return nil
}

func TestWriteAttribute(t *testing.T) {
	buf := new(bytes.Buffer)
	s := &seekTracker{Writer: buf}
	w := Writer{
		dbf: s,
		dbfFields: []Field{
			StringField("A_STRING", 6),
			FloatField("A_FLOAT", 8, 4),
			NumberField("AN_INT", 4),
		},
		dbfRecordLength: 100,
	}

	tests := []struct {
		name       string
		row        int
		field      int
		data       interface{}
		wantOffset int64
		wantData   string
	}{
		{"string-0", 0, 0, "test", 1, "test"},
		{"string-0-overflow-1", 0, 0, "overflo", 0, ""},
		{"string-0-overflow-n", 0, 0, "overflowing", 0, ""},
		{"string-3", 3, 0, "things", 301, "things"},
		{"float-0", 0, 1, 123.44, 7, "123.4400"},
		{"float-0-overflow-1", 0, 1, 1234.0, 0, ""},
		{"float-0-overflow-n", 0, 1, 123456789.0, 0, ""},
		{"int-0", 0, 2, 4242, 15, "4242"},
		{"int-0-overflow-1", 0, 2, 42424, 0, ""},
		{"int-0-overflow-n", 0, 2, 42424343, 0, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf.Reset()
			s.offset = 0

			err := w.WriteAttribute(test.row, test.field, test.data)

			if buf.String() != test.wantData {
				t.Errorf("got data: %v, want: %v", buf.String(), test.wantData)
			}
			if s.offset != test.wantOffset {
				t.Errorf("got seek offset: %v, want: %v", s.offset, test.wantOffset)
			}
			if err == nil && test.wantData == "" {
				t.Error("got no data and no error")
			}
		})
	}
}
