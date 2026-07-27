package excel

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

// createTestWorkbook writes rows to a temporary workbook and returns its path.
func createTestWorkbook(t *testing.T, rows [][]interface{}) string {
	t.Helper()

	file := excelize.NewFile()
	t.Cleanup(func() {
		_ = file.Close()
	})

	for index, row := range rows {
		cell, err := excelize.CoordinatesToCellName(1, index+1)
		if err != nil {
			t.Fatal(err)
		}
		if err := file.SetSheetRow("Sheet1", cell, &row); err != nil {
			t.Fatal(err)
		}
	}

	path := filepath.Join(t.TempDir(), "labels.xlsx")
	if err := file.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestCreateLabelMap verifies that data rows are mapped and incomplete rows are skipped.
func TestCreateLabelMap(t *testing.T) {
	path := createTestWorkbook(t, [][]interface{}{
		{"Label export"},
		{"Title", "UPC"},
		{"First", "00123"},
		{"Second"},
		{"Third", "00456"},
	})
	info := &LabelInfo{
		fname:             path,
		SelectedSheetName: "Sheet1",
		HeaderRow:         2,
		TitleCol:          "Title",
		UPCCol:            "UPC",
	}

	if err := info.CreateLabelMap(); err != nil {
		t.Fatalf("CreateLabelMap() error = %v", err)
	}

	want := TitleUpcMap{
		"First": "00123",
		"Third": "00456",
	}
	if !reflect.DeepEqual(info.TitleUpcMap, want) {
		t.Errorf("TitleUpcMap = %#v, want %#v", info.TitleUpcMap, want)
	}
}

// TestCreateLabelMapRejectsMissingHeader verifies that a missing selected header is rejected.
func TestCreateLabelMapRejectsMissingHeader(t *testing.T) {
	path := createTestWorkbook(t, [][]interface{}{
		{"Title", "Description"},
		{"First", "Example"},
	})
	info := &LabelInfo{
		fname:             path,
		SelectedSheetName: "Sheet1",
		HeaderRow:         1,
		TitleCol:          "Title",
		UPCCol:            "UPC",
	}

	err := info.CreateLabelMap()
	if err == nil || !strings.Contains(err.Error(), "header not found") {
		t.Fatalf("CreateLabelMap() error = %v, want missing-header error", err)
	}
}

// TestGetHeaderRowValuesStoresHeaderRow verifies that the selected row number is retained.
func TestGetHeaderRowValuesStoresHeaderRow(t *testing.T) {
	path := createTestWorkbook(t, [][]interface{}{
		{"Label export"},
		{"Title", "UPC"},
	})
	info := &LabelInfo{
		fname:             path,
		SelectedSheetName: "Sheet1",
	}

	if _, err := info.GetHeaderRowValues(2); err != nil {
		t.Fatalf("GetHeaderRowValues() error = %v", err)
	}
	if info.HeaderRow != 2 {
		t.Errorf("HeaderRow = %d, want 2", info.HeaderRow)
	}
}
