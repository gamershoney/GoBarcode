package main

import (
	"context"
	"errors"
	"fmt"
	"gobarcode/excel"

	_ "gobarcode/excel"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	_ "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx      context.Context
	WorkBook *excel.LabelInfo
	Layout   *Layout
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

func (a *App) SetSaveLocation() (string, error) {
	sname, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		DefaultFilename: "Barcodes.pdf",
		Title:           "Set the location for your file...",
	})
	if err != nil {
		return "", err
	}

	if sname == "" {
		return "", errors.New("error: no filepath set")
	}
	return sname, err
}

// GetHeaders loads and returns the cell values from the requested header row.
func (a *App) GetHeaders(hr int) ([]string, error) {
	_, err := a.WorkBook.GetHeaderRowValues(hr)
	return a.WorkBook.HeaderRowValues, err
}

// SetColumns selects the workbook columns containing UPC and title values.
func (a *App) SetColumns(upc string, title string) error {
	if err := a.WorkBook.SetColumns(upc, title); err != nil {
		return err
	}
	return a.WorkBook.CreateLabelMap()
}

func (a *App) SetLayout(l Layout) error {
	layout := &l
	if err := layout.ValidateLayout(); err != nil {
		return err
	}
	a.Layout = layout
	return nil
}
