package shp

import "fmt"

var (
	// ErrNoShpFileInZip is the error thrown when the zip file does not contain at least one shape file
	ErrNoShpFileInZip = fmt.Errorf("archive does not contain a .shp file")
	// ErrMultipleShpFileInZip is the error thrown when the zip file contains more than one shape file
	ErrMultipleShpFileInZip = fmt.Errorf("archive does contain multiple .shp files")
)
