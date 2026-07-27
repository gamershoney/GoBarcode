// Package excel is used to reference and read the columns from the file barcodes are being generated from.
package excel

import (
	"errors"
	"fmt"

	"github.com/xuri/excelize/v2"
)

// TitleUpcMap is a map with a key for the Title and the Upc for the value
type TitleUpcMap map[string]string

type LabelInfo struct {
	fname             string
	title             string
	upc               string
	Sheetmap          map[int]string `json:"sheet_map"`
	SelectedSheet     int            `json:"selected_sheet"`
	SelectedSheetName string         `json:"selected_sheet_name"`
	HeaderRow         int            `json:"header_row"`
	TitleCol          string         `json:"header_col"`
	UPCCol            string         `json:"upc_col"`
	HeaderRowValues   []string       `json:"header_row_values"`
	TitleUpcMap       TitleUpcMap
}

// GetWorkBookInfo opens a workbook and returns metadata for its initially selected sheet.
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

// GetHeaderRowValues loads the cell values from row hr on the selected sheet.
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
	defer file.Close()
	rows, err := file.Rows(li.SelectedSheetName)
	if err != nil {
		return nil, err
	}
	fmt.Println("reading rows")
	defer rows.Close()
	currentrow := 1
	for rows.Next() {
		if currentrow == hr {
			cols, err := rows.Columns()
			if err != nil {
				return nil, err
			}
			li.HeaderRow = hr
			li.HeaderRowValues = cols
			return li, nil
		}
		currentrow++
	}
	fmt.Println("could not find row")
	err = errors.New("error: row not found")
	return li, err
}

// SetColumns selects the header names used for UPC and title values.
func (li *LabelInfo) SetColumns(upc string, title string) error {
	if upc == "" || title == "" {
		return errors.New("error: missing required text")
	}
	li.UPCCol = upc
	li.TitleCol = title
	return nil
}

// CreateLabelMap maps each title to its UPC using the selected sheet and columns.
func (li *LabelInfo) CreateLabelMap() error {
	if li.TitleCol == "" || li.UPCCol == "" {
		return errors.New("error: both column names must be set before creating map")
	}
	if li.HeaderRow < 1 {
		return errors.New("error: header row must be set before creating map")
	}
	if li.SelectedSheetName == "" {
		return errors.New("error: missing sheet name")
	}

	file, err := excelize.OpenFile(li.fname)
	if err != nil {
		return err
	}
	defer file.Close()

	rows, err := file.Rows(li.SelectedSheetName)
	if err != nil {
		return err
	}
	defer rows.Close()

	titleIndex := -1
	upcIndex := -1
	labelmap := make(map[string]string)
	currentIndex := 1
	for rows.Next() {
		row, err := rows.Columns()
		if err != nil {
			return err
		}

		if currentIndex == li.HeaderRow {
			for i, cell := range row {
				if cell == li.TitleCol {
					titleIndex = i
				}
				if cell == li.UPCCol {
					upcIndex = i
				}
			}
			if titleIndex < 0 || upcIndex < 0 {
				return errors.New("error: title or upc header not found")
			}
		} else if currentIndex > li.HeaderRow {
			if titleIndex >= len(row) || upcIndex >= len(row) {
				currentIndex++
				continue
			}
			labelmap[row[titleIndex]] = row[upcIndex]
		}

		currentIndex++
	}
	if err := rows.Error(); err != nil {
		return err
	}
	if titleIndex < 0 || upcIndex < 0 {
		return errors.New("error: header row not found")
	}

	li.TitleUpcMap = labelmap
	return nil
}
