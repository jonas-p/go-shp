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
	defer os.RemoveAll(dir)
	zr, err := OpenZip(filepath.Join(dir, filename))
	if err != nil {
		t.Errorf("Error when opening zip file: %v", err)
	}
	for zr.Next() {
		_, shape := zr.Shape()
		shapes = append(shapes, shape)
	}
	if err := zr.Err(); err != nil {
		t.Errorf("Error when iterating over the shapes: %v", err)
	}

	if err := zr.Close(); err != nil {
		t.Errorf("Could not close zipreader: %v", err)
	}
	return shapes
}

func TestZipReader(t *testing.T) {
	for prefix, _ := range dataForReadTests {
		t.Logf("Testing zipped reading for %s", prefix)
		test_shapeIdentity(t, prefix, getShapesZipped)
	}
}

func TestNaturalEarthZip(t *testing.T) {
	zr, err := OpenZip("ne_110m_admin_0_countries.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	t.Log(len(zr.Fields()))
	for zr.Next() {
	}
	if zr.Err() != nil {
		t.Fatal(zr.Err())
	}
}
