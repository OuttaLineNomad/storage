package storage

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/OuttaLineNomad/excelxml"
	"github.com/OuttaLineNomad/storage/auth"
	"github.com/WedgeNix/xls"
	"github.com/tealeg/xlsx"
	"google.golang.org/api/drive/v3"

	"golang.org/x/oauth2/google"
)

type DriverService struct {
	*drive.Service
}

// Storage keeps services of all drives.
type Storage struct {
	Drive DriverService
	/*AWS   *awsapi.Controller*/
}

// New starts storage all services.
// NOT DONE YET
// func New() (*Storage, error) {
// 	// Google Drive Service
// 	ctx := context.Background()

// 	b, err := ioutil.ReadFile("credentials/google_client_secret.json")
// 	if err != nil {
// 		return nil, err
// 	}

// 	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
// 	if err != nil {
// 		return nil, err
// 	}
// 	client := auth.GetGoogleClient(ctx, config)

// 	srv, err := drive.New(client)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// AWS API Service
// 	myaws, err := awsapi.New()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &Storage{Drive: srv, AWS: myaws}, nil
// }

type Error struct {
	Func string
	Msg  string
	Err  error
}

func (er *Error) Error() string {
	return `storage: ` + er.Func + `: ` + er.Msg + `: ` + er.Err.Error()
}

// NewGoogle starts storage all services.
func NewGoogle() (*Storage, error) {
	// Google Drive Service
	ctx := context.Background()

	b, err := ioutil.ReadFile("credentials/drive_client_secret.json")
	if err != nil {
		return nil, &Error{"NewGoogle", "", err}
	}

	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return nil, &Error{"NewGoogle", "", err}
	}
	client := auth.GetGoogleClient(ctx, config)

	srv, err := drive.New(client)
	if err != nil {
		return nil, &Error{"NewGoogle", "", err}
	}
	return &Storage{Drive: DriverService{srv}}, nil
}

// XLToCSV converts excel file data to csv file data.
// Uses sheets parameter to save separate csv files.
//
// This function might be helpful to others, but it is made mostly
// for wedgenix use. Helps convert clunky xls/xlsx files into lighter csv files.
func XLToCSV(name string, data []byte, sheets ...int) ([][]byte, error) {
	if len(name) == 0 {
		return nil, &Error{"XLToCSV", "", errors.New("no file name")}
	}
	if len(sheets) == 0 {
		sheets = []int{0}
	}
	err := ioutil.WriteFile(name, data, os.ModePerm)
	if err != nil {
		return nil, &Error{"XLToCSV", "creating file", err}
	}
	defer os.Remove(name)

	var fileSheets [][][]string

	ext := strings.ToLower(filepath.Ext(name))

again:
	switch ext {
	case ".xlsx":
		c, err := xlsx.FileToSlice(name)
		if err != nil {
			return nil, &Error{"XLToCSV", "file to slice .xlsx", err}
		}
		fileSheets = c
	case ".xls":
		f, err := xls.Open(name, "")
		if err != nil {
			if err.Error() == "not an excel file" {
				data, err := testXML(name)
				if err != nil {
					if !strings.Contains(err.Error(), "file is not an excel xml file") {
						return nil, &Error{"XLToCSV", "find xml", err}
					}
				}
				if len(data) > 0 {
					fileSheets = [][][]string{data}
					break
				}
				ext = ".txt"
				goto again
			}
			return nil, &Error{"XLToCSV", "open .xls", err}
		}
		fileSheets = [][][]string{f.ReadAllCells()}
	case ".txt":
		noQuote := regexp.MustCompile(`"([^"]+)"`)
		b, err := ioutil.ReadFile(name)
		if err != nil {
			return nil, &Error{"XLToCSV", "read file .txt", err}
		}
		data := string(b)
		splitRow := strings.Split(data, "\n")
		newData := [][]string{}
		for _, row := range splitRow {
			cells := strings.Split(row, "\t")
			for i, cell := range cells {
				if noQuote.MatchString(cell) {
					cells[i] = noQuote.FindStringSubmatch(cell)[1]
				}
			}
			newData = append(newData, cells)
		}
		fileSheets = [][][]string{newData}

	default:
		return nil, &Error{"XLToCSV", "default", errors.New("'" + ext + "' is not an Excel format")}
	}
	files := [][]byte{}
	for _, sheet := range sheets {
		buf := bytes.Buffer{}
		c := csv.NewWriter(&buf)
		c.WriteAll(fileSheets[sheet])
		files = append(files, buf.Bytes())
	}

	return files, nil
}

func testXML(name string) ([][]string, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, &Error{"XLToCSV", "testXML", err}
	}
	defer f.Close()
	sheets, err := excelxml.SliceXML(f)
	if err != nil {
		return nil, &Error{"XLToCSV", "testXML", err}
	}
	return sheets.Worksheets[0].Table, nil
}
