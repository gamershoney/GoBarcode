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

// TestCreateLabelMap verifies that data rows remain ordered and incomplete rows are skipped.
func TestCreateLabelMap(t *testing.T) {
	path := createTestWorkbook(t, [][]interface{}{
		{"Label export"},
		{"Title", "UPC"},
		{"First", "00123"},
		{"Second"},
		{"Third", "00456"},
		{"First", "00789"},
	})
	info := &LabelInfo{
		fname:             path,
		SelectedSheetName: "Sheet1",
		HeaderRow:         2,
		TitleCol:          "Title",
		UPCCol:            "UPC",
		SkipMissingUPC:    true,
		PadOddUPC:         true,
	}

	if err := info.CreateLabelMap(); err != nil {
		t.Fatalf("CreateLabelMap() error = %v", err)
	}

	want := []LabelData{
		{Index: 0, Title: "First", UPC: "000123"},
		{Index: 1, Title: "Third", UPC: "000456"},
		{Index: 2, Title: "First", UPC: "000789"},
	}
	if !reflect.DeepEqual(info.Labels, want) {
		t.Errorf("Labels = %#v, want %#v", info.Labels, want)
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

// TestCreateLabelMapFiltersRows verifies optional filtering is exact, trimmed, and case-insensitive.
func TestCreateLabelMapFiltersRows(t *testing.T) {
	path := createTestWorkbook(t, [][]interface{}{
		{"Title", "UPC", "Status"},
		{"First", "001234", "Active"},
		{"Second", "004567", "Inactive"},
		{"Third", "007890", " active "},
		{"Fourth", "001122"},
	})
	info := &LabelInfo{
		fname:             path,
		SelectedSheetName: "Sheet1",
		HeaderRow:         1,
	}
	if err := info.SetColumns("UPC", "Title", "Status", " ACTIVE ", false, true); err != nil {
		t.Fatalf("SetColumns() error = %v", err)
	}

	if err := info.CreateLabelMap(); err != nil {
		t.Fatalf("CreateLabelMap() error = %v", err)
	}

	want := []LabelData{
		{Index: 0, Title: "First", UPC: "001234"},
		{Index: 1, Title: "Third", UPC: "007890"},
	}
	if !reflect.DeepEqual(info.Labels, want) {
		t.Errorf("Labels = %#v, want %#v", info.Labels, want)
	}
}

// TestCreateLabelMapRejectsMissingFilterHeader verifies an enabled filter requires its header.
func TestCreateLabelMapRejectsMissingFilterHeader(t *testing.T) {
	path := createTestWorkbook(t, [][]interface{}{
		{"Title", "UPC"},
		{"First", "001234"},
	})
	info := &LabelInfo{
		fname:             path,
		SelectedSheetName: "Sheet1",
		HeaderRow:         1,
	}
	if err := info.SetColumns("UPC", "Title", "Status", "Active", false, true); err != nil {
		t.Fatalf("SetColumns() error = %v", err)
	}

	err := info.CreateLabelMap()
	if err == nil || !strings.Contains(err.Error(), "filter header not found") {
		t.Fatalf("CreateLabelMap() error = %v, want missing-filter-header error", err)
	}
}

// TestCreateLabelMapIgnoresIncompleteFilter verifies either empty filter setting disables filtering.
func TestCreateLabelMapIgnoresIncompleteFilter(t *testing.T) {
	path := createTestWorkbook(t, [][]interface{}{
		{"Title", "UPC", "Status"},
		{"First", "001234", "Active"},
		{"Second", "004567", "Inactive"},
	})
	info := &LabelInfo{
		fname:             path,
		SelectedSheetName: "Sheet1",
		HeaderRow:         1,
	}
	if err := info.SetColumns("UPC", "Title", "Status", "", false, true); err != nil {
		t.Fatalf("SetColumns() error = %v", err)
	}

	if err := info.CreateLabelMap(); err != nil {
		t.Fatalf("CreateLabelMap() error = %v", err)
	}
	if len(info.Labels) != 2 {
		t.Errorf("len(Labels) = %d, want 2", len(info.Labels))
	}
}

// TestCreateLabelMapFailsOnMissingUPC verifies strict mode reports the 1-based spreadsheet row.
func TestCreateLabelMapFailsOnMissingUPC(t *testing.T) {
	path := createTestWorkbook(t, [][]interface{}{
		{"Title", "UPC"},
		{"First", "001234"},
		{"Missing"},
	})
	info := &LabelInfo{
		fname:             path,
		SelectedSheetName: "Sheet1",
		HeaderRow:         1,
	}
	if err := info.SetColumns("UPC", "Title", "", "", false, true); err != nil {
		t.Fatalf("SetColumns() error = %v", err)
	}

	err := info.CreateLabelMap()
	if err == nil || !strings.Contains(err.Error(), "missing UPC at row 3") {
		t.Fatalf("CreateLabelMap() error = %v, want missing UPC at row 3", err)
	}
}

// TestCreateLabelMapSkipsMissingUPC verifies permissive mode omits blank UPC rows and reindexes results.
func TestCreateLabelMapSkipsMissingUPC(t *testing.T) {
	path := createTestWorkbook(t, [][]interface{}{
		{"Title", "UPC"},
		{"First", "001234"},
		{"Missing"},
		{"Blank", "   "},
		{"Second", "004567"},
	})
	info := &LabelInfo{
		fname:             path,
		SelectedSheetName: "Sheet1",
		HeaderRow:         1,
	}
	if err := info.SetColumns("UPC", "Title", "", "", true, true); err != nil {
		t.Fatalf("SetColumns() error = %v", err)
	}

	if err := info.CreateLabelMap(); err != nil {
		t.Fatalf("CreateLabelMap() error = %v", err)
	}
	want := []LabelData{
		{Index: 0, Title: "First", UPC: "001234"},
		{Index: 1, Title: "Second", UPC: "004567"},
	}
	if !reflect.DeepEqual(info.Labels, want) {
		t.Errorf("Labels = %#v, want %#v", info.Labels, want)
	}
}

// TestCreateLabelMapPadsOddUPC verifies enabled padding prepends exactly one zero.
func TestCreateLabelMapPadsOddUPC(t *testing.T) {
	path := createTestWorkbook(t, [][]interface{}{
		{"Title", "UPC"},
		{"Odd", "12345"},
		{"Even", "123456"},
	})
	info := &LabelInfo{
		fname:             path,
		SelectedSheetName: "Sheet1",
		HeaderRow:         1,
	}
	if err := info.SetColumns("UPC", "Title", "", "", false, true); err != nil {
		t.Fatalf("SetColumns() error = %v", err)
	}

	if err := info.CreateLabelMap(); err != nil {
		t.Fatalf("CreateLabelMap() error = %v", err)
	}
	want := []LabelData{
		{Index: 0, Title: "Odd", UPC: "012345"},
		{Index: 1, Title: "Even", UPC: "123456"},
	}
	if !reflect.DeepEqual(info.Labels, want) {
		t.Errorf("Labels = %#v, want %#v", info.Labels, want)
	}
}

// TestCreateLabelMapRejectsOddUPC verifies disabled padding reports the spreadsheet row.
func TestCreateLabelMapRejectsOddUPC(t *testing.T) {
	path := createTestWorkbook(t, [][]interface{}{
		{"Title", "UPC"},
		{"Odd", "12345"},
	})
	info := &LabelInfo{
		fname:             path,
		SelectedSheetName: "Sheet1",
		HeaderRow:         1,
	}
	if err := info.SetColumns("UPC", "Title", "", "", false, false); err != nil {
		t.Fatalf("SetColumns() error = %v", err)
	}

	err := info.CreateLabelMap()
	if err == nil || !strings.Contains(err.Error(), "UPC at row 2 must contain an even number of digits") {
		t.Fatalf("CreateLabelMap() error = %v, want odd-length UPC error at row 2", err)
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
