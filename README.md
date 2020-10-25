go-shp
======

[![Build Status](https://travis-ci.org/jonas-p/go-shp.svg?branch=master)](https://travis-ci.org/jonas-p/go-shp)
[![Build status](https://ci.appveyor.com/api/projects/status/b64sntax4kxlouxa?svg=true)](https://ci.appveyor.com/project/fawick/go-shp)
[![Go Report Card](https://goreportcard.com/badge/github.com/jonas-p/go-shp)](https://goreportcard.com/report/github.com/jonas-p/go-shp)
[![Codevov](https://codecov.io/gh/jonas-p/go-shp/branch/master/graphs/badge.svg)](https://codecov.io/gh/jonas-p/go-shp)

Go library for reading and writing ESRI Shapefiles. This is a pure Golang implementation based on the ESRI Shapefile technical description.

### Usage
#### Installation

    go get github.com/jonas-p/go-shp
    
#### Importing

```go
import "github.com/jonas-p/go-shp"
```

### Examples
#### Reading a shapefile

```go
// open a shapefile (SHP) for reading
shp, err := os.Open(prefix + ".shp")
if err != nil {
	log.Fatal(err)
}
defer func() { _ = shp.Close() }()

// open a shapefile (DBF) for reading (needed to read attributes)
dbf, err := os.Open(prefix + ".dbf")
if err != nil {
	log.Fatal(err)
}
defer func() { _ = dbf.Close() }()

// Create the Reader
shape, err := New(shp, WithSeekableDBF(dbf))
if err != nil {
	log.Fatal(err)
}
	
// fields from the attribute table (DBF)
fields := shape.Fields()
	
// loop through all features in the shapefile
for shape.Next() {
	n, p := shape.Shape()
	
	// print feature
	fmt.Println(reflect.TypeOf(p).Elem(), p.BBox())
	
	// print attributes
	for k, f := range fields {
		val := shape.ReadAttribute(n, k)
		fmt.Printf("\t%v: %v\n", f, val)
	}
	fmt.Println()
}
```

#### Creating a shapefile

```go
// points to write
points := []shp.Point{
	shp.Point{10.0, 10.0},
	shp.Point{10.0, 15.0},
	shp.Point{15.0, 15.0},
	shp.Point{15.0, 10.0},
}
	
// fields to write
fields := []shp.Field{
	// String attribute field with length 25
	shp.StringField("NAME", 25),
}
	
// create and open a shapefile for writing points
shape, err := shp.Create("points.shp", shp.POINT)
if err != nil { log.Fatal(err) }
defer shape.Close()
	
// setup fields for attributes
shape.SetFields(fields)
	
// write points and attributes
for n, point := range points {
	shape.Write(&point)
	
	// write attribute for object n for field 0 (NAME)
	shape.WriteAttribute(n, 0, "Point " + strconv.Itoa(n + 1))
}
```

### Resources

- [Documentation on godoc.org](http://godoc.org/github.com/jonas-p/go-shp)
- [ESRI Shapefile Technical Description](http://www.esri.com/library/whitepapers/pdfs/shapefile.pdf)
