package main

import (
	"bytes"
	"errors"
	"fmt"
	"gobarcode/barcode"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"strings"
	"sync"

	"codeberg.org/go-pdf/fpdf"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Placement describes the pixel dimensions and origin of an element on a label.
type Placement struct {
	Height  int `json:"height"`
	Width   int `json:"width"`
	OriginX int `json:"origin_x"`
	OriginY int `json:"origin_y"`
}

// Area returns the placement area in square pixels.
func (p Placement) Area() int {
	return (p.Height * p.Width)
}

// Layout defines the label canvas, element placements, and physical page settings.
type Layout struct {
	ImageHeight      int       `json:"image_height"`
	ImageWidth       int       `json:"image_width"`
	BarcodePlacement Placement `json:"barcode_placement"`
	TitlePlacement   Placement `json:"title_placement"`
	PageHeight       float32   `json:"page_height"`
	PageWidth        float32   `json:"page_width"`
	PPI              int       `json:"ppi"`
}

// SizeError returns a validation error for a non-positive dimension.
func SizeError(prop string) error {
	err := fmt.Errorf("error: %s cannot be 0 or less", prop)
	return err
}

// ValidateLayout verifies label dimensions, page settings, and element boundaries.
func (l *Layout) ValidateLayout() error {
	if l.ImageHeight <= 0 {
		return SizeError("image_height")
	}
	if l.ImageWidth <= 0 {
		return SizeError("image_width")
	}
	if !validPositiveFloat(l.PageHeight) {
		return SizeError("page_height")
	}
	if !validPositiveFloat(l.PageWidth) {
		return SizeError("page_width")
	}
	if l.PPI <= 0 {
		return SizeError("ppi")
	}

	if err := l.validatePlacement("barcode", l.BarcodePlacement); err != nil {
		return err
	}
	if err := l.validatePlacement("title", l.TitlePlacement); err != nil {
		return err
	}
	if l.TitlePlacement.Height < basicfont.Face7x13.Height {
		return fmt.Errorf(
			"error: title height must be at least %d pixels for the configured font",
			basicfont.Face7x13.Height,
		)
	}

	pagePixelWidth := l.PagePixelWidth()
	pagePixelHeight := l.PagePixelHeight()
	if l.ImageWidth > pagePixelWidth {
		return fmt.Errorf("error: image width %d exceeds page width %d pixels", l.ImageWidth, pagePixelWidth)
	}
	if l.ImageHeight > pagePixelHeight {
		return fmt.Errorf("error: image height %d exceeds page height %d pixels", l.ImageHeight, pagePixelHeight)
	}

	return nil
}

// validPositiveFloat reports whether value is finite and greater than zero.
func validPositiveFloat(value float32) bool {
	number := float64(value)
	return number > 0 && !math.IsNaN(number) && !math.IsInf(number, 0)
}

// validatePlacement verifies that an element has positive dimensions and fits on the label.
func (l *Layout) validatePlacement(name string, placement Placement) error {
	if placement.Height <= 0 {
		return fmt.Errorf("error: %s height cannot be 0 or less", name)
	}
	if placement.Width <= 0 {
		return fmt.Errorf("error: %s width cannot be 0 or less", name)
	}
	if placement.OriginX < 0 {
		return fmt.Errorf("error: %s origin_x cannot be negative", name)
	}
	if placement.OriginY < 0 {
		return fmt.Errorf("error: %s origin_y cannot be negative", name)
	}

	// Subtraction avoids overflowing if origin and size are both very large.
	if placement.Width > l.ImageWidth || placement.OriginX > l.ImageWidth-placement.Width {
		return fmt.Errorf("error: %s placement exceeds image width", name)
	}
	if placement.Height > l.ImageHeight || placement.OriginY > l.ImageHeight-placement.Height {
		return fmt.Errorf("error: %s placement exceeds image height", name)
	}

	return nil
}

// DrawLabel creates a white label canvas and draws the barcode in its configured placement.
func (l *Layout) DrawLabel(src image.Image) *image.RGBA {
	labelBounds := image.Rect(0, 0, l.ImageWidth, l.ImageHeight)
	dst := image.NewRGBA(labelBounds)
	draw.Draw(dst, labelBounds, image.NewUniform(color.White), image.Point{}, draw.Src)
	p := l.BarcodePlacement
	barcodeBounds := image.Rect(p.OriginX, p.OriginY, p.OriginX+p.Width, p.OriginY+p.Height)
	draw.Draw(
		dst,
		barcodeBounds,
		src,
		src.Bounds().Min,
		draw.Over,
	)
	return dst
}

// DrawTitle centers a title within its configured placement on a label image.
func (l *Layout) DrawTitle(img *image.RGBA, text string) {
	placement := l.TitlePlacement
	face := basicfont.Face7x13
	text = truncateText(face, text, placement.Width)

	textWidth := font.MeasureString(face, text).Ceil()
	metrics := face.Metrics()
	textHeight := metrics.Height.Ceil()
	originX := placement.OriginX + (placement.Width-textWidth)/2
	baselineY := placement.OriginY + (placement.Height-textHeight)/2 + metrics.Ascent.Ceil()

	titleBounds := image.Rect(
		placement.OriginX,
		placement.OriginY,
		placement.OriginX+placement.Width,
		placement.OriginY+placement.Height,
	)
	titleImage := img.SubImage(titleBounds).(*image.RGBA)
	drwr := &font.Drawer{
		Dst:  titleImage,
		Src:  image.NewUniform(color.Black),
		Face: face,
		Dot:  fixed.P(originX, baselineY),
	}
	drwr.DrawString(text)
}

// truncateText returns the longest prefix of text that fits within maxWidth pixels.
func truncateText(face font.Face, text string, maxWidth int) string {
	if font.MeasureString(face, text).Ceil() <= maxWidth {
		return text
	}

	var result strings.Builder
	for _, character := range text {
		candidate := result.String() + string(character)
		if font.MeasureString(face, candidate).Ceil() > maxWidth {
			break
		}
		result.WriteRune(character)
	}
	return result.String()
}

// PagePixelWidth converts the configured page width from inches to pixels.
func (l *Layout) PagePixelWidth() int {
	return int(math.Round(float64(l.PageWidth) * float64(l.PPI)))
}

// PagePixelHeight converts the configured page height from inches to pixels.
func (l *Layout) PagePixelHeight() int {
	return int(math.Round(float64(l.PageHeight) * float64(l.PPI)))
}

// CalcPage returns the label capacity, columns, and rows available on one page.
func (l *Layout) CalcPage() (capacity int, cols int, rows int) {
	if l.ImageWidth <= 0 || l.ImageHeight <= 0 {
		return 0, 0, 0
	}

	cols = l.PagePixelWidth() / l.ImageWidth
	rows = l.PagePixelHeight() / l.ImageHeight
	capacity = cols * rows
	return capacity, cols, rows
}

// CompositeLabels concurrently renders labels while preserving spreadsheet order.
func (a *App) CompositeLabels() ([]*image.RGBA, error) {
	type labelResult struct {
		index int
		image *image.RGBA
		err   error
	}

	labels := a.WorkBook.Labels
	resultChan := make(chan labelResult, len(labels))
	var wg sync.WaitGroup

	for _, label := range labels {
		wg.Go(func() {
			gen := barcode.NewGenerator(a.Layout.BarcodePlacement.Width, a.Layout.BarcodePlacement.Height)
			barcodeImage, err := gen.GenerateBarcode(label.UPC)
			if err != nil {
				resultChan <- labelResult{index: label.Index, err: err}
				return
			}

			labelImage := a.Layout.DrawLabel(barcodeImage)
			a.Layout.DrawTitle(labelImage, label.Title)
			resultChan <- labelResult{index: label.Index, image: labelImage}
		})
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	images := make([]*image.RGBA, len(labels))
	var firstErr error
	for result := range resultChan {
		if result.err != nil {
			if firstErr == nil {
				firstErr = result.err
			}
			continue
		}
		images[result.index] = result.image
	}

	if firstErr != nil {
		return nil, firstErr
	}
	return images, nil
}

// Page groups label images with their composed page canvas and placement metadata.
type Page struct {
	Page      int `json:"page"`
	Columns   int
	Rows      int
	Images    []*image.RGBA
	PageImage *image.RGBA
	Error     error `json:"error"`
}

// BuildPage draws a page's labels in row-major order onto a page-sized canvas.
func (l *Layout) BuildPage(page *Page) {
	if page == nil {
		return
	}
	if page.Columns <= 0 || page.Rows <= 0 {
		page.Error = errors.New("error: page must have at least one row and column")
		return
	}

	pageBounds := image.Rect(0, 0, l.PagePixelWidth(), l.PagePixelHeight())
	canvas := image.NewRGBA(pageBounds)
	draw.Draw(canvas, pageBounds, image.NewUniform(color.White), image.Point{}, draw.Src)

	capacity := page.Columns * page.Rows
	for index, img := range page.Images {
		if index >= capacity {
			page.Error = fmt.Errorf("error: page %d contains more images than its capacity of %d", page.Page, capacity)
			break
		}
		if img == nil {
			page.Error = fmt.Errorf("error: page %d contains a nil image at index %d", page.Page, index)
			break
		}

		currentCol := index % page.Columns
		currentRow := index / page.Columns
		rect := image.Rect(
			currentCol*l.ImageWidth,
			currentRow*l.ImageHeight,
			(currentCol+1)*l.ImageWidth,
			(currentRow+1)*l.ImageHeight,
		)
		draw.Draw(canvas, rect, img, img.Bounds().Min, draw.Over)
	}
	page.PageImage = canvas
}

// DrawPages divides ordered label images into pages and composes their canvases concurrently.
func (l *Layout) DrawPages(imgs []*image.RGBA) ([]*Page, error) {
	if len(imgs) == 0 {
		return []*Page{}, nil
	}

	capacity, cols, rows := l.CalcPage()
	if capacity <= 0 {
		return nil, errors.New("error: page dimensions cannot fit a label")
	}
	for index, img := range imgs {
		if img == nil {
			return nil, fmt.Errorf("error: image %d is nil", index)
		}
	}

	pageCount := (len(imgs) + capacity - 1) / capacity
	var wg sync.WaitGroup
	pGroup := make([]*Page, pageCount)
	for index := range pageCount {
		starting := index * capacity
		final := min(starting+capacity, len(imgs))
		page := &Page{
			Page:    index,
			Columns: cols,
			Rows:    rows,
			Images:  imgs[starting:final],
		}
		pGroup[index] = page

		wg.Go(func() {
			l.BuildPage(page)
		})
	}
	wg.Wait()

	for _, page := range pGroup {
		if page.Error != nil {
			return pGroup, page.Error
		}
	}

	return pGroup, nil
}

// EncodeImage encodes a composed page canvas as PNG data for PDF embedding.
func EncodeImage(p *Page) (*bytes.Reader, error) {
	if p == nil {
		return nil, errors.New("error: page is nil")
	}
	if p.PageImage == nil {
		return nil, errors.New("error: page image is nil")
	}

	var b bytes.Buffer
	err := png.Encode(&b, p.PageImage)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b.Bytes()), err
}

// CreatePDF embeds each composed page image into a physical-size PDF page.
// It consumes successfully embedded page images to release their raster memory.
func (l *Layout) CreatePDF(pgs []*Page) (*fpdf.Fpdf, error) {
	if len(pgs) == 0 {
		return nil, errors.New("error: no pages to add to PDF")
	}

	opts := &fpdf.InitType{
		Size: fpdf.SizeType{
			Wd: float64(l.PageWidth),
			Ht: float64(l.PageHeight),
		},
		UnitStr:        fpdf.UnitInch,
		OrientationStr: fpdf.OrientationPortrait,
	}
	pdf := fpdf.NewCustom(opts)
	for i, p := range pgs {
		name := fmt.Sprintf("image_%d", i)
		img, err := EncodeImage(p)
		if err != nil {
			return nil, fmt.Errorf("encode PDF page %d: %w", i+1, err)
		}
		pdf.AddPage()
		pdf.RegisterImageOptionsReader(
			name,
			fpdf.ImageOptions{
				ImageType: "PNG",
			},
			img,
		)

		pdf.ImageOptions(name,
			0,
			0,
			opts.Size.Wd,
			opts.Size.Ht,
			false,
			fpdf.ImageOptions{
				ImageType: "PNG",
			},
			0,
			"")

		if err := pdf.Error(); err != nil {
			return nil, fmt.Errorf("add PDF page %d: %w", i+1, err)
		}
		p.PageImage = nil
		p.Images = nil
		pgs[i] = nil

	}

	return pdf, nil
}
