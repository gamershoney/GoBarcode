package main

import (
	"bytes"
	"gobarcode/barcode"
	"gobarcode/excel"
	"image"
	"image/color"
	"image/draw"
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

func pageTestLayout() Layout {
	return Layout{
		ImageHeight: 100,
		ImageWidth:  100,
		PageHeight:  2,
		PageWidth:   2,
		PPI:         100,
	}
}

func solidTestImage(width, height int, fill color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.NewUniform(fill), image.Point{}, draw.Src)
	return img
}

func assertPixelColor(t *testing.T, img image.Image, point image.Point, want color.RGBA) {
	t.Helper()

	got := color.RGBAModel.Convert(img.At(point.X, point.Y)).(color.RGBA)
	if got != want {
		t.Errorf("pixel at %v = %#v, want %#v", point, got, want)
	}
}

func drawPagesWithoutPanic(t *testing.T, layout *Layout, images []*image.RGBA) (pages []*Page, err error) {
	t.Helper()
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("DrawPages() panicked: %v", recovered)
		}
	}()
	return layout.DrawPages(images)
}

func TestCalcPageUsesPixelDimensions(t *testing.T) {
	layout := validTestLayout()

	if got := layout.PagePixelWidth(); got != 1275 {
		t.Errorf("PagePixelWidth() = %d, want 1275", got)
	}
	if got := layout.PagePixelHeight(); got != 1650 {
		t.Errorf("PagePixelHeight() = %d, want 1650", got)
	}

	capacity, columns, rows := layout.CalcPage()

	if capacity != 8 {
		t.Errorf("CalcPage() capacity = %v, want 8", capacity)
	}
	if columns != 2 {
		t.Errorf("CalcPage() columns = %d, want 2", columns)
	}
	if rows != 4 {
		t.Errorf("CalcPage() rows = %d, want 4", rows)
	}
}

func TestBuildPagePlacesImagesInRowMajorOrder(t *testing.T) {
	layout := pageTestLayout()
	red := color.RGBA{R: 255, A: 255}
	green := color.RGBA{G: 255, A: 255}
	blue := color.RGBA{B: 255, A: 255}
	yellow := color.RGBA{R: 255, G: 255, A: 255}
	page := &Page{
		Page:    0,
		Columns: 2,
		Rows:    2,
		Images: []*image.RGBA{
			solidTestImage(100, 100, red),
			solidTestImage(100, 100, green),
			solidTestImage(100, 100, blue),
			solidTestImage(100, 100, yellow),
		},
	}

	layout.BuildPage(page)

	if page.PageImage == nil {
		t.Fatal("BuildPage() did not assign PageImage")
	}
	if got, want := page.PageImage.Bounds(), image.Rect(0, 0, 200, 200); got != want {
		t.Fatalf("BuildPage() bounds = %v, want %v", got, want)
	}

	assertPixelColor(t, page.PageImage, image.Pt(50, 50), red)
	assertPixelColor(t, page.PageImage, image.Pt(150, 50), green)
	assertPixelColor(t, page.PageImage, image.Pt(50, 150), blue)
	assertPixelColor(t, page.PageImage, image.Pt(150, 150), yellow)
}

func TestDrawPagesSplitsImagesWithoutReordering(t *testing.T) {
	layout := pageTestLayout()
	images := []*image.RGBA{
		solidTestImage(100, 100, color.RGBA{R: 1, A: 255}),
		solidTestImage(100, 100, color.RGBA{R: 2, A: 255}),
		solidTestImage(100, 100, color.RGBA{R: 3, A: 255}),
		solidTestImage(100, 100, color.RGBA{R: 4, A: 255}),
		solidTestImage(100, 100, color.RGBA{R: 5, A: 255}),
	}

	pages, err := drawPagesWithoutPanic(t, &layout, images)
	if err != nil {
		t.Fatalf("DrawPages() error = %v", err)
	}
	if len(pages) != 2 {
		t.Fatalf("DrawPages() returned %d pages, want 2", len(pages))
	}

	wantLengths := []int{4, 1}
	nextImage := 0
	for pageIndex, page := range pages {
		if page == nil {
			t.Fatalf("page %d is nil", pageIndex)
		}
		if page.Page != pageIndex {
			t.Errorf("page %d has page number %d", pageIndex, page.Page)
		}
		if len(page.Images) != wantLengths[pageIndex] {
			t.Fatalf("page %d contains %d images, want %d", pageIndex, len(page.Images), wantLengths[pageIndex])
		}
		if page.PageImage == nil {
			t.Errorf("page %d has no composed PageImage", pageIndex)
		}
		for _, img := range page.Images {
			if img != images[nextImage] {
				t.Errorf("page %d image order changed at source index %d", pageIndex, nextImage)
			}
			nextImage++
		}
	}
	if nextImage != len(images) {
		t.Errorf("DrawPages() consumed %d images, want %d", nextImage, len(images))
	}
}

func TestDrawPagesWithNoImagesReturnsNoPages(t *testing.T) {
	layout := pageTestLayout()

	pages, err := drawPagesWithoutPanic(t, &layout, nil)
	if err != nil {
		t.Fatalf("DrawPages(nil) error = %v", err)
	}
	if len(pages) != 0 {
		t.Errorf("DrawPages(nil) returned %d pages, want 0", len(pages))
	}
}

func TestDrawPagesRejectsPageThatCannotFitLabel(t *testing.T) {
	layout := pageTestLayout()
	layout.ImageWidth = layout.PagePixelWidth() + 1
	images := []*image.RGBA{
		solidTestImage(layout.ImageWidth, layout.ImageHeight, color.RGBA{A: 255}),
	}

	pages, err := drawPagesWithoutPanic(t, &layout, images)
	if err == nil || !strings.Contains(err.Error(), "cannot fit a label") {
		t.Fatalf("DrawPages() error = %v, want cannot-fit-label error", err)
	}
	if pages != nil {
		t.Errorf("DrawPages() pages = %#v, want nil", pages)
	}
}

func TestDrawPagesRejectsNilImage(t *testing.T) {
	layout := pageTestLayout()

	pages, err := drawPagesWithoutPanic(t, &layout, []*image.RGBA{nil})
	if err == nil || !strings.Contains(err.Error(), "image 0 is nil") {
		t.Fatalf("DrawPages() error = %v, want nil-image error", err)
	}
	if pages != nil {
		t.Errorf("DrawPages() pages = %#v, want nil", pages)
	}
}
