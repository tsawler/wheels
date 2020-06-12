package clienthandlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"net/http"
	"os"
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
	75,
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

func RefreshFromPBS(w http.ResponseWriter, r *http.Request) {
	lastPage := session.GetString(r.Context(), "last-page")
	if lastPage == "" {
		lastPage = "/"
	}

	// read u/p from .env
	err := godotenv.Load("./.env")
	if err != nil {
		errorLog.Println("Error loading .env file")
		session.Put(r.Context(), "error", "Error loading .env file")
		http.Redirect(w, r, lastPage, http.StatusSeeOther)
		return
	}

	userName := os.Getenv("PBSUSER")
	password := os.Getenv("PBSPASS")

	parameters := Query{
		SerialNumber:         "2675",
		Year:                 "",
		Status:               "Used",
		IncludeInactive:      false,
		IncludeBuildVehicles: false,
		ModifiedSince:        time.Now().Add(-24 * time.Hour),
		ModifiedUntil:        time.Now(),
	}

	reqBody, err := json.MarshalIndent(parameters, "", "    ")
	if err != nil {
		errorLog.Println(err)
		session.Put(r.Context(), "error", "error unmarshalling ")
		http.Redirect(w, r, lastPage, http.StatusSeeOther)
		return
	}

	resp, err := http.Post(fmt.Sprintf("https://%s:%s@partnerhub.pbsdealers.com/api/json/reply/VehicleGet", userName, password),
		"application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		errorLog.Println(err)
		session.Put(r.Context(), "error", "Error Connecting to PBS!")
		http.Redirect(w, r, lastPage, http.StatusSeeOther)
		return
	}
	defer resp.Body.Close()

	infoLog.Println("Response status:", resp.Status)
	infoLog.Println("Response status code:", resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errorLog.Println(err)
	}

	var usedItems PBSFeed
	err = json.Unmarshal(body, &usedItems)
	if err != nil {
		errorLog.Println(err)
		session.Put(r.Context(), "error", "Error unmarshalling json from PBS!")
		http.Redirect(w, r, lastPage, http.StatusSeeOther)
		return
	}

	count := 0
	for _, x := range usedItems.Vehicles {
		infoLog.Println(x.StockNumber)
		exists := vehicleModel.CheckIfVehicleExists(x.StockNumber)
		if !exists {
			infoLog.Println("we don't have", x.StockNumber)
			count++
		}
		infoLog.Println("-----------------------------------")
	}

	session.Put(r.Context(), "flash", fmt.Sprintf("Refreshed from PBS. %d items added.", count))
	http.Redirect(w, r, lastPage, http.StatusSeeOther)
}
