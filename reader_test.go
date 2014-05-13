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

func TestReadPoints(t *testing.T) {
	points, err := Open("test_files/points.shp")
	if err != nil {
		t.Fatal("Failed to open points.shp: " + err.Error())
	}
	defer points.Close()

	if points.GeometryType != POINT {
		t.Error("File was not Point type")
	}

	n := 0
	for points.Next() {
		n += 1
		_, shape := points.Shape()
		if err != nil {
			t.Error("Failed to read shape")
		}

		p, ok := shape.(*Point)
		if ok != true {
			t.Error("Failed to type assert shape to Point")
		}
		if p.X == 0.0 || p.Y == 0.0 {
			t.Error("Point had 0.0 X or Y value")
		}
	}

	if n != 3 {
		t.Error("Number of points read was not three")
	}
}

func TestReadPolyLines(t *testing.T) {
	numParts := []int32{1, 1, 1}
	numPoints := []int32{2, 4, 5}
	lines, err := Open("test_files/lines.shp")
	if err != nil {
		t.Error("Failed to open lines.shp: " + err.Error())
	}
	defer lines.Close()

	if lines.GeometryType != POLYLINE {
		t.Fatal("File was not Polygon type")
	}

	n := 0
	for lines.Next() {
		n += 1
		_, shape := lines.Shape()
		if err != nil {
			t.Error("Failed to read shape")
		}

		line, ok := shape.(*PolyLine)
		if ok != true {
			t.Error("Failed to type assert shape to PolyLine")
		}

		if numParts[n-1] != line.NumParts {
			t.Error("numParts mismatch")
		}
		if numPoints[n-1] != line.NumPoints {
			t.Error("numPoints mismatch")
		}

		if line.NumParts != int32(len(line.Parts)) {
			t.Error("numParts was not the same as the length of Parts")
		}
		if line.NumPoints != int32(len(line.Points)) {
			t.Error("numPoints was not the same as the length of Points")
		}
	}

	if n != len(numParts) {
		t.Error("Number of polylines read was not ", len(numParts))
	}
}

func TestReadPolygons(t *testing.T) {
	numParts := []int32{1, 1}
	numPoints := []int32{5, 8}
	polygons, err := Open("test_files/polygons.shp")
	if err != nil {
		t.Fatal("Failed to open polygons.shp: " + err.Error())
	}
	defer polygons.Close()

	if polygons.GeometryType != POLYGON {
		t.Error("File was not Polygon type")
	}

	n := 0
	for polygons.Next() {
		n += 1
		_, shape := polygons.Shape()
		if err != nil {
			t.Error("Failed to read shape")
		}

		polygon, ok := shape.(*Polygon)
		if ok != true {
			t.Error("Failed to type assert shape to Polygon")
		}

		if numParts[n-1] != polygon.NumParts {
			t.Error("numParts mismatch")
		}
		if numPoints[n-1] != polygon.NumPoints {
			t.Error("numPoints mismatch")
		}

		if polygon.NumParts != int32(len(polygon.Parts)) {
			t.Error("numParts was not the same as the length of Parts")
		}
		if polygon.NumPoints != int32(len(polygon.Points)) {
			t.Error("numPoints was not the same as the length of Points")
		}
	}

	if n != len(numParts) {
		t.Error("Number of polygons read was not ", len(numParts))
	}
}

func TestReadMultiPoints(t *testing.T) {
	numPoints := []int32{5, 3}
	multipoints, err := Open("test_files/multipoints.shp")
	if err != nil {
		t.Fatal("Failed to open multipoints.shp: " + err.Error())
	}
	defer multipoints.Close()

	if multipoints.GeometryType != MULTIPOINT {
		t.Error("File was not Polygon type")
	}

	n := 0
	var shape Shape
	for multipoints.Next() {
		n, shape = multipoints.Shape()
		mp, ok := shape.(*MultiPoint)
		if ok != true {
			t.Error("Failed to type assert shape to MultiPoint")
		}

		if mp.NumPoints != numPoints[n] {
			t.Error("NumPoints mismatch")
		}
	}

	if n != 1 {
		t.Error("Number of MultiPoints read was not 2")
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
