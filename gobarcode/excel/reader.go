// Package excel is used to reference and read the columns from the file barcodes are being generated from.
package excel

import "github.com/xuri/excelize/v2"

type LabelInfo struct {
	fname             string
	title             string
	upc               string
	Sheetmap          map[int]string `json:"sheet_map"`
	SelectedSheet     int            `json:"selected_sheet"`
	SelectedSheetName string         `json:"selected_sheet_name"`
	HeaderRow         string         `json:"header_row"`
	HeaderCol         string         `json:"header_col"`
	HeaderRowValues   []string       `json:"header_row_values"`
}

func GetHeaders(name string) (*LabelInfo, error) {
	li := &LabelInfo{fname: name}
	f, err := excelize.OpenFile(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	li.Sheetmap = f.GetSheetMap()
	if len(li.Sheetmap) == 1 {
		li.SelectedSheet = 0
	}
	rows, err := f.Rows(li.Sheetmap[li.SelectedSheet])
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		cols, err := rows.Columns()
		if err != nil {
			return nil, err
		}
		li.HeaderRowValues = cols
	}
	return li, err
}
