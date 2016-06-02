package shp

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func compressFileToZIP(zw *zip.Writer, src, tgt string, t *testing.T) {
	r, err := os.Open(src)
	if err != nil {
		t.Fatalf("Could not open for compression %s: %v", src, err)
	}
	w, err := zw.Create(tgt)
	if err != nil {
		t.Fatalf("Could not start to compress %s: %v", tgt, err)
	}
	_, err = io.Copy(w, r)
	if err != nil {
		t.Fatalf("Could not compress contents for %s: %v", tgt, err)
	}
}

// createTempZIP packs the SHP, SHX, and DBF into a ZIP in a temporary
// directory
func createTempZIP(prefix string, t *testing.T) (dir, filename string) {
	dir, err := ioutil.TempDir("", "go-shp-test")
	if err != nil {
		t.Fatalf("Could not create temporary directory: %v", err)
	}
	base := filepath.Base(prefix)
	zipName := base + ".zip"
	w, err := os.Create(filepath.Join(dir, zipName))
	if err != nil {
		t.Fatalf("Could not create temporary zip file: %v", err)
	}
	zw := zip.NewWriter(w)
	for _, suffix := range []string{".shp", ".shx", ".dbf"} {
		compressFileToZIP(zw, prefix+suffix, base+suffix, t)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("Could not close the written zip: %v", err)
	}
	return dir, zipName
}

func getShapesZipped(prefix string, t *testing.T) (shapes []Shape) {
	dir, filename := createTempZIP(prefix, t)
	t.Logf("%s: %s %s", prefix, dir, filename)
	//defer os.RemoveAll(dir)
	zr, err := OpenZip(filepath.Join(dir, filename))
	if err != nil {
		t.Fatalf("Error when opening zip file: %v", err)
	}
	for zr.Next() {
		_, shape := zr.Shape()
		shapes = append(shapes, shape)
	}
	if err := zr.Err(); err != nil {
		t.Fatalf("Error when iterating over the shapes: %v", err)
	}

	if err := zr.Close(); err != nil {
		t.Fatalf("Could not close zipreader: %v", err)
	}
	return shapes
}

func TestZippedReadPoint(t *testing.T) {
	prefix := "test_files/point"
	points := dataForReadTests[prefix]
	shapes := getShapesZipped(prefix, t)
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
