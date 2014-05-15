package shp

import (
	"os"
	"testing"
)

var filename_prefix string = "test_files/write_"

func removeShapefile(filename string) {
	os.Remove(filename + ".shp")
	os.Remove(filename + ".shx")
	os.Remove(filename + ".dbf")
}

func TestWritePoint(t *testing.T) {
	filename := filename_prefix + "point"
	defer removeShapefile(filename)

	points := [][]float64{
		{0.0, 0.0},
		{5.0, 5.0},
		{10.0, 10.0},
	}

	shape, err := Create(filename+".shp", POINT)
	if err != nil {
		t.Fatal(err)
	}

	for _, p := range points {
		shape.Write(&Point{p[0], p[1]})
	}

	test_Point(t, filename+".shp", points, len(points))
	shape.Close()
}
