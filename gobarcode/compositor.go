package main

import (
	"fmt"
	"gobarcode/barcode"
	"image"
	"image/color"
	"image/draw"
	"math"
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

func (a *App) CompositeLabels() error {
	umap := a.WorkBook.TitleUpcMap
	gen := barcode.NewGenerator(a.Layout.BarcodePlacement.Width, a.Layout.BarcodePlacement.Height)

	for _, v := range umap {
		barcode, err := gen.GenerateBarcode(v)
		if err != nil {
			return err
		}
		labelImage := a.Layout.DrawLabel(barcode)
		_ = labelImage
	}
	return nil
}
