package main

import (
	"fmt"
	"gobarcode/barcode"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type Placement struct {
	Height  int `json:"height"`
	Width   int `json:"width"`
	OriginX int `json:"origin_x"`
	OriginY int `json:"origin_y"`
}

func (p Placement) Area() int {
	return (p.Height * p.Width)
}

type Layout struct {
	ImageHeight      int       `json:"image_height"`
	ImageWidth       int       `json:"image_width"`
	BarcodePlacement Placement `json:"barcode_placement"`
	TitlePlacement   Placement `json:"title_placement"`
	PageHeight       float32   `json:"page_height"`
	PageWidth        float32   `json:"page_width"`
	PPI              int       `json:"ppi"`
}

func SizeError(prop string) error {
	err := fmt.Errorf("error: %s cannot be 0 or less", prop)
	return err
}

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

	pagePixelWidth := float64(l.PageWidth) * float64(l.PPI)
	pagePixelHeight := float64(l.PageHeight) * float64(l.PPI)
	if float64(l.ImageWidth) > pagePixelWidth {
		return fmt.Errorf("error: image width %d exceeds page width %.0f pixels", l.ImageWidth, pagePixelWidth)
	}
	if float64(l.ImageHeight) > pagePixelHeight {
		return fmt.Errorf("error: image height %d exceeds page height %.0f pixels", l.ImageHeight, pagePixelHeight)
	}

	return nil
}

func validPositiveFloat(value float32) bool {
	number := float64(value)
	return number > 0 && !math.IsNaN(number) && !math.IsInf(number, 0)
}

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

func (l *Layout) CalcPage() (a float64, cols int, rows int) {
	imageArea := l.ImageHeight * l.ImageWidth
	pageArea := l.PageHeight * l.PageWidth
	fit := pageArea / float32(imageArea)
	count := math.Floor(float64(fit))
	a = count

	return
}

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

type Page struct {
	Page    int `json:"page"`
	Columns int
	Rows    int
	Images  []*image.RGBA
	Error   error `json:"error"`
}

func (l *Layout) BuildPage(page *Page) {
	canvas := image.NewRGBA(image.Rect(0, 0, int(l.PageWidth)*l.PPI, l.PageHeight*float32(l.PPI)))
	rect := image.Rect(0, 0, l.ImageWidth, l.ImageHeight)
	for _, image := range page.Images {
		// draw.Draw(canvas,rect, image, sp image.Point, op draw.Op)
	}
}

func (l *Layout) DrawPages(imgs []*image.RGBA) error {
	count := l.CalcPage()

	return nil
}
