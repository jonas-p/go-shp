package goshp

import (
	"encoding/binary"
	"os"
)

type Field struct {
	name      [11]byte
	fieldtype byte
	addr      [4]byte // not used
	size      uint8
	decimal   uint8
	padding   [14]byte
}

func StringField(name string, length uint8) Field {
	// TODO: Error checking
	field := Field{fieldtype: 'C', size: length}
	copy(field.name[:], []byte(name))
	return field
}

func (w *Writer) SetFields(fields []Field) {
	w.fields = fields
	w.dbf.Seek(int64(len(w.fields)*32+33), os.SEEK_SET)
}

func (w *Writer) StartRecord() {
	binary.Write(w.dbf, binary.LittleEndian, byte(' '))
}

func (w *Writer) WriteAttribute(field int, value string) {
	v := make([]byte, w.fields[field].size)
	copy(v[:], []byte(value))
	binary.Write(w.dbf, binary.LittleEndian, v)
}

func (w *Writer) writeDbfHeader(file *os.File) {
	headerLength := int16(len(w.fields)*32 + 33)
	recordLength := int16(1)
	for _, field := range w.fields {
		recordLength += int16(field.size)
	}

	file.Seek(0, 0)
	// version, year (YEAR-1990), month, day
	binary.Write(file, binary.LittleEndian, []byte{3, 24, 5, 3})
	// number of records
	binary.Write(file, binary.LittleEndian, w.num)
	// header length, record length
	binary.Write(file, binary.LittleEndian, []int16{headerLength, recordLength})
	// padding
	binary.Write(file, binary.LittleEndian, make([]byte, 20))

	for _, field := range w.fields {
		binary.Write(file, binary.LittleEndian, field)
	}

	// end with return
	file.WriteString("\r")
}
