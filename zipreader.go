package shp

import (
	"archive/zip"
	"fmt"
	"io"
	"path"
	"strings"
)

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

func countAndFirstShape(z *zip.ReadCloser) (int, string) {
	count := 0
	firstShape := ""
	for _, f := range z.File {
		if strings.HasSuffix(f.Name, ".shp") {
			count++
			if firstShape == "" {
				firstShape = f.Name
			}
		}
	}
	return count, firstShape
}

// OpenZip opens a ZIP file that contains a single shapefile.
func OpenZip(zipFilePath string) (*ZipReader, error) {
	z, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return nil, err
	}
	zr := &ZipReader{
		z: z,
	}
	count, first := countAndFirstShape(z)
	if count == 0 {
		return nil, fmt.Errorf("archive does not contain a .shp file")
	}
	if count > 1 {
		return nil, fmt.Errorf("archive does contain multiple .shp files")
	}
	prefix := strings.TrimSuffix(first, path.Ext(first))

	shp, err := openFromZIP(zr.z, prefix+".shp")
	if err != nil {
		return nil, err
	}
	// dbf is optional, so no error checking here
	dbf, _ := openFromZIP(zr.z, prefix+".dbf")
	zr.sr = SequentialReaderFromExt(shp, dbf)
	return zr, nil
}

// OpenShapeFromZip opens a shape file that is contained in a ZIP achive. The
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
