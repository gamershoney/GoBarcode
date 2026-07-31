// Package excel is used to reference and read the columns from the file barcodes are being generated from.
package excel

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// LabelData contains the spreadsheet values needed to generate one label.
// Index records the label's position after incomplete spreadsheet rows are
// removed so concurrent work can be assembled in the original order.
type LabelData struct {
	Index int    `json:"index"`
	Title string `json:"title"`
	UPC   string `json:"upc"`
}

// LabelInfo stores workbook selections and the ordered records used to create labels.
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
	FilterCol         string         `json:"filter_col"`
	FilterText        string         `json:"filter_text"`
	SkipMissingUPC    bool           `json:"skip_missing_upc"`
	PadOddUPC         bool           `json:"pad_odd_upc"`
	HeaderRowValues   []string       `json:"header_row_values"`
	Labels            []LabelData    `json:"labels"`
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

// SetColumns selects the headers used for UPC, title, and optional row filtering.
func (li *LabelInfo) SetColumns(upc string, title string, filterCol string, filterText string, skipMissingUPC bool, padOddUPC bool) error {
	if upc == "" || title == "" {
		return errors.New("error: missing required text")
	}
	li.UPCCol = upc
	li.TitleCol = title
	li.FilterCol = strings.TrimSpace(filterCol)
	li.FilterText = strings.TrimSpace(filterText)
	li.SkipMissingUPC = skipMissingUPC
	li.PadOddUPC = padOddUPC
	return nil
}

// CreateLabelMap loads title and UPC pairs in spreadsheet order.
func (li *LabelInfo) CreateLabelMap() error {
	if li.TitleCol == "" || li.UPCCol == "" {
		return errors.New("error: both column names must be set before loading labels")
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
	filterIndex := -1
	filtering := li.FilterCol != "" && li.FilterText != ""
	labels := make([]LabelData, 0)
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
				if filtering && cell == li.FilterCol {
					filterIndex = i
				}
			}
			if titleIndex < 0 || upcIndex < 0 {
				return errors.New("error: title or upc header not found")
			}
			if filtering && filterIndex < 0 {
				return errors.New("error: filter header not found")
			}
		} else if currentIndex > li.HeaderRow {
			if filtering {
				if filterIndex >= len(row) || !strings.EqualFold(strings.TrimSpace(row[filterIndex]), li.FilterText) {
					currentIndex++
					continue
				}
			}
			if titleIndex >= len(row) {
				currentIndex++
				continue
			}
			if upcIndex >= len(row) || strings.TrimSpace(row[upcIndex]) == "" {
				if li.SkipMissingUPC {
					currentIndex++
					continue
				}
				return fmt.Errorf("error: missing UPC at row %d", currentIndex)
			}
			upc := strings.TrimSpace(row[upcIndex])
			if len(upc)%2 != 0 {
				if !li.PadOddUPC {
					return fmt.Errorf("error: UPC at row %d must contain an even number of digits", currentIndex)
				}
				upc = "0" + upc
			}
			labels = append(labels, LabelData{
				Index: len(labels),
				Title: row[titleIndex],
				UPC:   upc,
			})
		}

		currentIndex++
	}
	if err := rows.Error(); err != nil {
		return err
	}
	if titleIndex < 0 || upcIndex < 0 {
		return errors.New("error: header row not found")
	}

	li.Labels = labels
	return nil
}
