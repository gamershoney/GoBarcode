// Package barcode is used to generate the barcode images
package barcode

import (
	"github.com/boombuler/barcode"
	_ "github.com/boombuler/barcode"
	"github.com/boombuler/barcode/twooffive"
)

type Generator struct {
	Width  int
	Height int
}

func NewGenerator(w int, h int) *Generator {
	gen := &Generator{Width: w, Height: h}
	return gen
}

func (gen *Generator) GenerateBarcode(upc string) (barcode.Barcode, error) {
	code, err := twooffive.Encode(upc, true)
	if err != nil {
		return nil, err
	}
	code, err = barcode.Scale(code, gen.Width, gen.Height)
	if err != nil {
		return nil, err
	}

	return code, err
}
