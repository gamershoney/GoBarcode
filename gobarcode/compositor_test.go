package main

import (
	"bytes"
	"gobarcode/barcode"
	"gobarcode/excel"
	"strings"
	"testing"
)

func validTestLayout() Layout {
	return Layout{
		ImageHeight: 360,
		ImageWidth:  600,
		BarcodePlacement: Placement{
			Height:  100,
			Width:   280,
			OriginX: 160,
			OriginY: 190,
		},
		TitlePlacement: Placement{
			Height:  54,
			Width:   360,
			OriginX: 120,
			OriginY: 92,
		},
		PageHeight: 11,
		PageWidth:  8.5,
		PPI:        150,
	}
}

func TestValidateLayout(t *testing.T) {
	tests := []struct {
		name      string
		change    func(*Layout)
		wantError string
	}{
		{
			name:   "valid layout",
			change: func(*Layout) {},
		},
		{
			name: "missing image width",
			change: func(layout *Layout) {
				layout.ImageWidth = 0
			},
			wantError: "image_width",
		},
		{
			name: "missing page height",
			change: func(layout *Layout) {
				layout.PageHeight = 0
			},
			wantError: "page_height",
		},
		{
			name: "missing ppi",
			change: func(layout *Layout) {
				layout.PPI = 0
			},
			wantError: "ppi",
		},
		{
			name: "negative barcode origin",
			change: func(layout *Layout) {
				layout.BarcodePlacement.OriginX = -1
			},
			wantError: "barcode origin_x",
		},
		{
			name: "barcode exceeds right edge",
			change: func(layout *Layout) {
				layout.BarcodePlacement.OriginX = 400
			},
			wantError: "barcode placement exceeds image width",
		},
		{
			name: "title exceeds bottom edge",
			change: func(layout *Layout) {
				layout.TitlePlacement.OriginY = 320
			},
			wantError: "title placement exceeds image height",
		},
		{
			name: "title is shorter than font",
			change: func(layout *Layout) {
				layout.TitlePlacement.Height = 12
			},
			wantError: "title height must be at least 13 pixels",
		},
		{
			name: "image is wider than page",
			change: func(layout *Layout) {
				layout.PageWidth = 3
			},
			wantError: "image width",
		},
		{
			name: "image is taller than page",
			change: func(layout *Layout) {
				layout.PageHeight = 2
			},
			wantError: "image height",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			layout := validTestLayout()
			test.change(&layout)

			err := layout.ValidateLayout()
			if test.wantError == "" {
				if err != nil {
					t.Fatalf("ValidateLayout() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("ValidateLayout() error = nil, want error containing %q", test.wantError)
			}
			if !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("ValidateLayout() error = %q, want error containing %q", err, test.wantError)
			}
		})
	}
}

func TestCompositeLabelsPreservesInputOrder(t *testing.T) {
	layout := validTestLayout()
	labels := []excel.LabelData{
		{Index: 0, Title: "First", UPC: "001234"},
		{Index: 1, Title: "Second", UPC: "004567"},
		{Index: 2, Title: "First", UPC: "007890"},
	}
	app := &App{
		WorkBook: &excel.LabelInfo{Labels: labels},
		Layout:   &layout,
	}

	images, err := app.CompositeLabels()
	if err != nil {
		t.Fatalf("CompositeLabels() error = %v", err)
	}
	if len(images) != len(labels) {
		t.Fatalf("CompositeLabels() returned %d images, want %d", len(images), len(labels))
	}

	generator := barcode.NewGenerator(
		layout.BarcodePlacement.Width,
		layout.BarcodePlacement.Height,
	)
	for index, label := range labels {
		barcodeImage, err := generator.GenerateBarcode(label.UPC)
		if err != nil {
			t.Fatalf("GenerateBarcode(%q) error = %v", label.UPC, err)
		}
		want := layout.DrawLabel(barcodeImage)
		layout.DrawTitle(want, label.Title)

		if images[index] == nil {
			t.Fatalf("image at index %d is nil", index)
		}
		if !bytes.Equal(images[index].Pix, want.Pix) {
			t.Errorf("image at index %d does not match label %q", index, label.Title)
		}
	}
}
