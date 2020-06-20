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

// PullFromPBS pulls data from PBS
func PullFromPBS() (int, bool) {
	// read u/p from .env
	err := godotenv.Load("./.env")
	if err != nil {
		errorLog.Println("Error loading .env file")
		return 0, false
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
		return 0, false
	}

	resp, err := http.Post(fmt.Sprintf("https://%s:%s@partnerhub.pbsdealers.com/api/json/reply/VehicleGet", userName, password),
		"application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		errorLog.Println(err)
		return 0, false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errorLog.Println(err)
		return 0, false
	}

	var usedItems PBSFeed
	err = json.Unmarshal(body, &usedItems)
	if err != nil {
		errorLog.Println(err)
		return 0, false
	}

	count := 0
	for _, x := range usedItems.Vehicles {
		exists := vehicleModel.CheckIfVehicleExists(x.StockNumber)
		if !exists {
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
			if modelID == 0 {
				// add new make
				id, err := vehicleModel.InsertModel(makeID, x.Model)
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
				Cost:            float32(x.Retail),
				CreatedAt:       time.Now(),
				Description:     defaultDescription,
				DriveTrain:      x.DriveWheel,
				Engine:          x.Engine,
				ExteriorColour:  x.ExteriorColor.Description,
				InteriorColour:  x.InteriorColor.Description,
				Odometer:        x.Odometer,
				Status:          2,
				StockNo:         x.StockNumber,
				TotalMSR:        float32(x.MSR),
				Transmission:    x.Transmission,
				Trim:            x.Trim,
				UpdatedAt:       time.Now(),
				Used:            1,
				VehicleMakesID:  makeID,
				VehicleModelsID: modelID,
				VehicleType:     vehicleType,
				Vin:             x.VIN,
				Year:            year,
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
		}
	}

	// now do NEW (for powersports)
	parameters = Query{
		SerialNumber:         "2675",
		Year:                 "",
		Status:               "New",
		IncludeInactive:      false,
		IncludeBuildVehicles: false,
		ModifiedSince:        time.Now().Add(-24 * time.Hour),
		ModifiedUntil:        time.Now(),
	}

	reqBody2, err := json.MarshalIndent(parameters, "", "    ")
	if err != nil {
		errorLog.Println(err)
		return 0, false
	}

	resp2, err := http.Post(fmt.Sprintf("https://%s:%s@partnerhub.pbsdealers.com/api/json/reply/VehicleGet", userName, password),
		"application/json", bytes.NewBuffer(reqBody2))
	if err != nil {
		errorLog.Println(err)
		return 0, false
	}
	defer resp2.Body.Close()

	body, err = ioutil.ReadAll(resp2.Body)
	if err != nil {
		errorLog.Println(err)
		return 0, false
	}

	var newItems PBSFeed

	err = json.Unmarshal(body, &newItems)
	if err != nil {
		errorLog.Println(err)
		return 0, false
	}

	for _, x := range newItems.Vehicles {
		exists := vehicleModel.CheckIfVehicleExists(x.StockNumber)
		if !exists {
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
			if modelID == 0 {
				// add new make
				id, err := vehicleModel.InsertModel(makeID, x.Model)
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
				Cost:            float32(x.Retail),
				CreatedAt:       time.Now(),
				Description:     defaultDescription,
				DriveTrain:      x.DriveWheel,
				Engine:          x.Engine,
				ExteriorColour:  x.ExteriorColor.Description,
				InteriorColour:  x.InteriorColor.Description,
				Odometer:        x.Odometer,
				Status:          2,
				StockNo:         x.StockNumber,
				TotalMSR:        float32(x.MSR),
				Transmission:    x.Transmission,
				Trim:            x.Trim,
				UpdatedAt:       time.Now(),
				Used:            0,
				VehicleMakesID:  makeID,
				VehicleModelsID: modelID,
				VehicleType:     vehicleType,
				Vin:             x.VIN,
				Year:            year,
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
		}
	}
	return count, true
}

func PushToKijiji() {

}

func PushToKijiPowerSports() {

}

func PushToCarGurus() {

}

func CleanImages() {

}

func CleanVideos() {

}
