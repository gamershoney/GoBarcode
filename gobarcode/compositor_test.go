package main

import (
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
