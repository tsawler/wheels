package clienthandlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/tsawler/goblender/client/clienthandlers/clientmodels"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
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

	//fmt.Print(string(body))

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
		exists := vehicleModel.CheckIfVehicleExists(x.StockNumber)
		infoLog.Println("Checking", x.StockNumber)
		if !exists {
			infoLog.Println("ADDING", x.StockNumber)
			count++

			// see if we have this make
			makeID := vehicleModel.GetMakeByName(x.Make)
			if makeID == 0 {
				// add new make
				id, err := vehicleModel.InsertMake(x.Make)
				if err != nil {
					errorLog.Print(err)
				}
				makeID = id
			}

			// see if we have this model
			modelID := vehicleModel.GetModelByName(x.Model)
			if makeID == 0 {
				// add new make
				id, err := vehicleModel.InsertMake(x.Make)
				if err != nil {
					errorLog.Print(err)
				}
				modelID = id
			}

			if makeID == 0 || modelID == 0 {
				errorLog.Print("Cannot process!")
				continue
			}

			year, _ := strconv.Atoi(x.Year)

			vehicleType := 0

			if strings.ToUpper(x.VehicleType) == "CAR" {
				vehicleType = 1
			} else if strings.ToUpper(x.VehicleType) == "P" {
				vehicleType = 1
			} else if strings.ToUpper(x.VehicleType) == "PASSENGER" {
				vehicleType = 1
			} else if strings.ToUpper(x.VehicleType) == "T" {
				vehicleType = 2
			} else if strings.ToUpper(x.VehicleType) == "TRUCK" {
				vehicleType = 2
			} else {
				vehicleType = 3
			}

			v := clientmodels.Vehicle{
				StockNo:         x.StockNumber,
				Vin:             x.VIN,
				Odometer:        x.Odometer,
				Year:            year,
				VehicleMakesID:  makeID,
				VehicleModelsID: modelID,
				Trim:            x.Trim,
				Engine:          x.Engine,
				Transmission:    x.Transmission,
				TotalMSR:        float32(x.MSR),
				ExteriorColour:  x.ExteriorColor.Description,
				InteriorColour:  x.InteriorColor.Description,
				VehicleType:     vehicleType,
				DriveTrain:      x.DriveWheel,
				Status:          2,
				Description:     defaultDescription,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}

			vid, err := vehicleModel.InsertVehicle(v)
			if err != nil {
				errorLog.Print(err)
			} else {
				for _, y := range defaultOptions {
					o := clientmodels.VehicleOption{
						VehicleID: vid,
						OptionID:  y,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}
					err := vehicleModel.InsertVehicleOption(o)
					if err != nil {
						errorLog.Print(err)
					}
				}
			}
			infoLog.Println("Inserted vehicle id", vid, " - stock no ", x.StockNumber)
		}
		infoLog.Println("-----------------------------------")
	}

	session.Put(r.Context(), "flash", fmt.Sprintf("Refreshed from PBS. %d items added.", count))
	http.Redirect(w, r, lastPage, http.StatusSeeOther)
}
