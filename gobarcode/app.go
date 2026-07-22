package main

import (
	"context"
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
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) SelectFile() *excel.LabelInfo {
	fname, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: "Select a Spreadsheet to open..."})
	if err != nil {
		fmt.Println(err)
		return nil
	}

	label := excel.GetWorkBookInfo(fname)
	fmt.Println(fname)
	a.WorkBook = label
	return label
}

func (a *App) GetHeaders(hr int) ([]string, error) {
	a.WorkBook.GetHeaderRowValues(hr)
	return a.WorkBook.HeaderRowValues, a.WorkBook.Err
}
