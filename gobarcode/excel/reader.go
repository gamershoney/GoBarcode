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
	Err               error          `json:"error"`
}

func GetWorkBookInfo(name string) *LabelInfo {
	li := &LabelInfo{fname: name}
	f, err := excelize.OpenFile(name)
	if err != nil {
		li.Err = err
		return li
	}
	defer f.Close()
	li.Sheetmap = f.GetSheetMap()
	li.SelectedSheet = 1

	li.SelectedSheetName = li.Sheetmap[li.SelectedSheet]
	rows, err := f.Rows(li.Sheetmap[li.SelectedSheet])
	if err != nil {
		li.Err = err

		return li
	}
	defer rows.Close()
	if rows.Next() {
		cols, err := rows.Columns()
		if err != nil {
			li.Err = err
			return li
		}
		li.HeaderRowValues = cols
	}
	return li
}

func (li *LabelInfo) GetHeaderRowValues(hr int) *LabelInfo {
	if li.SelectedSheetName == "" {
		li.Err = errors.New("error: missing sheet name")
		return li
	}
	if li.fname == "" {
		li.Err = errors.New("error: no selected file")
	}
	file, err := excelize.OpenFile(li.fname)
	if err != nil {
		li.Err = err
		return li
	}
	rows, err := file.Rows(li.SelectedSheetName)
	if err != nil {
		li.Err = err
		return li
	}
	fmt.Println("reading rows")
	defer rows.Close()
	currentrow := 0
	for rows.Next() {
		if currentrow == hr {
			cols, err := rows.Columns()
			if err != nil {
				li.Err = err
			}
			li.HeaderRowValues = cols
			return li
		}
		currentrow++
	}
	fmt.Println("could not find row")
	li.Err = errors.New("error: row not found")
	return li
}
