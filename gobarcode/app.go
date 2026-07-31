package main

import (
	"context"
	"errors"
	"fmt"
	"gobarcode/excel"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	_ "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx          context.Context
	WorkBook     *excel.LabelInfo
	Layout       *Layout
	SaveLocation string
}

// NewApp creates an App instance for the Wails application.
func NewApp() *App {
	return &App{}
}

// startup stores the Wails application context for later runtime calls.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the provided name.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// SelectFile prompts the user to choose a workbook and loads its sheet metadata.
func (a *App) SelectFile() (*excel.LabelInfo, error) {
	fname, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: "Select a Spreadsheet to open..."})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	label, err := excel.GetWorkBookInfo(fname)
	if err != nil {
		return nil, err
	}
	fmt.Println(fname)
	a.WorkBook = label
	return label, err
}

// SetSaveLocation prompts the user for a PDF output path and stores the selection.
func (a *App) SetSaveLocation() (string, error) {
	sname, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: "Barcodes.pdf",
		Title:           "Set the location for your file...",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "PDF Files (*.pdf)",
				Pattern:     "*.pdf",
			},
		},
	})
	if err != nil {
		return "", err
	}

	if sname == "" {
		return "", errors.New("error: no filepath set")
	}
	if strings.EqualFold(filepath.Ext(sname), ".pdf") {
		sname += ".pdf"
	}

	a.SaveLocation = sname
	return sname, err
}

// GetHeaders loads and returns the cell values from the requested header row.
func (a *App) GetHeaders(hr int) ([]string, error) {
	if a.WorkBook == nil {
		return nil, errors.New("error: no workbook set")
	}
	_, err := a.WorkBook.GetHeaderRowValues(hr)
	return a.WorkBook.HeaderRowValues, err
}

// SetColumns selects workbook columns, optional filtering, and missing-UPC behavior.
func (a *App) SetColumns(upc string, title string, filterCol string, filterText string, skipMissingUPC bool, padOddUPC bool) error {
	if a.WorkBook == nil {
		return errors.New("error: no workbook set when setting columns")
	}
	if err := a.WorkBook.SetColumns(upc, title, filterCol, filterText, skipMissingUPC, padOddUPC); err != nil {
		return err
	}
	return a.WorkBook.CreateLabelMap()
}

// SetLayout validates and stores the label and page layout supplied by the frontend.
func (a *App) SetLayout(l Layout) error {
	layout := &l
	if err := layout.ValidateLayout(); err != nil {
		return err
	}
	a.Layout = layout
	return nil
}

// Start generates the labels, composes the pages, and writes the completed PDF.
func (a *App) Start() error {
	if a.WorkBook == nil {
		return errors.New("error: workbook not set")
	}
	if a.Layout == nil {
		return errors.New("error: layout not set")
	}
	if a.SaveLocation == "" {
		return errors.New("error: no save location set")
	}
	labels, err := a.CompositeLabels()
	if err != nil {
		return err
	}
	if len(labels) == 0 {
		return errors.New("error no labels found or generated")
	}
	pages, err := a.Layout.DrawPages(labels)
	if err != nil {
		return err
	}
	pdf, err := a.Layout.CreatePDF(pages)
	if err != nil {
		return err
	}

	err = pdf.OutputFileAndClose(a.SaveLocation)
	if err != nil {
		return err
	}
	return nil
}
