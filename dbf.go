// dbf.go A small app and library to read most common dBase files and write a text file (csv)
// written by Fabien Fivaz (CSCF), released under Apache 2.0 License. See license.txt for details.
// Part of the code has freely adapted from the go-dbf project, also released under Apache 2.0 license

package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type dBaseTableInfo struct {
	dBaseLabel			  uint8				
	year                  uint32
	month                 uint8
	day                   uint8
	numberOfBytesInHeader uint16
	numberOfFields        uint16
	numberOfRecords       uint32
	numberOfBytesInRecord uint16
}

type dBaseField struct {
	fieldName         string
	fieldType         string
	fieldLength       int
	fieldDecimalCount int
}

var fType = map[string]string{
	"C": "string",
	"F": "float",
	"D": "date",
	"L": "integer",
	"M": "logical",
	"N": "memo",
}

var filename string
var table dBaseTableInfo
var info bool = false
var csv bool = true

func main() {

	a := os.Args
	
	filename = a[1]

	//filename = "test2.dbf"
	reader, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error loading file %s: %s\n", filename, err)
	}
	defer reader.Close()
	s, err := reader.Stat()
	if err != nil {
		fmt.Printf("Error reading file %s: %s\n", filename, err)
	}
	buf := make([]byte, s.Size())
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		fmt.Printf("Error reading file %s: %s\n", filename, err)
	}

	table.dBaseLabel = uint8(buf[0])
	if info {
		fmt.Printf("Opening file: %s", filename)
		switch buf[0] {
		case 3:
			fmt.Println(" (DBASE Level 5)")
		case 4:
			fmt.Println(" (DBASE Level 7)")
			fmt.Println("Converter doesn't work with DBASE v7 files.")
		default:
			fmt.Println(" (DBASE Level 3 / 4)")
		}
	}

	table.year = 1900 + uint32(buf[1])
	table.month = uint8(buf[2])
	table.day = uint8(buf[3])
	table.numberOfFields = ((uint16(buf[8]) | uint16(buf[9])<<8) - 1 - 32) / 32
	table.numberOfBytesInHeader = (uint16(buf[8]) | uint16(buf[9])<<8)
	table.numberOfRecords = (uint32(buf[4]) | uint32(buf[5])<<8 | uint32(buf[6])<<16 | uint32(buf[7])<<24)
	table.numberOfBytesInRecord = (uint16(buf[10]) | uint16(buf[11])<<8)

	if info {
		fmt.Printf("Date of last update: %d-%d-%d\n", table.year, table.month, table.day)
		fmt.Printf("Number of bytes in header: %d\n", table.numberOfBytesInHeader)
		fmt.Printf("Number of records: %d\n", table.numberOfRecords)
		fmt.Printf("Number of bytes in the record: %d\n", table.numberOfBytesInRecord)
		fmt.Println("Number of records in the table:", table.numberOfRecords)
	}

	fields := make([]dBaseField, table.numberOfFields)

	for i := 0; i < (int(table.numberOfFields)); i++ {
		nextStop := (i * 32) + 32

		fields[i].fieldName = strings.Trim(string(buf[nextStop:nextStop+10]), string([]byte{0}))
		fields[i].fieldType = string(buf[nextStop+11])
		fields[i].fieldLength = int(buf[nextStop+16])
		fields[i].fieldDecimalCount = int(buf[nextStop+17])

		if info {
			fmt.Printf("%d | %s | ", i+1, fields[i].fieldName)
			fmt.Print(fType[fields[i].fieldType])
			fmt.Printf("(%d, %d)\n", fields[i].fieldLength, fields[i].fieldDecimalCount)
		}

		if csv {
			if i < (int(table.numberOfFields) - 1) {
				fmt.Print(fields[i].fieldName, ";")
			} else {
				fmt.Print(fields[i].fieldName, "\n")
			}
		}
	}

	if csv {
		for col := 0; col < int(table.numberOfRecords); col++ {
			for row := 0; row < int(table.numberOfFields); row++ {
				nextStop := int(table.numberOfBytesInHeader)
				nextStop = nextStop + (col * int(table.numberOfBytesInRecord))
				recordOffset := 1

				for k := 0; k < 13; k++ {
					if k == row {
						break
					} else {
						recordOffset += int(fields[k].fieldLength)
					}
				}
				cellValue := buf[(nextStop + recordOffset):((nextStop + recordOffset) + int(fields[row].fieldLength))]
				if row < (int(table.numberOfFields) - 1) {
					fmt.Print(strings.TrimSpace(string(cellValue)), ";")
				} else {
					fmt.Print(strings.TrimSpace(string(cellValue)))
				}
			}
			fmt.Print("\n")
		}
	}
}
