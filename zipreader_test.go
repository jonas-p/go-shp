package shp

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
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
	for prefix := range dataForReadTests {
		t.Logf("Testing zipped reading for %s", prefix)
		testshapeIdentity(t, prefix, getShapesZipped)
	}
}

func unzipToTempDir(t *testing.T, p string) string {
	td, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("%v", err)
	}
	zip, err := zip.OpenReader(p)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer zip.Close()
	for _, f := range zip.File {
		_, fn := path.Split(f.Name)
		pn := filepath.Join(td, fn)
		t.Logf("Uncompress: %s -> %s", f.Name, pn)
		w, err := os.Create(pn)
		if err != nil {
			t.Fatalf("Cannot unzip %s: %v", p, err)
		}
		defer w.Close()
		r, err := f.Open()
		if err != nil {
			t.Fatalf("Cannot unzip %s: %v", p, err)
		}
		defer r.Close()
		_, err = io.Copy(w, r)
		if err != nil {
			t.Fatalf("Cannot unzip %s: %v", p, err)
		}
	}
	return td
}

// TestZipReaderAttributes reads the same shapesfile twice, first directly from
// the Shp with a Reader, and, second, from a zip. It compares the fields as
// well as the shapes and the attributes. For this test, the Shapes are
// considered to be equal if their bounding boxes are equal.
func TestZipReaderAttribute(t *testing.T) {
	b := "ne_110m_admin_0_countries"
	skipOrDownloadNaturalEarth(t, b+".zip")
	d := unzipToTempDir(t, b+".zip")
	defer os.RemoveAll(d)

	dbf, err := os.Open(b + ".dbf")
	if err != nil {
		t.Fatal("Failed to open databaseFile: " + b + ".dbf (" + err.Error() + ")")
	}
	defer func() { _ = dbf.Close() }()

	shp, err := os.Open(b + ".shp")
	if err != nil {
		t.Fatal("Failed to open shapefile: " + b + ".shp (" + err.Error() + ")")
	}
	defer func() { _ = shp.Close() }()

	lr, err := New(shp, WithSeekableDBF(dbf))
	if err != nil {
		t.Fatal(err)
	}

	zr, err := OpenZip(b + ".zip")
	if os.IsNotExist(err) {
		t.Skipf("Skipping test, as Natural Earth dataset wasn't found")
	}
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
			t.Fatalf("Bounding boxes for shape #%d do not match", ln+1)
		}
		for i := range fsl {
			la := lr.Attribute(i)
			za := zr.Attribute(i)
			if la != za {
				t.Fatalf("Shape %d: Attribute %d (%s) are unequal: '%s' vs '%s'",
					ln+1, i, fsl[i].String(), la, za)
			}
		}
	}
	if lr.Err() != nil {
		t.Logf("Reader error: %v / ZipReader error: %v", lr.Err(), zr.Err())
		t.FailNow()
	}
}

func skipOrDownloadNaturalEarth(t *testing.T, p string) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		dl := false
		for _, a := range os.Args {
			if a == "download" {
				dl = true
				break
			}
		}
		u := "http://www.naturalearthdata.com/http//www.naturalearthdata.com/download/110m/cultural/ne_110m_admin_0_countries.zip"
		if !dl {
			t.Skipf("Skipped, as %s does not exist. Consider calling tests with '-args download` "+
				"or download manually from '%s'", p, u)
		} else {
			t.Logf("Downloading %s", u)
			w, err := os.Create(p)
			if err != nil {
				t.Fatalf("Could not create %q: %v", p, err)
			}
			defer w.Close()
			resp, err := http.Get(u)
			if err != nil {
				t.Fatalf("Could not download %q: %v", u, err)
			}
			defer resp.Body.Close()
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				t.Fatalf("Could not download %q: %v", u, err)
			}
			t.Logf("Download complete")
		}
	}
}

func TestNaturalEarthZip(t *testing.T) {
	type metaShape struct {
		Attributes map[string]string
		Shape
	}
	p := "ne_110m_admin_0_countries.zip"
	skipOrDownloadNaturalEarth(t, p)
	zr, err := OpenZip(p)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()

	fs := zr.Fields()
	if len(fs) != 63 {
		t.Fatalf("Expected 63 columns in Natural Earth dataset, got %d", len(fs))
	}
	var metas []metaShape
	for zr.Next() {
		m := metaShape{
			Attributes: make(map[string]string),
		}
		_, m.Shape = zr.Shape()
		for n := range fs {
			m.Attributes[fs[n].String()] = zr.Attribute(n)
		}
		metas = append(metas, m)
	}
	if zr.Err() != nil {
		t.Fatal(zr.Err())
	}
	for _, m := range metas {
		t.Log(m.Attributes["name"])
	}
}

func TestShapesInZip(t *testing.T) {
	p := "ne_110m_admin_0_countries"
	skipOrDownloadNaturalEarth(t, p+".zip")

	z, err := zip.OpenReader(p + ".zip")
	if err != nil {
		t.Fatal(err)
	}

	shapeFiles := shapesInZip(z)

	if shapeFiles.countExt(".shp") != 1 {
		t.Fatalf("Expected 1 shp file, got %d", shapeFiles.countExt(".shp"))
	}

	set, ok := shapeFiles[p]
	if !ok {
		t.Fatalf("Expected to find file [%s]", p)
	}

	if _, ok := set[".shp"]; !ok {
		t.Fatalf("Expected to find file [%s.shp]", p)
	}

	if _, ok := set[".dbf"]; !ok {
		t.Fatalf("Expected to find file [%s.dbf]", p)
	}

}
