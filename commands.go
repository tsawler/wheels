package clienthandlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/tsawler/goblender/client/clienthandlers/clientmodels"
	"github.com/tsawler/goblender/pkg/handlers"
	"github.com/tushar2708/altcsv"
	"io/ioutil"
	"log"
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

// CarGuruFeed manually pushes to CarGurus
func CarGuruFeed(w http.ResponseWriter, r *http.Request) {
	lastPage := session.GetString(r.Context(), "last-page")
	if lastPage == "" {
		lastPage = "/"
	}

	err := PushToCarGurus()
	if err != nil {
		// audit trail
		history := handlers.History{
			UserID:     1,
			Message:    "Push to CarGurus failed",
			ChangeType: "page",
			UserName:   fmt.Sprintf("%s %s", "System", "System"),
		}
		repo.AddHistory(history)
		session.Put(r.Context(), "flash", fmt.Sprintf("Error pushing to CarGurus:", err.Error()))
		http.Redirect(w, r, lastPage, http.StatusSeeOther)
		return
	}

	session.Put(r.Context(), "flash", "Pushed to CarGurus.")
	http.Redirect(w, r, lastPage, http.StatusSeeOther)
}

// KijijiFeed manually pushes feed to kijiji
func KijijiFeed(w http.ResponseWriter, r *http.Request) {
	lastPage := session.GetString(r.Context(), "last-page")
	if lastPage == "" {
		lastPage = "/"
	}

	err := PushToKijiji()
	if err != nil {
		// audit trail
		history := handlers.History{
			UserID:     1,
			Message:    "Push to Kijiji failed",
			ChangeType: "page",
			UserName:   fmt.Sprintf("%s %s", "System", "System"),
		}
		repo.AddHistory(history)

		session.Put(r.Context(), "flash", fmt.Sprintf("Error pushing to CarGurus:", err.Error()))
		http.Redirect(w, r, lastPage, http.StatusSeeOther)
		return
	}

	session.Put(r.Context(), "flash", "Pushed to Kijiji")
	http.Redirect(w, r, lastPage, http.StatusSeeOther)
}

// KijijiPSFeed manually pushes feed to kijiji
func KijijiPSFeed(w http.ResponseWriter, r *http.Request) {
	lastPage := session.GetString(r.Context(), "last-page")
	if lastPage == "" {
		lastPage = "/"
	}

	err := PushToKijijiPowerSports()
	if err != nil {
		// audit trail
		history := handlers.History{
			UserID:     1,
			Message:    "Push to Kijiji (PowerSports) failed",
			ChangeType: "page",
			UserName:   fmt.Sprintf("%s %s", "System", "System"),
		}
		repo.AddHistory(history)

		session.Put(r.Context(), "flash", fmt.Sprintf("Error pushing:", err.Error()))
		http.Redirect(w, r, lastPage, http.StatusSeeOther)
		return
	}

	session.Put(r.Context(), "flash", "Pushed PowerSports to Kijiji")
	http.Redirect(w, r, lastPage, http.StatusSeeOther)
}

// PushToCarGurus does push of CSV
func PushToCarGurus() error {
	records, err := vehicleModel.CarGurus()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fileName := "./tmp/car_gurus.csv"
	fileWriter, _ := os.Create(fileName)
	feedWriter := altcsv.NewWriter(fileWriter)
	feedWriter.AllQuotes = true

	for _, record := range records {
		if err := feedWriter.Write(record); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}

	// Write any buffered data to the underlying writer (standard output).
	feedWriter.Flush()

	if err := feedWriter.Error(); err != nil {
		errorLog.Println(err)
		return err
	}

	//// FTP the file up to CarGurus
	//err := godotenv.Load("./.env")
	//if err != nil {
	//	errorLog.Println("Error loading .env file")
	//}
	//
	//// ftp file
	//userName := os.Getenv("CARGURUSUSER")
	//password := os.Getenv("CARGURUSPASS")
	//host := os.Getenv("CARGURUHOST")
	//err = PushFTPFile(userName, password, fmt.Sprintf("%s:21", host), fileName, "feed.csv")
	//if err != nil {
	//	errorLog.Println(err)
	//
	//}

	// audit trail
	history := handlers.History{
		UserID:     1,
		Message:    "Push to CarGurus complete",
		ChangeType: "page",
		UserName:   fmt.Sprintf("%s %s", "System", "System"),
	}
	repo.AddHistory(history)

	return nil
}

// PushToKijiji does push of CSV
func PushToKijiji() error {
	records, err := vehicleModel.Kijiji()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fileName := "./tmp/Kijiji.csv"
	fileWriter, _ := os.Create(fileName)
	feedWriter := altcsv.NewWriter(fileWriter)
	feedWriter.AllQuotes = true

	for _, record := range records {
		if err := feedWriter.Write(record); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}

	// Write any buffered data to the underlying writer (standard output).
	feedWriter.Flush()

	if err := feedWriter.Error(); err != nil {
		errorLog.Println(err)
		return err
	}

	//// FTP the file up to Kijiji
	//err := godotenv.Load("./.env")
	//if err != nil {
	//	errorLog.Println("Error loading .env file")
	//}
	//
	//// ftp file
	//userName := os.Getenv("KIJIJIUSER")
	//password := os.Getenv("KIJIJIPASS")
	//host := os.Getenv("KIJIJIHOST")
	//err = PushFTPFile(userName, password, fmt.Sprintf("%s:21", host), fileName, "Kijiji.csv")
	//if err != nil {
	//	errorLog.Println(err)
	//
	//}

	// audit trail
	history := handlers.History{
		UserID:     1,
		Message:    "Push to Kijiji complete",
		ChangeType: "page",
		UserName:   fmt.Sprintf("%s %s", "System", "System"),
	}
	repo.AddHistory(history)

	return nil
}

// PushToKijijiPowerSports does push of CSV
func PushToKijijiPowerSports() error {
	records, err := vehicleModel.KijijiPS()
	if err != nil {
		fmt.Println(err)
	}

	fileName := "./tmp/PowerSportsKijiji.csv"
	fileWriter, _ := os.Create(fileName)
	feedWriter := altcsv.NewWriter(fileWriter)
	feedWriter.AllQuotes = true

	for _, record := range records {
		if err := feedWriter.Write(record); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}

	// Write any buffered data to the underlying writer (standard output).
	feedWriter.Flush()

	if err := feedWriter.Error(); err != nil {
		errorLog.Println(err)
		return err
	}

	//// FTP the file up to Kijiji
	//err := godotenv.Load("./.env")
	//if err != nil {
	//	errorLog.Println("Error loading .env file")
	//}
	//
	//// ftp file
	//userName := os.Getenv("KIJIJIPSUSER")
	//password := os.Getenv("KIJIJIPSPASS")
	//host := os.Getenv("KIJIJIPSHOST")
	//err = PushFTPFile(userName, password, fmt.Sprintf("%s:21", host), fileName, "Kijiji.csv")
	//if err != nil {
	//	errorLog.Println(err)
	//
	//}

	// audit trail
	history := handlers.History{
		UserID:     1,
		Message:    "Push to Kijiji (PowerSports) complete",
		ChangeType: "page",
		UserName:   fmt.Sprintf("%s %s", "System", "System"),
	}
	repo.AddHistory(history)

	return nil
}

func CleanImages() {

}

func CleanVideos() {

}

func CleanPanoramas() {

}
