// dbf.go A small app and library to read most common dBase files and write a text file (csv)
// written by Fabien Fivaz, released under Apache 2.0 License. See license.txt for details.
// Part of the code has freely adapted from the go-dbf project (code.google.com/p/go-dbf/) released under Apache 2.0 license

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type dBaseTableInfo struct {
	dBaseLabel            uint8
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

var outfilename string
var dirname string
var table dBaseTableInfo

// Switch from informations about the dbf (info is true) file to exporting the file (csv is true)
// with something like dbf_convert test.dbf > test.csv
var info bool = false
var csv bool = true

// flags
var sf string  // single file name flag
var of bool    // output file flag (input file name + csv) else
var inf string // single file name for file infos
var df string  // directory name flag
var cf string    // concatenate flag
var afn bool   // add filename in concatenation file
var vf bool    // verbose flag

func init() {
	flag.StringVar(&sf, "f", "", "convert a single file")
	flag.BoolVar(&of, "o", false, "output to file instead of stdout. File name is the same, except csv extension.")
	flag.StringVar(&inf, "i", "", "return file informations")
	flag.StringVar(&df, "d", "", "convert all files in directory")
	flag.StringVar(&cf, "c", "", "concatenate all files from directory in a single file")
	flag.BoolVar(&afn, "a", true, "add filename to concatenation file")
	flag.BoolVar(&vf, "v", false, "verbose")
}

func dbfInfo(infile string) bool {
	reader, err := os.Open(infile)
	if err != nil {
		panic(fmt.Sprintf("Error loading file %s: %s\n", infile, err))
	}

	defer reader.Close()

	s, err := reader.Stat()
	if err != nil {
		panic(fmt.Sprintf("Error loading file %s: %s\n", infile, err))
	}

	buf := make([]byte, s.Size())
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		panic(fmt.Sprintf("Error loading file %s: %s\n", infile, err))
	}

	table.dBaseLabel = uint8(buf[0])
	fmt.Println("-----------------------------------------------------------")
	fmt.Printf("%s: DBase file informations", infile)
	fmt.Println("\n-----------------------------------------------------------")
	fmt.Print("File encoding: ")
	switch buf[0] {
	case 3:
		fmt.Println("DBASE Level 5")
	case 4:
		fmt.Println("DBASE Level 7")
		return false
	default:
		fmt.Println("DBASE Level 3 / 4")
	}
	fmt.Println("For details, see https://en.wikipedia.org/wiki/.dbf")
	fmt.Println("-----------------------------------------------------------")
	table.year = 1900 + uint32(buf[1])
	table.month = uint8(buf[2])
	table.day = uint8(buf[3])
	table.numberOfFields = ((uint16(buf[8]) | uint16(buf[9])<<8) - 1 - 32) / 32
	table.numberOfBytesInHeader = (uint16(buf[8]) | uint16(buf[9])<<8)
	table.numberOfRecords = (uint32(buf[4]) | uint32(buf[5])<<8 | uint32(buf[6])<<16 | uint32(buf[7])<<24)
	table.numberOfBytesInRecord = (uint16(buf[10]) | uint16(buf[11])<<8)

	fmt.Printf("Date of last update: %d-%d-%d\n", table.year, table.month, table.day)
	fmt.Printf("Number of bytes in header: %d\n", table.numberOfBytesInHeader)
	fmt.Printf("Number of records: %d\n", table.numberOfRecords)
	fmt.Printf("Number of bytes in the record: %d\n", table.numberOfBytesInRecord)
	fmt.Println("Number of records in the table:", table.numberOfRecords)
	fmt.Println("-----------------------------------------------------------")

	fields := make([]dBaseField, table.numberOfFields)

	fmt.Println("List of fields:")
	fmt.Println("ID | NAME | FORMAT")
	for i := 0; i < (int(table.numberOfFields)); i++ {
		nextStop := (i * 32) + 32

		fields[i].fieldName = strings.Trim(string(buf[nextStop:nextStop+10]), string([]byte{0}))
		fields[i].fieldType = string(buf[nextStop+11])
		fields[i].fieldLength = int(buf[nextStop+16])
		fields[i].fieldDecimalCount = int(buf[nextStop+17])

		fmt.Printf("%d | %s | ", i+1, fields[i].fieldName)
		fmt.Print(fType[fields[i].fieldType])
		fmt.Printf("(%d, %d)\n", fields[i].fieldLength, fields[i].fieldDecimalCount)
	}

	return true

}

func readDbf(infile string, tofile bool) bool {

	var outfile = strings.TrimSuffix(infile, filepath.Ext(infile)) + ".csv"

	reader, err := os.Open(infile)
	if err != nil {
		panic(fmt.Sprintf("Error loading file %s: %s\n", infile, err))
	}

	defer reader.Close()

	s, err := reader.Stat()
	if err != nil {
		panic(fmt.Sprintf("Error loading file %s: %s\n", infile, err))
	}

	buf := make([]byte, s.Size())
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		panic(fmt.Sprintf("Error loading file %s: %s\n", infile, err))
	}

	table.numberOfFields = ((uint16(buf[8]) | uint16(buf[9])<<8) - 1 - 32) / 32
	table.numberOfBytesInHeader = (uint16(buf[8]) | uint16(buf[9])<<8)
	table.numberOfRecords = (uint32(buf[4]) | uint32(buf[5])<<8 | uint32(buf[6])<<16 | uint32(buf[7])<<24)
	table.numberOfBytesInRecord = (uint16(buf[10]) | uint16(buf[11])<<8)

	fields := make([]dBaseField, table.numberOfFields)
	var header []string

	for i := 0; i < (int(table.numberOfFields)); i++ {
		nextStop := (i * 32) + 32

		fields[i].fieldName = strings.Trim(string(buf[nextStop:nextStop+10]), string([]byte{0}))
		fields[i].fieldType = string(buf[nextStop+11])
		fields[i].fieldLength = int(buf[nextStop+16])
		fields[i].fieldDecimalCount = int(buf[nextStop+17])

		header = append(header, fields[i].fieldName)

	}

	if !tofile {
		fmt.Println(strings.Join(header, ";"))
	} else {
		if vf {
			fmt.Printf("writing to file %s.csv", outfile)
			fmt.Print("\n")
		}
		f, err := os.Create(outfile)
		if err != nil {
			panic(fmt.Sprintf("Error creating file %s: %s\n", outfile, err))
		}
		defer f.Close()

		f.Write([]byte(strings.Join(header, ";")))
		f.Write([]byte("\n"))
	}

	var coldata []string

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
			coldata = append(coldata, strings.TrimSpace(string(cellValue)))
		}
		if !tofile {
			fmt.Println(strings.Join(coldata, ";"))
		} else {
			f, err := os.OpenFile(outfile, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				panic(fmt.Sprintf("Error creating file %s: %s\n", outfile, err))
			}
			defer f.Close()

			f.Write([]byte(strings.Join(coldata, ";")))
			f.Write([]byte("\n"))
		}
		coldata = nil
	}

	return true
}

func concatDbf(indir string) bool {

	cnt := 0

	files, _ := ioutil.ReadDir(indir)
	for _, f := range files {

		cnt++

		reader, err := os.Open(indir + "/" + f.Name())
		if err != nil {
			panic(fmt.Sprintf("Error loading file %s: %s\n", f.Name(), err))
		}

		defer reader.Close()

		s, err := reader.Stat()
		if err != nil {
			panic(fmt.Sprintf("Error loading file %s: %s\n", f.Name(), err))
		}

		buf := make([]byte, s.Size())
		_, err = io.ReadFull(reader, buf)
		if err != nil {
			panic(fmt.Sprintf("Error loading file %s: %s\n", f.Name(), err))
		}

		table.numberOfFields = ((uint16(buf[8]) | uint16(buf[9])<<8) - 1 - 32) / 32
		table.numberOfBytesInHeader = (uint16(buf[8]) | uint16(buf[9])<<8)
		table.numberOfRecords = (uint32(buf[4]) | uint32(buf[5])<<8 | uint32(buf[6])<<16 | uint32(buf[7])<<24)
		table.numberOfBytesInRecord = (uint16(buf[10]) | uint16(buf[11])<<8)

		fields := make([]dBaseField, table.numberOfFields)
		var header []string

		if afn {
			header = append(header, "FROM_FILE")
		}

		for i := 0; i < (int(table.numberOfFields)); i++ {
			nextStop := (i * 32) + 32

			fields[i].fieldName = strings.Trim(string(buf[nextStop:nextStop+10]), string([]byte{0}))
			fields[i].fieldType = string(buf[nextStop+11])
			fields[i].fieldLength = int(buf[nextStop+16])
			fields[i].fieldDecimalCount = int(buf[nextStop+17])

			header = append(header, fields[i].fieldName)

		}

		if cnt == 1 {
			fmt.Println(strings.Join(header, ";"))
		}

		var coldata []string

		for col := 0; col < int(table.numberOfRecords); col++ {
			if afn {
				coldata = append(coldata, f.Name())
			}
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
				coldata = append(coldata, strings.TrimSpace(string(cellValue)))
			}

			fmt.Println(strings.Join(coldata, ";"))
			coldata = nil
		}

	}
	return true
}

func main() {

	flag.Parse()

	if df != "" {
		cnt := 0 // Counter

		files, _ := ioutil.ReadDir(df)
		for _, f := range files {
			if vf {
				fmt.Printf("Converting %s...", f.Name())
			}
			if readDbf(df+"/"+f.Name(), true) {
				cnt++
			} else {
				fmt.Println("Error converting %s...", f.Name())
			}
		}

		if vf {
			fmt.Printf("%d files converted.", cnt)
			fmt.Print("\n")
		}

	} else if sf != "" {
		if !readDbf(sf, of) {
			fmt.Println("Error")
		}
	} else if inf != "" {
		if !dbfInfo(inf) {
			fmt.Println("Error")
		}
	} else if cf != "" {
		if !concatDbf(cf) {
			fmt.Println("Error")
		}
	} else {
		fmt.Print("Arguments required. Please use -h for help.\n")
	}

}
