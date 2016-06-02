package shp

import "testing"

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

func test_Point(t *testing.T, filename string, points [][]float64, shapes_num int) {
	shapes := getShapes(filename, t)
	if len(shapes) != shapes_num {
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

func test_PolyLine(t *testing.T, filename string, points [][]float64, shapes_num int) {
	shapes := getShapes(filename, t)
	if len(shapes) != shapes_num {
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

func test_Polygon(t *testing.T, filename string, points [][]float64, shapes_num int) {
	shapes := getShapes(filename, t)
	if len(shapes) != shapes_num {
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

func test_MultiPoint(t *testing.T, filename string, points [][]float64, shapes_num int) {
	shapes := getShapes(filename, t)
	if len(shapes) != shapes_num {
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

func test_PointZ(t *testing.T, filename string, points [][]float64, shapes_num int) {
	shapes := getShapes(filename, t)
	if len(shapes) != shapes_num {
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

func test_PolyLineZ(t *testing.T, filename string, points [][]float64, shapes_num int) {
	shapes := getShapes(filename, t)
	if len(shapes) != shapes_num {
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

func test_PolygonZ(t *testing.T, filename string, points [][]float64, shapes_num int) {
	shapes := getShapes(filename, t)
	if len(shapes) != shapes_num {
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

func test_MultiPointZ(t *testing.T, filename string, points [][]float64, shapes_num int) {
	shapes := getShapes(filename, t)
	if len(shapes) != shapes_num {
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

func test_PointM(t *testing.T, filename string, points [][]float64, shapes_num int) {
	shapes := getShapes(filename, t)
	if len(shapes) != shapes_num {
		t.Error("Number of shapes read was wrong.")
	}
	for n, s := range shapes {
		p, ok := s.(*PointM)
		if !ok {
			t.Fatal("Failed to type assert.")
		}
		if !pointsEqual([]float64{p.X, p.Y, p.M}, points[n]) {
			t.Error("Points did not match.")
		}
	}
}

func test_PolyLineM(t *testing.T, filename string, points [][]float64, shapes_num int) {
	shapes := getShapes(filename, t)
	if len(shapes) != shapes_num {
		t.Error("Number of shapes read was wrong.")
	}
	for n, s := range shapes {
		p, ok := s.(*PolyLineM)
		if !ok {
			t.Fatal("Failed to type assert.")
		}
		for k, point := range p.Points {
			if !pointsEqual(points[n*3+k], []float64{point.X, point.Y, p.MArray[k]}) {
				t.Error("Points did not match.")
			}
		}
	}
}

func test_PolygonM(t *testing.T, filename string, points [][]float64, shapes_num int) {
	shapes := getShapes(filename, t)
	if len(shapes) != shapes_num {
		t.Error("Number of shapes read was wrong.")
	}
	for n, s := range shapes {
		p, ok := s.(*PolygonM)
		if !ok {
			t.Fatal("Failed to type assert.")
		}
		for k, point := range p.Points {
			if !pointsEqual(points[n*3+k], []float64{point.X, point.Y, p.MArray[k]}) {
				t.Error("Points did not match.")
			}
		}
	}
}

func test_MultiPointM(t *testing.T, filename string, points [][]float64, shapes_num int) {
	shapes := getShapes(filename, t)
	if len(shapes) != shapes_num {
		t.Error("Number of shapes read was wrong.")
	}
	for n, s := range shapes {
		p, ok := s.(*MultiPointM)
		if !ok {
			t.Fatal("Failed to type assert.")
		}
		for k, point := range p.Points {
			if !pointsEqual(points[n*3+k], []float64{point.X, point.Y, p.MArray[k]}) {
				t.Error("Points did not match.")
			}
		}
	}
}

func test_MultiPatch(t *testing.T, filename string, points [][]float64, shapes_num int) {
	shapes := getShapes(filename, t)
	if len(shapes) != shapes_num {
		t.Error("Number of shapes read was wrong.")
	}
	for n, s := range shapes {
		p, ok := s.(*MultiPatch)
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

func TestReadBBox(t *testing.T) {
	tests := []struct {
		filename string
		want     Box
	}{
		{"test_files/multipatch.shp", Box{0, 0, 10, 10}},
		{"test_files/multipoint.shp", Box{0, 5, 10, 10}},
		{"test_files/multipointm.shp", Box{0, 5, 10, 10}},
		{"test_files/multipointz.shp", Box{0, 5, 10, 10}},
		{"test_files/point.shp", Box{0, 5, 10, 10}},
		{"test_files/pointm.shp", Box{0, 5, 10, 10}},
		{"test_files/pointz.shp", Box{0, 5, 10, 10}},
		{"test_files/polygon.shp", Box{0, 0, 5, 5}},
		{"test_files/polygonm.shp", Box{0, 0, 5, 5}},
		{"test_files/polygonz.shp", Box{0, 0, 5, 5}},
		{"test_files/polyline.shp", Box{0, 0, 25, 25}},
		{"test_files/polylinem.shp", Box{0, 0, 25, 25}},
		{"test_files/polylinez.shp", Box{0, 0, 25, 25}},
	}
	for _, tt := range tests {
		r, err := Open(tt.filename)
		if err != nil {
			t.Fatalf("%v", err)
		}
		if got := r.BBox().MinX; got != tt.want.MinX {
			t.Errorf("got MinX = %v, want %v", got, tt.want.MinX)
		}
		if got := r.BBox().MinY; got != tt.want.MinY {
			t.Errorf("got MinY = %v, want %v", got, tt.want.MinY)
		}
		if got := r.BBox().MaxX; got != tt.want.MaxX {
			t.Errorf("got MaxX = %v, want %v", got, tt.want.MaxX)
		}
		if got := r.BBox().MaxY; got != tt.want.MaxY {
			t.Errorf("got MaxY = %v, want %v", got, tt.want.MaxY)
		}
	}
}

var dataForReadTests = map[string][][]float64{
	"test_files/polygonm": [][]float64{
		{0, 0, 0},
		{0, 5, 5},
		{5, 5, 10},
		{5, 0, 15},
		{0, 0, 0},
	},
	"test_files/multipointm": [][]float64{
		{10, 10, 100},
		{5, 5, 50},
		{0, 10, 75},
	},
	"test_files/multipatch": [][]float64{
		{0, 0, 0},
		{10, 0, 0},
		{10, 10, 0},
		{0, 10, 0},
		{0, 0, 0},
		{0, 10, 0},
		{0, 10, 10},
		{0, 0, 10},
		{0, 0, 0},
		{0, 10, 0},
		{10, 0, 0},
		{10, 0, 10},
		{10, 10, 10},
		{10, 10, 0},
		{10, 0, 0},
		{0, 0, 0},
		{0, 0, 10},
		{10, 0, 10},
		{10, 0, 0},
		{0, 0, 0},
		{10, 10, 0},
		{10, 10, 10},
		{0, 10, 10},
		{0, 10, 0},
		{10, 10, 0},
		{0, 0, 10},
		{0, 10, 10},
		{10, 10, 10},
		{10, 0, 10},
		{0, 0, 10},
	},
	"test_files/point": [][]float64{
		{10, 10},
		{5, 5},
		{0, 10},
	},
	"test_files/polyline": [][]float64{
		{0, 0},
		{5, 5},
		{10, 10},
		{15, 15},
		{20, 20},
		{25, 25},
	},
	"test_files/polygon": [][]float64{
		{0, 0},
		{0, 5},
		{5, 5},
		{5, 0},
		{0, 0},
	},
	"test_files/multipoint": [][]float64{
		{10, 10},
		{5, 5},
		{0, 10},
	},
	"test_files/pointz": [][]float64{
		{10, 10, 100},
		{5, 5, 50},
		{0, 10, 75},
	},
	"test_files/polylinez": [][]float64{
		{0, 0, 0},
		{5, 5, 5},
		{10, 10, 10},
		{15, 15, 15},
		{20, 20, 20},
		{25, 25, 25},
	},
	"test_files/polygonz": [][]float64{
		{0, 0, 0},
		{0, 5, 5},
		{5, 5, 10},
		{5, 0, 15},
		{0, 0, 0},
	},
	"test_files/multipointz": [][]float64{
		{10, 10, 100},
		{5, 5, 50},
		{0, 10, 75},
	},
	"test_files/pointm": [][]float64{
		{10, 10, 100},
		{5, 5, 50},
		{0, 10, 75},
	},
	"test_files/polylinem": [][]float64{
		{0, 0, 0},
		{5, 5, 5},
		{10, 10, 10},
		{15, 15, 15},
		{20, 20, 20},
		{25, 25, 25},
	},
}

func TestReadPoint(t *testing.T) {
	prefix := "test_files/point"
	test_Point(t, prefix+".shp", dataForReadTests[prefix], 3)
}

func TestReadPolyLine(t *testing.T) {
	prefix := "test_files/polyline"
	test_PolyLine(t, prefix+".shp", dataForReadTests[prefix], 2)
}

func TestReadPolygon(t *testing.T) {
	prefix := "test_files/polygon"
	test_Polygon(t, prefix+".shp", dataForReadTests[prefix], 1)
}

func TestReadMultiPoint(t *testing.T) {
	prefix := "test_files/multipoint"
	test_MultiPoint(t, prefix+".shp", dataForReadTests[prefix], 1)
}

func TestReadPointZ(t *testing.T) {
	prefix := "test_files/pointz"
	test_PointZ(t, prefix+".shp", dataForReadTests[prefix], 3)
}

func TestReadPolyLineZ(t *testing.T) {
	prefix := "test_files/polylinez"
	test_PolyLineZ(t, prefix+".shp", dataForReadTests[prefix], 2)
}

func TestReadPolygonZ(t *testing.T) {
	prefix := "test_files/polygonz"
	test_PolygonZ(t, prefix+".shp", dataForReadTests[prefix], 1)
}

func TestReadMultiPointZ(t *testing.T) {
	prefix := "test_files/multipointz"
	test_MultiPointZ(t, prefix+".shp", dataForReadTests[prefix], 1)
}

func TestReadPointM(t *testing.T) {
	prefix := "test_files/pointm"
	test_PointM(t, prefix+".shp", dataForReadTests[prefix], 3)
}

func TestReadPolyLineM(t *testing.T) {
	prefix := "test_files/polylinem"
	test_PolyLineM(t, prefix+".shp", dataForReadTests[prefix], 2)
}

func TestReadPolygonM(t *testing.T) {
	prefix := "test_files/polygonm"
	test_PolygonM(t, prefix+".shp", dataForReadTests[prefix], 1)
}

func TestReadMultiPointM(t *testing.T) {
	prefix := "test_files/multipointm"
	test_MultiPointM(t, prefix+".shp", dataForReadTests[prefix], 1)
}

func TestReadMultipatch(t *testing.T) {
	prefix := "test_files/multipatch"
	test_MultiPatch(t, prefix+".shp", dataForReadTests[prefix], 1)
}
