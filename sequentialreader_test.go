package shp

import (
	"os"
	"testing"
)

func openFile(name string, t *testing.T) *os.File {
	f, err := os.Open(name)
	if err != nil {
		t.Fatalf("Failed to open %s: %v", name, err)
	}
	return f
}

func getShapesSequentially(prefix string, t *testing.T) (shapes []Shape) {
	shp := openFile(prefix+".shp", t)
	dbf := openFile(prefix+".dbf", t)

	sr := SequentialReaderFromExt(shp, dbf)
	for sr.Next() {
		_, shape := sr.Shape()
		shapes = append(shapes, shape)
	}
	if err := sr.Err(); err != nil {
		t.Fatalf("Error when iterating over the shapes: %v", err)
	}

	if err := sr.Close(); err != nil {
		t.Fatalf("Could not close sequential reader: %v", err)
	}
	return shapes
}

func TestSequentialReadPoint(t *testing.T) {
	prefix := "test_files/point"
	points := dataForReadTests[prefix]

	shapes := getShapesSequentially(prefix, t)
	if len(shapes) != len(points) {
		t.Error("Number of shapes read was wrong.")
	}
	for n, s := range shapes {
		p, ok := s.(*Point)
		if !ok {
			t.Fatal("Failed to type assert.")
		}
		if !pointsEqual([]float64{p.X, p.Y}, points[n]) {
			t.Error("Points did not match.")
		}
	}
}
