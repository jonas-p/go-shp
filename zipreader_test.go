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

// TestZipReaderAttributes reads the same shapesfile twice, first directly from
// the Shp with a Reader, and, second, from a zip. It compares the fields as
// well as the shapes and the attributes. For this test, the Shapes are
// considered to be equal if their bounding boxes are equal.
func TestZipReaderAttribute(t *testing.T) {
	lr, err := Open("ne_110m_admin_0_countries.shp")
	if err != nil {
		t.Fatal(err)
	}
	defer lr.Close()
	zr, err := OpenZip("ne_110m_admin_0_countries.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	fsl := lr.Fields()
	fsz := zr.Fields()
	if len(fsl) != len(fsz) {
		t.Fatalf("Number of attributes do not match: Wanted %d, got %d", len(fsl), len(fsz))
	}
	for i := range fsl {
		if fsl[i] != fsz[i] {
			t.Fatalf("Attribute %d (%s) does not match (%s)", i, fsl[i], fsz[i])
		}
	}
	for zr.Next() && lr.Next() {
		ln, ls := lr.Shape()
		zn, zs := zr.Shape()
		if ln != zn {
			t.Fatalf("Sequence number wrong: Wanted %d, got %d", ln, zn)
		}
		if ls.BBox() != zs.BBox() {
			t.Fatalf("Bounding boxes for shape #%d do not match", ln)
		}
		for i := range fsl {
			la := lr.Attribute(i)
			za := zr.Attribute(i)
			if la != za {
				t.Fatalf("Shape %d: Attribute %d (%s) are unequal: '%s' vs '%s'",
					ln, i, fsl[i].FooString(), la, za)
			}
		}
	}
	if lr.Err() != nil {
		t.Logf("Reader error: %v / ZipReader error: %v", lr.Err(), zr.Err())
		t.FailNow()
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
