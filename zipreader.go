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
	prefix string
	sr     SequentialReader
	z      *zip.ReadCloser
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

// OpenZip opens a ZIP file that contains a shapefile. The basename of the
// SHP, SHX, DBF file in the ZIP must be the same as the basename of the ZIP
// itself.
func OpenZip(zipFilePath string) (*ZipReader, error) {
	z, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return nil, err
	}
	b := path.Base(zipFilePath)
	zr := &ZipReader{
		prefix: strings.TrimSuffix(b, path.Ext(b)),
		z:      z,
	}
	shp, err := openFromZIP(zr.z, zr.prefix+".shp")
	if err != nil {
		return nil, err
	}
	shx, err := openFromZIP(zr.z, zr.prefix+".shx")
	if err != nil {
		return nil, err
	}
	// dbf is optional, so no error checking here
	dbf, _ := openFromZIP(zr.z, zr.prefix+".dbf")
	zr.sr = SequentialReaderFromExt(shp, shx, dbf)
	return zr, nil
}

// Close closes the the ZipReader and frees the allocated resources.
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
