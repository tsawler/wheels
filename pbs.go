package clienthandlers

import (
	"fmt"
	"net/http"
	"time"
)

var defaultOptions []int = []int{
	33,
	34,
	35,
	36,
	37,
	45,
	62,
	63,
	65,
	76,
	80,
	81,
}

type ExteriorColor struct {
	Description string `json:"Description"`
}

type InteriorColor struct {
	Description string `json:"Description"`
}

type PBSVehicle struct {
	ID            string        `json:"Id"`
	VehicleID     string        `json:"VehicleId"`
	SerialNumber  string        `json:"SerialNumber"`
	StockNumber   string        `json:"StockNumber"`
	VIN           string        `json:"VIN"`
	Status        string        `json:"Status"`
	OwnerRef      string        `json:"OwnerRef"`
	Make          string        `json:"Make"`
	Model         string        `json:"Model"`
	Trim          string        `json:"Trim"`
	VehicleType   string        `json:"VehicleType"`
	Year          string        `json:"Year"`
	Odometer      int           `json:"Odometer"`
	ExteriorColor ExteriorColor `json:"ExteriorColor"`
	InteriorColor InteriorColor `json:"InteriorColor"`
	Engine        string        `json:"Engine"`
	Cylinders     string        `json:"Cylinders"`
	Transmission  string        `json:"Transmission"`
	MSR           float64       `json:"MSR"`
	Retail        float64       `json:"Retail"`
	DriveWheel    string        `json:"DriveWheel"`
}

type PBSFeed struct {
	Vehicles []PBSVehicle `json:"vehicles"`
}

type Query struct {
	SerialNumber         string
	Year                 string
	Status               string
	IncludeInactive      bool
	IncludeBuildVehicles bool
	ModifiedSince        time.Time
	ModifiedUntil        time.Time
}

const defaultDescription string = `Factory Warranty Plus Our 12 Month Huggable Guarantee!! COMPARE AT NEW MSRP "Pay Less-Owe Less"`

// RefreshFromPBS pulls feed from PBS
func RefreshFromPBS(w http.ResponseWriter, r *http.Request) {
	lastPage := session.GetString(r.Context(), "last-page")
	if lastPage == "" {
		lastPage = "/"
	}

	count, done := PullFromPBS()
	if !done {
		session.Put(r.Context(), "error", "Error connecting to PBS. Try again later")
		http.Redirect(w, r, lastPage, http.StatusSeeOther)
		return
	}

	session.Put(r.Context(), "flash", fmt.Sprintf("Refreshed from PBS. %d items added.", count))
	http.Redirect(w, r, lastPage, http.StatusSeeOther)
}
