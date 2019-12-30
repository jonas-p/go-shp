package shp

import (
	"archive/zip"
	"fmt"
	"io"
	"path"
	"strings"
)

type zipShapeFileSet map[string]*zip.File

type zipFileSet map[string]zipShapeFileSet

// ZipReader provides an interface for reading Shapefiles that are compressed in a ZIP archive.
type ZipReader struct {
	sr SequentialReader
	z  *zip.ReadCloser
}

// openFromZIP is convenience function for opening the file called name that is
// compressed in z for reading.
func openFromZIP(z *zip.ReadCloser, name string) (io.ReadCloser, error) {
	for _, f := range z.File {
		if f.Name == name {
			return f.Open()

		}
	}
	return nil, fmt.Errorf("No such file in archive: %s", name)
}

// OpenZip opens a ZIP file that contains a single shapefile.
func OpenZip(zipFilePath string) (*ZipReader, error) {
	z, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return nil, err
	}
	return OpenZipFromReadCloser(z)
}

// OpenZipFromReadCloser opens a ZIP shapes from a zip ReadCloser.
func OpenZipFromReadCloser(z *zip.ReadCloser) (*ZipReader, error) {
	zr := &ZipReader{
		z: z,
	}

	shapeFiles := shapesInZip(z)

	if shapeFiles.countExt(".shp") > 1 {
		return nil, ErrMultipleShpFileInZip
	}

	for _, set := range shapeFiles {
		if set.hasExt(".shp") {
			shp, err := openFromZIP(zr.z, set[".shp"].Name)
			if err != nil {
				return nil, err
			}
			// dbf is optional, so no error checking here
			dbfFile, ok := set[".dbf"]
			if ok {
				dbf, _ := openFromZIP(zr.z, dbfFile.Name)
				zr.sr = SequentialReaderFromExt(shp, dbf)
			}
			return zr, nil
		}
	}

	return nil, ErrNoShpFileInZip
}

// ShapesInZip returns a string-slice with the names (i.e. relatives paths in
// archive file tree) of all shapes that are in the ZIP archive at zipFilePath.
func ShapesInZip(zipFilePath string) ([]string, error) {
	var names []string
	z, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return nil, err
	}

	for name, set := range shapesInZip(z) {
		if set.hasExt(".shp") {
			names = append(names, name+".shp")
		}
	}
	return names, nil
}

// shapesInZip build a map of files in the ZIP file:
// "/hello/world": {
//     ".shp": File
//     ".dbf": File
// }
func shapesInZip(z *zip.ReadCloser) zipFileSet {
	result := zipFileSet{}
	for _, f := range z.File {
		ext := path.Ext(f.Name)
		filename := f.Name[:len(ext)]

		if _, ok := result[filename]; !ok {
			result[filename] = zipShapeFileSet{}
		}

		result[filename][strings.ToLower(ext)] = f
	}
	return result
}

// hasExt detect an extention in a part of zip file map
func (z zipShapeFileSet) hasExt(ext string) bool {
	for i := range z {
		if i == ext {
			return true
		}
	}
	return false
}

// countExt counts the number of file with `ext` in a zip file
func (z zipFileSet) countExt(ext string) int {
	counter := 0
	for _, set := range z {
		if set.hasExt(ext) {
			counter++
		}
	}
	return counter
}

// OpenShapeFromZip opens a shape file that is contained in a ZIP archive. The
// parameter name is name of the shape file.
// The name of the shapefile must be a relative path: it must not start with a
// drive letter (e.g. C:) or leading slash, and only forward slashes are
// allowed. These rules are the same as in
// https://golang.org/pkg/archive/zip/#FileHeader.
func OpenShapeFromZip(zipFilePath string, name string) (*ZipReader, error) {
	z, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return nil, err
	}
	zr := &ZipReader{
		z: z,
	}

	shp, err := openFromZIP(zr.z, name)
	if err != nil {
		return nil, err
	}
	// dbf is optional, so no error checking here
	prefix := strings.TrimSuffix(name, path.Ext(name))
	dbf, _ := openFromZIP(zr.z, prefix+".dbf")
	zr.sr = SequentialReaderFromExt(shp, dbf)
	return zr, nil
}

// Close closes the ZipReader and frees the allocated resources.
func (zr *ZipReader) Close() error {
	s := ""
	err := zr.sr.Close()
	if err != nil {
		s += err.Error() + ". "
	}
	err = zr.z.Close()
	if err != nil {
		s += err.Error() + ". "
	}
	if s != "" {
		return fmt.Errorf(s)
	}
	return nil
}

// Next reads the next shape in the shapefile and the next row in the DBF. Call
// Shape() and Attribute() to access the values.
func (zr *ZipReader) Next() bool {
	return zr.sr.Next()
}

// Shape returns the shape that was last read as well as the current index.
func (zr *ZipReader) Shape() (int, Shape) {
	return zr.sr.Shape()
}

// Attribute returns the n-th field of the last row that was read. If there
// were any errors before, the empty string is returned.
func (zr *ZipReader) Attribute(n int) string {
	return zr.sr.Attribute(n)
}

// Fields returns a slice of Fields that are present in the
// DBF table.
func (zr *ZipReader) Fields() []Field {
	return zr.sr.Fields()
}

// Err returns the last non-EOF error that was encountered by this ZipReader.
func (zr *ZipReader) Err() error {
	return zr.sr.Err()
}
