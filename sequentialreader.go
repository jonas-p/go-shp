package shp

import (
	"fmt"
	"io"
)

// SequentialReader is the interface that allows reading shapes and attributes one after another. It also embeds io.Closer.
type SequentialReader interface {
	// Close() frees the resources allocated by the SequentialReader.
	io.Closer

	// Next() tries to advance the reading by one shape and one attribute row
	// and returns true if the read operation could be performed without any
	// error.
	Next() bool

	// Shape returns the index and the last read shape. If the SequentialReader
	// encountered any errors, nil is returned for the Shape.
	Shape() (int, Shape)

	// Attribute returns the value of the n-th attribute in the current row. If
	// the SequentialReader encountered any errors, the empty string is
	// returned.
	Attribute(n int) string

	// Fields returns the fields of the database. If the SequentialReader
	// encountered any errors, nil is returned.
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

// seqReader implements SequentialReader based on external io.ReadCloser
// instances
type seqReader struct {
	shp, shx, dbf io.ReadCloser
	err           error
}

// Next implements a method of interface SequentialReader for seqReader.
func (sr *seqReader) Next() bool {
	if sr.err != nil {
		return false
	}
	sr.err = fmt.Errorf("Not implemented yet")
	// TODO implement this
	return false
}

// Shape implements a method of interface SequentialReader for seqReader.
func (sr *seqReader) Shape() (int, Shape) {
	panic("not implemented")
}

// Attribute implements a method of interface SequentialReader for seqReader.
func (sr *seqReader) Attribute(n int) string {
	if sr.err != nil {
		return ""
	}
	panic("not implemented")
	// TODO implement this
}

// Err returns the first non-EOF error that was encountered.
func (sr *seqReader) Err() error {
	if sr.err == io.EOF {
		return nil
	}
	return sr.err
}

// Close closes the seqReader and free all the allocated resources.
func (sr *seqReader) Close() error {
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
func (sr *seqReader) Fields() []Field {
	// TODO implement
	return nil
}

// SequentialReaderFromExt returns a new SequentialReader that interprets shp
// as a source of shapes that are indexed in shx and whose attributes can be
// retrieved from dbf.
func SequentialReaderFromExt(shp, shx, dbf io.ReadCloser) SequentialReader {
	return &seqReader{shp: shp, shx: shx, dbf: dbf}
}
