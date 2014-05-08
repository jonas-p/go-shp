package goshp

import (
	"testing"
)

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
		shape := points.Shape()
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
		shape := lines.Shape()
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
		shape := polygons.Shape()
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
