package shp

import (
	"testing"
)

func pointsEqual(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if v != b[k] {
			return false
		}
	}
	return true
}

func getShapes(filename string, t *testing.T) (shapes []Shape) {
	file, err := Open(filename)
	if err != nil {
		t.Fatal("Failed to open shapefile: " + filename + " (" + err.Error() + ")")
	}
	defer file.Close()

	for file.Next() {
		_, shape := file.Shape()
		shapes = append(shapes, shape)
	}

	return shapes
}

func TestPoint(t *testing.T) {
	shapes := getShapes("test_files/point.shp", t)
	points := [][]float64{
		{10, 10},
		{5, 5},
		{0, 10},
	}
	if len(shapes) != 3 {
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

func TestPolyLine(t *testing.T) {
	shapes := getShapes("test_files/polyline.shp", t)
	points := [][]float64{
		{0, 0},
		{5, 5},
		{10, 10},
		{15, 15},
		{20, 20},
		{25, 25},
	}
	if len(shapes) != 2 {
		t.Error("Number of shapes read was wrong.")
	}
	for n, s := range shapes {
		p, ok := s.(*PolyLine)
		if !ok {
			t.Fatal("Failed to type assert.")
		}
		for k, point := range p.Points {
			if !pointsEqual(points[n*3+k], []float64{point.X, point.Y}) {
				t.Error("Points did not match.")
			}
		}
	}
}

func TestPolygon(t *testing.T) {
	shapes := getShapes("test_files/polygon.shp", t)
	points := [][]float64{
		{0, 0},
		{0, 5},
		{5, 5},
		{5, 0},
		{0, 0},
	}
	if len(shapes) != 1 {
		t.Error("Number of shapes read was wrong.")
	}
	for n, s := range shapes {
		p, ok := s.(*Polygon)
		if !ok {
			t.Fatal("Failed to type assert.")
		}
		for k, point := range p.Points {
			if !pointsEqual(points[n*3+k], []float64{point.X, point.Y}) {
				t.Error("Points did not match.")
			}
		}
	}
}

func TestMultiPoint(t *testing.T) {
	shapes := getShapes("test_files/multipoint.shp", t)
	points := [][]float64{
		{10, 10},
		{5, 5},
		{0, 10},
	}
	if len(shapes) != 1 {
		t.Error("Number of shapes read was wrong.")
	}
	for n, s := range shapes {
		p, ok := s.(*MultiPoint)
		if !ok {
			t.Fatal("Failed to type assert.")
		}
		for k, point := range p.Points {
			if !pointsEqual(points[n*3+k], []float64{point.X, point.Y}) {
				t.Error("Points did not match.")
			}
		}
	}
}

func TestPointZ(t *testing.T) {
	shapes := getShapes("test_files/pointz.shp", t)
	points := [][]float64{
		{10, 10, 100},
		{5, 5, 50},
		{0, 10, 75},
	}
	if len(shapes) != 3 {
		t.Error("Number of shapes read was wrong.")
	}
	for n, s := range shapes {
		p, ok := s.(*PointZ)
		if !ok {
			t.Fatal("Failed to type assert.")
		}
		if !pointsEqual([]float64{p.X, p.Y, p.Z}, points[n]) {
			t.Error("Points did not match.")
		}
	}
}

func TestPolyLineZ(t *testing.T) {
	shapes := getShapes("test_files/polylinez.shp", t)
	points := [][]float64{
		{0, 0, 0},
		{5, 5, 5},
		{10, 10, 10},
		{15, 15, 15},
		{20, 20, 20},
		{25, 25, 25},
	}
	if len(shapes) != 2 {
		t.Error("Number of shapes read was wrong.")
	}
	for n, s := range shapes {
		p, ok := s.(*PolyLineZ)
		if !ok {
			t.Fatal("Failed to type assert.")
		}
		for k, point := range p.Points {
			if !pointsEqual(points[n*3+k], []float64{point.X, point.Y, p.ZArray[k]}) {
				t.Error("Points did not match.")
			}
		}
	}
}

func TestPolygonZ(t *testing.T) {
	shapes := getShapes("test_files/polygonz.shp", t)
	points := [][]float64{
		{0, 0, 0},
		{0, 5, 5},
		{5, 5, 10},
		{5, 0, 15},
		{0, 0, 0},
	}
	if len(shapes) != 1 {
		t.Error("Number of shapes read was wrong.")
	}
	for n, s := range shapes {
		p, ok := s.(*PolygonZ)
		if !ok {
			t.Fatal("Failed to type assert.")
		}
		for k, point := range p.Points {
			if !pointsEqual(points[n*3+k], []float64{point.X, point.Y, p.ZArray[k]}) {
				t.Error("Points did not match.")
			}
		}
	}
}

func TestMultiPointZ(t *testing.T) {
	shapes := getShapes("test_files/multipointz.shp", t)
	points := [][]float64{
		{10, 10, 100},
		{5, 5, 50},
		{0, 10, 75},
	}
	if len(shapes) != 1 {
		t.Error("Number of shapes read was wrong.")
	}
	for n, s := range shapes {
		p, ok := s.(*MultiPointZ)
		if !ok {
			t.Fatal("Failed to type assert.")
		}
		for k, point := range p.Points {
			if !pointsEqual(points[n*3+k], []float64{point.X, point.Y, p.ZArray[k]}) {
				t.Error("Points did not match.")
			}
		}
	}
}
