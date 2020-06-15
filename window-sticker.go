package clienthandlers

import (
	"fmt"
	"github.com/jung-kurt/gofpdf"
	"github.com/tsawler/goblender/client/clienthandlers/clientmodels"
)

// CreateWindowSticker creates the window sticker as a PDF
func CreateWindowSticker(v clientmodels.Vehicle) (*gofpdf.Fpdf, error) {
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, fmt.Sprintf("Hello, world, vehicle is %d %s %s %s", v.Year, v.Make.Make, v.Model.Model, v.Trim))

	// TODO actually create the sticker!

	return pdf, nil
}
