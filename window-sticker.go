package clienthandlers

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/jung-kurt/gofpdf"
	"github.com/jung-kurt/gofpdf/contrib/gofpdi"
	"github.com/tsawler/goblender/client/clienthandlers/clientmodels"
)

// CreateWindowSticker creates the window sticker as a PDF
func CreateWindowSticker(v clientmodels.Vehicle) (*gofpdf.Fpdf, error) {
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(10, 13, 10)
	importer := gofpdi.NewImporter()
	var t int

	pdf.AddUTF8Font("CenturyGothic-Bold", "", "./client/clienthandlers/fonts/gothicb.ttf")

	if v.HandPicked == 0 {
		// get template to write on
		t = importer.ImportPage(pdf, "./client/clienthandlers/pdf-templates/window-sticker-oct-2019.pdf", 1, "/MediaBox")
		pdf.AddPage()
		importer.UseImportedTemplate(pdf, t, 0, 0, 215.9, 0)

		// write make/model/year/trim
		pdf.SetFont("Arial", "BI", 24)
		pdf.Write(0, fmt.Sprintf("%d %s %s %s", v.Year, v.Make.Make, v.Model.Model, v.Trim))
		pdf.SetX(162)
		pdf.SetFont("Arial", "BIS", 28)
		pdf.Write(0, fmt.Sprintf(" $%d  ", int(v.TotalMSR)))

		// write odometer
		pdf.SetY(24)
		pdf.SetFont("Arial", "B", 20)
		pdf.Write(0, fmt.Sprintf("%s km", humanize.Comma(int64(v.Odometer))))

		// write pricing details
		pdf.SetFont("CenturyGothic-Bold", "", 16)
		if v.PriceForDisplay == "" {
			pdf.SetY(22)
			pdf.SetFont("Arial", "B", 16)
			pdf.MultiCell(193, 3, fmt.Sprintf("$%s", humanize.Comma(int64(v.Cost))), "", "R", false)
		} else {
			pdf.SetY(22)
			pdf.SetFont("Arial", "B", 16)
			pdf.MultiCell(193, 3, fmt.Sprintf("%s OFF NEW MSRP = $%s", v.PriceForDisplay, humanize.Comma(int64(v.Cost))), "", "R", false)
		}

		// write options

		// write Stock #
		// TODO actually create the sticker!
	} else {
		// mvi select
		t = importer.ImportPage(pdf, "./client/clienthandlers/pdf-templates/mv-plus-select.pdf", 1, "/MediaBox")
	}

	return pdf, nil
}
