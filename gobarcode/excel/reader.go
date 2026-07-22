// Package excel is used to reference and read the columns from the file barcodes are being generated from.
package excel

import (
	"errors"
	"fmt"

	"github.com/xuri/excelize/v2"
)

type LabelInfo struct {
	fname             string
	title             string
	upc               string
	Sheetmap          map[int]string `json:"sheet_map"`
	SelectedSheet     int            `json:"selected_sheet"`
	SelectedSheetName string         `json:"selected_sheet_name"`
	HeaderRow         int            `json:"header_row"`
	HeaderCol         string         `json:"header_col"`
	HeaderRowValues   []string       `json:"header_row_values"`
}

func GetWorkBookInfo(name string) (*LabelInfo, error) {
	li := &LabelInfo{fname: name}
	f, err := excelize.OpenFile(name)
	if err != nil {
		return li, err
	}
	defer f.Close()
	li.Sheetmap = f.GetSheetMap()
	li.SelectedSheet = 1

	li.SelectedSheetName = li.Sheetmap[li.SelectedSheet]
	rows, err := f.Rows(li.Sheetmap[li.SelectedSheet])
	if err != nil {
		return li, err
	}
	defer rows.Close()
	if rows.Next() {
		cols, err := rows.Columns()
		if err != nil {
			return li, err
		}
		li.HeaderRowValues = cols
	}
	return li, err
}

func (li *LabelInfo) GetHeaderRowValues(hr int) (*LabelInfo, error) {
	var err error
	if li.SelectedSheetName == "" {
		err = errors.New("error: missing sheet name")
		return nil, err
	}
	if li.fname == "" {
		err = errors.New("error: no selected file")
		return nil, err
	}
	file, err := excelize.OpenFile(li.fname)
	if err != nil {
		return nil, err
	}
	rows, err := file.Rows(li.SelectedSheetName)
	if err != nil {
		return nil, err
	}
	fmt.Println("reading rows")
	defer rows.Close()
	currentrow := 0
	for rows.Next() {
		if currentrow == hr {
			cols, err := rows.Columns()
			if err != nil {
				return nil, err
			}
			li.HeaderRowValues = cols
			return li, err
		}
		currentrow++
	}
	fmt.Println("could not find row")
	err = errors.New("error: row not found")
	return li, err
}
