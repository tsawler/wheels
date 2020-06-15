package clienthandlers

import (
	"fmt"
	"github.com/jung-kurt/gofpdf"
)

func CreateWindowSticker(id int) (*gofpdf.Fpdf, error) {
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, fmt.Sprintf("Hello, world, id is %d", id))

	return pdf, nil
}
