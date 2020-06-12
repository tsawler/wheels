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

	infoLog.Print("userName:", userName)

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
		session.Put(r.Context(), "error", err)
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
	infoLog.Println(string(body))

	session.Put(r.Context(), "flash", "Refreshed from PBS!")
	http.Redirect(w, r, lastPage, http.StatusSeeOther)
}
