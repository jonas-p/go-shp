package shp

import "io"

// errReader is a helper to perform multiple successive read from another reader
// and do the error checking only once afterwards. It will not perform any new
// reads in case there was an error encountered earlier.
type errReader struct {
	io.Reader
	e error
}

func (er *errReader) Read(p []byte) (n int, err error) {
	if er.e != nil {
		return 0, er.e
	}
	n, er.e = er.Reader.Read(p)
	return n, er.e
}
