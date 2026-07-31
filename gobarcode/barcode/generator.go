// Package barcode is used to generate the barcode images
package barcode

import (
	"github.com/boombuler/barcode"
	_ "github.com/boombuler/barcode"
	"github.com/boombuler/barcode/twooffive"
)

// Generator configures the pixel dimensions used for generated barcodes.
type Generator struct {
	Width  int
	Height int
}

// NewGenerator creates a barcode generator with the requested pixel dimensions.
func NewGenerator(w int, h int) *Generator {
	gen := &Generator{Width: w, Height: h}
	return gen
}

// GenerateBarcode encodes an Interleaved 2 of 5 value and scales it to the configured size.
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
