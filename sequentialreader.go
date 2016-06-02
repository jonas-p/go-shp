package shp

import (
	"fmt"
	"io"
)

// SequentialReader is the interface that allows reading shapes and attributes one after another. It also embeds io.Closer.
type SequentialReader interface {
	// Close() frees the resources allocated by the SequentialReader.
	io.Closer

	// Next() tries to advance the reading by one shape and one attribute row and returns in
	// case no error other than io.EOF was encountered.
	Next() bool

	// Shape returns the index and the last read shape.
	Shape() (int, Shape)

	// Attribute returns the value of the n-th attribute in the current row.
	Attribute(n int) string

	// Fields returns the fields of the database.
	Fields() []Field

	// Err returns the last non-EOF error encountered.
	Err() error
}

// Attributes() returns all attributes of the shape that sr was last advanced to.
func Attributes(sr SequentialReader) []string {
	if sr.Err() != nil {
		return nil
	}
	s := make([]string, len(sr.Fields()))
	for i := range s {
		s[i] = sr.Attribute(i)
	}
	return s
}

// AttributeCount returns the number of fields of the database.
func AttributeCount(sr SequentialReader) int {
	return len(sr.Fields())
}

// fromExternal implements SequentialReader based on external io.ReadClosers
type fromExternal struct {
	shp, shx, dbf io.ReadCloser
	err           error
}

// Next implements a method of interface SequentialReader for fromExternal.
func (sr *fromExternal) Next() bool {
	if sr.err != nil {
		return false
	}
	sr.err = fmt.Errorf("Not implemented yet")
	// TODO implement this
	return false
}

// Shape implements a method of interface SequentialReader for fromExternal.
func (sr *fromExternal) Shape() (int, Shape) {
	// TODO implement this
	return 0, nil
}

// Attribute implements a method of interface SequentialReader for fromExternal.
func (sr *fromExternal) Attribute(n int) string {
	// TODO implement this
	return ""
}

// Err returns the first non-EOF error that was encountered.
func (sr *fromExternal) Err() error {
	if sr.err == io.EOF {
		return nil
	}
	return sr.err
}

// Close returns the first non-EOF error that was encountered.
func (sr *fromExternal) Close() error {
	var s string
	if sr.err != nil {
		s = sr.err.Error() + ". "
	}
	if err := sr.shp.Close(); err != nil {
		s += err.Error() + ". "
	}
	if err := sr.shx.Close(); err != nil {
		s += err.Error() + ". "
	}
	if err := sr.dbf.Close(); err != nil {
		s += err.Error() + ". "
	}
	if s != "" {
		sr.err = fmt.Errorf(s)
	}
	return sr.err
}

// Fields returns a slice of the Fields that are present in the DBF table.
func (sr *fromExternal) Fields() []Field {
	// TODO implement
	return nil
}

// SequentialReaderFromExt returns a new SequentialReader that interprets shp
// as a source of shapes that are indexed in shx and whose attributes can be
// retrieved from dbf.
func SequentialReaderFromExt(shp, shx, dbf io.ReadCloser) SequentialReader {
	return &fromExternal{shp: shp, shx: shx, dbf: dbf}
}
