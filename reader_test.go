package shp

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
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

func getShapesFromFile(prefix string, t *testing.T) (shapes []Shape) {
	dbf, err := os.Open(prefix + ".dbf")
	if err != nil {
		t.Fatal("Failed to open databaseFile: " + prefix + ".dbf (" + err.Error() + ")")
	}
	defer func() { _ = dbf.Close() }()

	shp, err := os.Open(prefix + ".shp")
	if err != nil {
		t.Fatal("Failed to open shapefile: " + prefix + ".shp (" + err.Error() + ")")
	}
	defer func() { _ = shp.Close() }()

	file, err := New(shp, WithSeekableDBF(dbf))
	if err != nil {
		t.Fatal("Failed to open shapefile: " + prefix + " (" + err.Error() + ")")
	}

	for file.Next() {
		_, shape := file.Shape()
		shapes = append(shapes, shape)
	}
	if file.Err() != nil {
		t.Errorf("Error while getting shapes for %s: %v", prefix, file.Err())
	}

	return shapes
}

type shapeGetterFunc func(string, *testing.T) []Shape

type identityTestFunc func(*testing.T, [][]float64, []Shape)

func testPoint(t *testing.T, points [][]float64, shapes []Shape) {
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

func testPolyLine(t *testing.T, points [][]float64, shapes []Shape) {
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

func testPolygon(t *testing.T, points [][]float64, shapes []Shape) {
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

func testMultiPoint(t *testing.T, points [][]float64, shapes []Shape) {
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

func testPointZ(t *testing.T, points [][]float64, shapes []Shape) {
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

func testPolyLineZ(t *testing.T, points [][]float64, shapes []Shape) {
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

func testPolygonZ(t *testing.T, points [][]float64, shapes []Shape) {
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

func testMultiPointZ(t *testing.T, points [][]float64, shapes []Shape) {
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

func testPointM(t *testing.T, points [][]float64, shapes []Shape) {
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

func testPolyLineM(t *testing.T, points [][]float64, shapes []Shape) {
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

func testPolygonM(t *testing.T, points [][]float64, shapes []Shape) {
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

func testMultiPointM(t *testing.T, points [][]float64, shapes []Shape) {
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

func testMultiPatch(t *testing.T, points [][]float64, shapes []Shape) {
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

func testshapeIdentity(t *testing.T, prefix string, getter shapeGetterFunc) {
	shapes := getter(prefix, t)
	d := dataForReadTests[prefix]
	if len(shapes) != d.count {
		t.Errorf("Number of shapes for %s read was wrong. Wanted %d, got %d.", prefix, d.count, len(shapes))
	}
	d.tester(t, d.points, shapes)
}

func TestReadBBox(t *testing.T) {
	tests := []struct {
		filename string
		want     Box
	}{
		{"test_files/multipatch", Box{0, 0, 10, 10}},
		{"test_files/multipoint", Box{0, 5, 10, 10}},
		{"test_files/multipointm", Box{0, 5, 10, 10}},
		{"test_files/multipointz", Box{0, 5, 10, 10}},
		{"test_files/point", Box{0, 5, 10, 10}},
		{"test_files/pointm", Box{0, 5, 10, 10}},
		{"test_files/pointz", Box{0, 5, 10, 10}},
		{"test_files/polygon", Box{0, 0, 5, 5}},
		{"test_files/polygonm", Box{0, 0, 5, 5}},
		{"test_files/polygonz", Box{0, 0, 5, 5}},
		{"test_files/polyline", Box{0, 0, 25, 25}},
		{"test_files/polylinem", Box{0, 0, 25, 25}},
		{"test_files/polylinez", Box{0, 0, 25, 25}},
	}
	for _, tt := range tests {
		dbf, err := os.Open(tt.filename + ".dbf")
		if err != nil {
			t.Fatal("Failed to open databaseFile: " + tt.filename + ".dbf (" + err.Error() + ")")
		}
		defer func() { _ = dbf.Close() }()

		shp, err := os.Open(tt.filename + ".shp")
		if err != nil {
			t.Fatal("Failed to open shapefile: " + tt.filename + ".shp (" + err.Error() + ")")
		}
		defer func() { _ = shp.Close() }()

		r, err := New(shp, WithSeekableDBF(dbf))

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

type testCaseData struct {
	points [][]float64
	tester identityTestFunc
	count  int
}

var dataForReadTests = map[string]testCaseData{
	"test_files/polygonm": {
		points: [][]float64{
			{0, 0, 0},
			{0, 5, 5},
			{5, 5, 10},
			{5, 0, 15},
			{0, 0, 0},
		},
		tester: testPolygonM,
		count:  1,
	},
	"test_files/multipointm": {
		points: [][]float64{
			{10, 10, 100},
			{5, 5, 50},
			{0, 10, 75},
		},
		tester: testMultiPointM,
		count:  1,
	},
	"test_files/multipatch": {
		points: [][]float64{
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
		tester: testMultiPatch,
		count:  1,
	},
	"test_files/point": {
		points: [][]float64{
			{10, 10},
			{5, 5},
			{0, 10},
		},
		tester: testPoint,
		count:  3,
	},
	"test_files/polyline": {
		points: [][]float64{
			{0, 0},
			{5, 5},
			{10, 10},
			{15, 15},
			{20, 20},
			{25, 25},
		},
		tester: testPolyLine,
		count:  2,
	},
	"test_files/polygon": {
		points: [][]float64{
			{0, 0},
			{0, 5},
			{5, 5},
			{5, 0},
			{0, 0},
		},
		tester: testPolygon,
		count:  1,
	},
	"test_files/multipoint": {
		points: [][]float64{
			{10, 10},
			{5, 5},
			{0, 10},
		},
		tester: testMultiPoint,
		count:  1,
	},
	"test_files/pointz": {
		points: [][]float64{
			{10, 10, 100},
			{5, 5, 50},
			{0, 10, 75},
		},
		tester: testPointZ,
		count:  3,
	},
	"test_files/polylinez": {
		points: [][]float64{
			{0, 0, 0},
			{5, 5, 5},
			{10, 10, 10},
			{15, 15, 15},
			{20, 20, 20},
			{25, 25, 25},
		},
		tester: testPolyLineZ,
		count:  2,
	},
	"test_files/polygonz": {
		points: [][]float64{
			{0, 0, 0},
			{0, 5, 5},
			{5, 5, 10},
			{5, 0, 15},
			{0, 0, 0},
		},
		tester: testPolygonZ,
		count:  1,
	},
	"test_files/multipointz": {
		points: [][]float64{
			{10, 10, 100},
			{5, 5, 50},
			{0, 10, 75},
		},
		tester: testMultiPointZ,
		count:  1,
	},
	"test_files/pointm": {
		points: [][]float64{
			{10, 10, 100},
			{5, 5, 50},
			{0, 10, 75},
		},
		tester: testPointM,
		count:  3,
	},
	"test_files/polylinem": {
		points: [][]float64{
			{0, 0, 0},
			{5, 5, 5},
			{10, 10, 10},
			{15, 15, 15},
			{20, 20, 20},
			{25, 25, 25},
		},
		tester: testPolyLineM,
		count:  2,
	},
}

func TestReadPoint(t *testing.T) {
	testshapeIdentity(t, "test_files/point", getShapesFromFile)
}

func TestReadPolyLine(t *testing.T) {
	testshapeIdentity(t, "test_files/polyline", getShapesFromFile)
}

func TestReadPolygon(t *testing.T) {
	testshapeIdentity(t, "test_files/polygon", getShapesFromFile)
}

func TestReadMultiPoint(t *testing.T) {
	testshapeIdentity(t, "test_files/multipoint", getShapesFromFile)
}

func TestReadPointZ(t *testing.T) {
	testshapeIdentity(t, "test_files/pointz", getShapesFromFile)
}

func TestReadPolyLineZ(t *testing.T) {
	testshapeIdentity(t, "test_files/polylinez", getShapesFromFile)
}

func TestReadPolygonZ(t *testing.T) {
	testshapeIdentity(t, "test_files/polygonz", getShapesFromFile)
}

func TestReadMultiPointZ(t *testing.T) {
	testshapeIdentity(t, "test_files/multipointz", getShapesFromFile)
}

func TestReadPointM(t *testing.T) {
	testshapeIdentity(t, "test_files/pointm", getShapesFromFile)
}

func TestReadPolyLineM(t *testing.T) {
	testshapeIdentity(t, "test_files/polylinem", getShapesFromFile)
}

func TestReadPolygonM(t *testing.T) {
	testshapeIdentity(t, "test_files/polygonm", getShapesFromFile)
}

func TestReadMultiPointM(t *testing.T) {
	testshapeIdentity(t, "test_files/multipointm", getShapesFromFile)
}

func TestReadMultiPatch(t *testing.T) {
	testshapeIdentity(t, "test_files/multipatch", getShapesFromFile)
}

func newReadSeekCloser(b []byte) readSeekCloser {
	return struct {
		io.Closer
		io.ReadSeeker
	}{
		ioutil.NopCloser(nil),
		bytes.NewReader(b),
	}
}

func TestReadInvalidShapeType(t *testing.T) {
	record := []byte{
		0, 0, 0, 0,
		0, 0, 0, 0,
		255, 255, 255, 255, // shape type
	}

	tests := []struct {
		r interface {
			Next() bool
			Err() error
		}
		name string
	}{
		{&Reader{shp: newReadSeekCloser(record), filelength: int64(len(record))}, "reader"},
		{&seqReader{shp: newReadSeekCloser(record), filelength: int64(len(record))}, "seqReader"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.r.Next() {
				t.Fatal("read unsupported shape type without stopping")
			}
			if test.r.Err() == nil {
				t.Fatal("read unsupported shape type without error")
			}
		})
	}
}
