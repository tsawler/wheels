package clienthandlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gosimple/slug"
	"github.com/tsawler/goblender/client/clienthandlers/clientmodels"
	"github.com/tsawler/goblender/pkg/datatables"
	"github.com/tsawler/goblender/pkg/forms"
	"github.com/tsawler/goblender/pkg/helpers"
	"github.com/tsawler/goblender/pkg/images"
	"github.com/tsawler/goblender/pkg/templates"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	All           = 0
	ATVBruteForce = 8
	ATVMule       = 11
	ATVTeryx      = 12
	Car           = 1
	ElectricBike  = 16
	JetSki        = 13
	Mercury       = 10
	Motorcycle    = 7
	Other         = 3
	PontoonBoat   = 9
	PowerBoat     = 15
	Scooter       = 17
	SUV           = 5
	Trailer       = 14
	Truck         = 2
	MiniVan       = 6
	Unknown       = 4
)

// JSONResp holds a json response message
type JsonResponse struct {
	Ok      bool   `json:"okay"`
	Message string `json:"message"`
}

// DataTablesJSON holds the json for datatables
type DataTablesJSON struct {
	Draw            int64                       `json:"draw"`
	RecordsTotal    int64                       `json:"recordsTotal"`
	RecordsFiltered int64                       `json:"recordsFiltered"`
	DataRows        []*clientmodels.VehicleJSON `json:"data"`
}

// CreditAppJSON holds the json for datatables
type CreditAppJSON struct {
	Draw            int64                     `json:"draw"`
	RecordsTotal    int64                     `json:"recordsTotal"`
	RecordsFiltered int64                     `json:"recordsFiltered"`
	DataRows        []*clientmodels.CreditApp `json:"data"`
}

// QuickQuoteJSON holds the json for datatables
type QuickQuoteJSON struct {
	Draw            int64                      `json:"draw"`
	RecordsTotal    int64                      `json:"recordsTotal"`
	RecordsFiltered int64                      `json:"recordsFiltered"`
	DataRows        []*clientmodels.QuickQuote `json:"data"`
}

// TestDriveJSON holds the json for datatables
type TestDriveJSON struct {
	Draw            int64                     `json:"draw"`
	RecordsTotal    int64                     `json:"recordsTotal"`
	RecordsFiltered int64                     `json:"recordsFiltered"`
	DataRows        []*clientmodels.TestDrive `json:"data"`
}

// SortOrder struct for sorting images
type SortOrder struct {
	ImageID    string `json:"id"`
	StepNumber int    `json:"order"`
}

// DisplayVehicleForAdmin shows vehicle for edit
func DisplayVehicleForAdmin(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get(":ID"))
	if err != nil {
		errorLog.Println(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

	src := r.URL.Query().Get(":SRC")
	segment := r.URL.Query().Get(":TYPE")
	category := r.URL.Query().Get(":CATEGORY")
	stringMap := make(map[string]string)
	stringMap["segment"] = segment
	stringMap["src"] = src
	stringMap["category"] = category

	vehicle, err := vehicleModel.GetVehicleByID(id)
	if err != nil {
		errorLog.Println(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

	rowSets := make(map[string]interface{})
	rowSets["vehicle"] = vehicle

	var years []int
	for i := (time.Now().Year() + 1); i >= 1900; i-- {
		years = append(years, i)
	}

	rowSets["years"] = years

	makes, err := vehicleModel.GetMakes()
	if err != nil {
		errorLog.Println(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	rowSets["makes"] = makes

	models, err := vehicleModel.GetModelsForMakeID(vehicle.VehicleMakesID)
	if err != nil {
		errorLog.Println(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	rowSets["models"] = models

	options, err := vehicleModel.AllActiveOptions()
	if err != nil {
		errorLog.Println(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	rowSets["options"] = options

	// add map of options
	intMap := make(map[string]int)
	for _, x := range vehicle.VehicleOptions {
		intMap[fmt.Sprintf("option_%d", x.OptionID)] = 1
	}

	helpers.Render(w, r, "vehicle.page.tmpl", &templates.TemplateData{
		RowSets:   rowSets,
		IntMap:    intMap,
		Form:      forms.New(nil),
		StringMap: stringMap,
	})

}

// DisplayVehicleForAdminPost handles post of vehicle
func DisplayVehicleForAdminPost(w http.ResponseWriter, r *http.Request) {
	vehicleID, _ := strconv.Atoi(r.URL.Query().Get(":ID"))

	form := forms.New(r.PostForm, app.Database)
	category := form.Get("category")
	segment := form.Get("segment")
	src := form.Get("src")

	v, err := vehicleModel.GetVehicleByID(vehicleID)
	if err != nil {
		errorLog.Println(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	oldVideoID := v.Video.VideoID

	form.Required("stock_no", "vin", "cost", "total_msr")
	form.IsFloat("cost")
	form.IsFloat("total_msr")
	form.IsInt("odometer")

	if !form.Valid() {
		stringMap := make(map[string]string)
		stringMap["segment"] = segment
		stringMap["src"] = src
		stringMap["category"] = category

		rowSets := make(map[string]interface{})
		rowSets["vehicle"] = v
		var years []int
		for i := time.Now().Year() + 1; i >= 1900; i-- {
			years = append(years, i)
		}

		rowSets["years"] = years

		makes, err := vehicleModel.GetMakes()
		if err != nil {
			errorLog.Println(err)
			helpers.ClientError(w, http.StatusBadRequest)
			return
		}
		rowSets["makes"] = makes
		models, err := vehicleModel.GetModelsForMakeID(v.VehicleMakesID)
		if err != nil {
			errorLog.Println(err)
			helpers.ClientError(w, http.StatusBadRequest)
			return
		}
		rowSets["models"] = models

		options, err := vehicleModel.AllActiveOptions()
		if err != nil {
			errorLog.Println(err)
			helpers.ClientError(w, http.StatusBadRequest)
			return
		}
		rowSets["options"] = options

		// add map of options
		intMap := make(map[string]int)
		for _, x := range v.VehicleOptions {
			intMap[fmt.Sprintf("option_%d", x.OptionID)] = 1
		}

		helpers.Render(w, r, "vehicle.page.tmpl", &templates.TemplateData{
			RowSets:   rowSets,
			IntMap:    intMap,
			Form:      form,
			StringMap: stringMap,
		})
		return
	}
	year, _ := strconv.Atoi(form.Get("year"))
	v.Year = year

	vehicleType, _ := strconv.Atoi(form.Get("vehicle_type"))
	v.VehicleType = vehicleType

	vehicleMakesID, _ := strconv.Atoi(form.Get("vehicle_makes_id"))
	v.VehicleMakesID = vehicleMakesID

	vehicleModelsID, _ := strconv.Atoi(form.Get("vehicle_models_id"))
	v.VehicleModelsID = vehicleModelsID

	used, _ := strconv.Atoi(form.Get("used"))
	v.Used = used

	handPicked, _ := strconv.Atoi(form.Get("hand_picked"))
	v.HandPicked = handPicked

	if cost, err := strconv.ParseFloat(form.Get("cost"), 32); err == nil {
		v.Cost = float32(cost)
	}

	if totalMSR, err := strconv.ParseFloat(form.Get("total_msr"), 32); err == nil {
		v.TotalMSR = float32(totalMSR)
	}

	v.PriceForDisplay = form.Get("price_for_display")

	v.Trim = r.Form.Get("trim")

	odometer, _ := strconv.Atoi(form.Get("odometer"))
	v.Odometer = odometer

	v.InteriorColour = form.Get("interior_color")
	v.ExteriorColour = form.Get("exterior_color")
	v.Body = form.Get("body")
	v.Engine = form.Get("engine")
	v.Transmission = form.Get("transmission")
	v.DriveTrain = form.Get("drive_train")

	status, _ := strconv.Atoi(form.Get("status"))
	v.Status = status
	v.StockNo = form.Get("stock_no")
	v.Vin = form.Get("vin")
	action, _ := strconv.Atoi(form.Get("action"))
	v.Description = form.Get("description")

	err = vehicleModel.UpdateVehicle(v)
	if err != nil {
		errorLog.Println(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

	// vehicle options
	// first delete all options
	_ = vehicleModel.DeleteAllVehicleOptions(v.ID)

	// loop through all posted vars, and add options
	for key, value := range r.Form {
		if strings.HasPrefix(key, "option_") {
			optionID, _ := strconv.Atoi(value[0])
			o := clientmodels.VehicleOption{
				VehicleID: v.ID,
				OptionID:  optionID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			err = vehicleModel.InsertVehicleOption(o)
			if err != nil {
				errorLog.Println(err)
			}
		}
	}

	// update sort order for images
	sortList := r.Form.Get("sort_list")
	var sorted []SortOrder

	err = json.Unmarshal([]byte(sortList), &sorted)
	if err != nil {
		app.ErrorLog.Println(err)
	}

	for _, v := range sorted {
		imageID, _ := strconv.Atoi(v.ImageID)
		err := vehicleModel.UpdateSortOrderForImage(imageID, v.StepNumber)
		if err != nil {
			app.ErrorLog.Println(err)
		}
	}

	// handle video
	videoID, _ := strconv.Atoi(form.Get("video_id"))
	if videoID != oldVideoID {
		if (videoID == 0 && oldVideoID != 0) || (videoID > 0 && oldVideoID != 0) {
			vv := clientmodels.VehicleVideo{
				VehicleID: v.ID,
				VideoID:   videoID,
				UpdatedAt: time.Now(),
			}
			err := vehicleModel.UpdateVideoForVehicle(vv)
			if err != nil {
				errorLog.Println("Error updating video:", err)
			}
		} else if videoID > 0 {
			vv := clientmodels.VehicleVideo{
				VehicleID: v.ID,
				VideoID:   videoID,
				UpdatedAt: time.Now(),
			}
			err := vehicleModel.InsertVideoForVehicle(vv)
			if err != nil {
				errorLog.Println("Error inserting video:", err)
			}
		}
	}

	// handle panorama
	if form.HasFile("panorama", r) {
		// we have a panorama
		fileName, _, _ := helpers.UploadOneFile(r, "./tmp/")

		oldLocation := fmt.Sprintf("./tmp/%s", fileName)
		newLocation := fmt.Sprintf("./ui/static/site-content/panoramas/%d-%s", vehicleID, fileName)

		err := MoveFile(oldLocation, newLocation)
		if err != nil {
			app.ErrorLog.Println("could not move from", oldLocation, "to", newLocation)
		}

		// update in DB
		vp := clientmodels.Panorama{
			VehicleID: vehicleID,
			Panorama:  fmt.Sprintf("%d-%s", vehicleID, fileName),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		v.Panorama = vp

		err = vehicleModel.UpdatePanorama(vp)
		if err != nil {
			errorLog.Println(err)
		}
	}

	// redirect
	session.Put(r.Context(), "flash", "Changes saved")
	if action == 1 {
		http.Redirect(w, r, fmt.Sprintf("/admin/%s/%s/%s", category, segment, src), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/%s/%s/%s/%d", category, segment, src, vehicleID), http.StatusSeeOther)
}

// AllVehicles displays all vehicles
func AllVehicles(w http.ResponseWriter, r *http.Request) {
	helpers.Render(w, r, "all-vehicles.page.tmpl", &templates.TemplateData{})
}

// AllVehiclesJSON returns  json for all vehicles regardless of status
func AllVehiclesJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

	dtinfo, err := datatables.ParseDatatablesRequest(r)
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	draw := dtinfo.Draw

	query, baseQuery, err := dtinfo.BuildQuery("v_all_vehicles")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Do the queries and get back our data, the row count, and the filtered row count
	v, rowCount, filterCount, err := vehicleModel.VehicleJSON(query, baseQuery, "vehicle_type < 7")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	theData := DataTablesJSON{
		Draw:            int64(draw),
		RecordsTotal:    int64(rowCount),
		RecordsFiltered: int64(filterCount),
		DataRows:        v,
	}

	out, err := json.MarshalIndent(theData, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// AllVehiclesForSale displays table of all cars/trucks for sale
func AllVehiclesForSale(w http.ResponseWriter, r *http.Request) {
	helpers.Render(w, r, "all-vehicles-for-sale.page.tmpl", &templates.TemplateData{})
}

// AllVehiclesForSaleJSON returns json for cars/trucks for sale
func AllVehiclesForSaleJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

	dtinfo, err := datatables.ParseDatatablesRequest(r)
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	draw := dtinfo.Draw

	query, baseQuery, err := dtinfo.BuildQuery("v_all_vehicles")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Do the queries and get back our data, the row count, and the filtered row count
	v, rowCount, filterCount, err := vehicleModel.VehicleJSON(query, baseQuery, "vehicle_status = 1 and vehicle_type <  7")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	theData := DataTablesJSON{
		Draw:            int64(draw),
		RecordsTotal:    int64(rowCount),
		RecordsFiltered: int64(filterCount),
		DataRows:        v,
	}

	out, err := json.MarshalIndent(theData, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// AllPowerSportsForSale displays table of all powersports for sale
func AllPowerSportsForSale(w http.ResponseWriter, r *http.Request) {
	helpers.Render(w, r, "all-powersports-for-sale.page.tmpl", &templates.TemplateData{})
}

// AllPowerSportsForSaleJSON returns json for powersports for sale
func AllPowerSportsForSaleJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

	dtinfo, err := datatables.ParseDatatablesRequest(r)
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	draw := dtinfo.Draw

	query, baseQuery, err := dtinfo.BuildQuery("v_all_vehicles")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Do the queries and get back our data, the row count, and the filtered row count
	v, rowCount, filterCount, err := vehicleModel.VehicleJSON(query, baseQuery, "vehicle_status = 1 and vehicle_type >=  7")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	theData := DataTablesJSON{
		Draw:            int64(draw),
		RecordsTotal:    int64(rowCount),
		RecordsFiltered: int64(filterCount),
		DataRows:        v,
	}

	out, err := json.MarshalIndent(theData, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// AllSold displays table of all sold cars/trucks
func AllSold(w http.ResponseWriter, r *http.Request) {
	helpers.Render(w, r, "all-sold.page.tmpl", &templates.TemplateData{})
}

// AllSoldJSON returns json for sold cars/trucks
func AllSoldJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

	dtinfo, err := datatables.ParseDatatablesRequest(r)
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	draw := dtinfo.Draw

	query, baseQuery, err := dtinfo.BuildQuery("v_all_vehicles")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Do the queries and get back our data, the row count, and the filtered row count
	v, rowCount, filterCount, err := vehicleModel.VehicleJSON(query, baseQuery, "vehicle_status = 0 and vehicle_type < 7")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	theData := DataTablesJSON{
		Draw:            int64(draw),
		RecordsTotal:    int64(rowCount),
		RecordsFiltered: int64(filterCount),
		DataRows:        v,
	}

	out, err := json.MarshalIndent(theData, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// AllSoldThisMonth displays table of all sold cars/trucks this month
func AllSoldThisMonth(w http.ResponseWriter, r *http.Request) {
	helpers.Render(w, r, "all-sold-this-month.page.tmpl", &templates.TemplateData{})
}

// AllSoldThisMonthJSON returns json for sold cars/trucks this month
func AllSoldThisMonthJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

	dtinfo, err := datatables.ParseDatatablesRequest(r)
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	draw := dtinfo.Draw

	query, baseQuery, err := dtinfo.BuildQuery("v_all_vehicles")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Do the queries and get back our data, the row count, and the filtered row count
	thisMonth := fmt.Sprintf("%d-%d-01", time.Now().Year(), time.Now().Month())
	v, rowCount, filterCount, err := vehicleModel.VehicleJSON(query, baseQuery, fmt.Sprintf("vehicle_status = 0 and vehicle_type < 7 and updated_at > '%s'", thisMonth))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	theData := DataTablesJSON{
		Draw:            int64(draw),
		RecordsTotal:    int64(rowCount),
		RecordsFiltered: int64(filterCount),
		DataRows:        v,
	}

	out, err := json.MarshalIndent(theData, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// AllPowerSportsSold displays table of all sold cars/trucks
func AllPowerSportsSold(w http.ResponseWriter, r *http.Request) {
	helpers.Render(w, r, "all-powersports-sold.page.tmpl", &templates.TemplateData{})
}

// AllPowerSportsSoldJSON returns json for sold cars/trucks
func AllPowerSportsSoldJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

	dtinfo, err := datatables.ParseDatatablesRequest(r)
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	draw := dtinfo.Draw

	query, baseQuery, err := dtinfo.BuildQuery("v_all_vehicles")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Do the queries and get back our data, the row count, and the filtered row count
	v, rowCount, filterCount, err := vehicleModel.VehicleJSON(query, baseQuery, "vehicle_status = 0 and vehicle_type >= 7")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	theData := DataTablesJSON{
		Draw:            int64(draw),
		RecordsTotal:    int64(rowCount),
		RecordsFiltered: int64(filterCount),
		DataRows:        v,
	}

	out, err := json.MarshalIndent(theData, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// AllPowerSportsSoldThisMonth displays table of all sold powersports for this month
func AllPowerSportsSoldThisMonth(w http.ResponseWriter, r *http.Request) {
	helpers.Render(w, r, "all-powersports-sold-this-month.page.tmpl", &templates.TemplateData{})
}

// AllPowerSportsSoldThisMonthJSON returns json for sold powersports for this month
func AllPowerSportsSoldThisMonthJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

	dtinfo, err := datatables.ParseDatatablesRequest(r)
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	draw := dtinfo.Draw

	query, baseQuery, err := dtinfo.BuildQuery("v_all_vehicles")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Do the queries and get back our data, the row count, and the filtered row count
	thisMonth := fmt.Sprintf("%d-%d-01", time.Now().Year(), time.Now().Month())
	v, rowCount, filterCount, err := vehicleModel.VehicleJSON(query, baseQuery, fmt.Sprintf("vehicle_status = 0 and vehicle_type >= 7 and updated_at > '%s'", thisMonth))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	theData := DataTablesJSON{
		Draw:            int64(draw),
		RecordsTotal:    int64(rowCount),
		RecordsFiltered: int64(filterCount),
		DataRows:        v,
	}

	out, err := json.MarshalIndent(theData, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// AllVehiclesPending displays table of all pending
func AllVehiclesPending(w http.ResponseWriter, r *http.Request) {
	helpers.Render(w, r, "all-vehicles-pending.page.tmpl", &templates.TemplateData{})
}

// AllVehiclesPendingJSON returns json for pending
func AllVehiclesPendingJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

	dtinfo, err := datatables.ParseDatatablesRequest(r)
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	draw := dtinfo.Draw

	query, baseQuery, err := dtinfo.BuildQuery("v_all_vehicles")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Do the queries and get back our data, the row count, and the filtered row count
	v, rowCount, filterCount, err := vehicleModel.VehicleJSON(query, baseQuery, "vehicle_status = 2")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	theData := DataTablesJSON{
		Draw:            int64(draw),
		RecordsTotal:    int64(rowCount),
		RecordsFiltered: int64(filterCount),
		DataRows:        v,
	}

	out, err := json.MarshalIndent(theData, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// AllVehiclesTradeIns displays table of all pending
func AllVehiclesTradeIns(w http.ResponseWriter, r *http.Request) {
	helpers.Render(w, r, "all-vehicles-trade-ins.page.tmpl", &templates.TemplateData{})
}

// AllVehiclesTradeInsJSON returns json for pending
func AllVehiclesTradeInsJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

	dtinfo, err := datatables.ParseDatatablesRequest(r)
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	draw := dtinfo.Draw

	query, baseQuery, err := dtinfo.BuildQuery("v_all_vehicles")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Do the queries and get back our data, the row count, and the filtered row count
	v, rowCount, filterCount, err := vehicleModel.VehicleJSON(query, baseQuery, "vehicle_status = 3")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	theData := DataTablesJSON{
		Draw:            int64(draw),
		RecordsTotal:    int64(rowCount),
		RecordsFiltered: int64(filterCount),
		DataRows:        v,
	}

	out, err := json.MarshalIndent(theData, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// VehicleImagesJSON returns a vehicle's images as JSON
func VehicleImagesJSON(w http.ResponseWriter, r *http.Request) {
	vehicleID, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	v, _ := vehicleModel.GetVehicleByID(vehicleID)

	out, err := json.MarshalIndent(v.Images, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// VehicleImageDelete deletes an image and returns json
func VehicleImageDelete(w http.ResponseWriter, r *http.Request) {
	imageID, _ := strconv.Atoi(r.URL.Query().Get(":ID"))

	image, _ := vehicleModel.GetVehicleImageByID(imageID)
	// delete this image file
	sourcePath := fmt.Sprintf("./ui/static/site-content/inventory/%d/%s", image.VehicleID, image.Image)
	_ = os.Remove(sourcePath)

	okay := true
	message := ""
	err := vehicleModel.DeleteVehicleImage(imageID)
	if err != nil {
		errorLog.Println(err)
		okay = false
		message = err.Error()
	}

	resp := JsonResponse{}
	resp.Ok = okay
	resp.Message = message

	// build the json response from the struct
	out, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// send json to client
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(out)
	if err != nil {
		errorLog.Println(err)
	}
}

// PrintWindowSticker prints a window sticker to pdf (as stream) and downloads it to client
func PrintWindowSticker(w http.ResponseWriter, r *http.Request) {
	vehicleID, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	v, err := vehicleModel.GetVehicleByID(vehicleID)
	if err != nil {
		lastPage := app.Session.GetString(r.Context(), "last-page")
		session.Put(r.Context(), "error", "Unable to find vehicle!")
		http.Redirect(w, r, lastPage, http.StatusSeeOther)
		return
	}

	pdf, err := CreateWindowSticker(v)
	if err != nil {
		lastPage := app.Session.GetString(r.Context(), "last-page")
		errorLog.Println(err)
		session.Put(r.Context(), "error", "Unable to generate PDF!")
		http.Redirect(w, r, lastPage, http.StatusSeeOther)
		return
	}

	//w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.pdf", v.StockNo))

	out := &bytes.Buffer{}
	if err := pdf.Output(out); err != nil {
		errorLog.Println(err)
		lastPage := app.Session.GetString(r.Context(), "last-page")
		session.Put(r.Context(), "error", "Unable to write PDF!")
		http.Redirect(w, r, lastPage, http.StatusSeeOther)
		return
	}
	b := out.Bytes()

	_, err = w.Write(b)
	if err != nil {
		errorLog.Println(err)
	}

}

// CompareVehicles Show 2 or 3 vehicles in table TODO
func CompareVehicles(w http.ResponseWriter, r *http.Request) {
	idString := r.Form.Get("ids")
	infoLog.Println("Ids:", idString)

	ids := strings.Split(idString, ",")
	var items []clientmodels.Vehicle

	for _, x := range ids {
		infoLog.Println("ID:", x)
		vid, _ := strconv.Atoi(x)
		v, _ := vehicleModel.GetVehicleByID(vid)
		items = append(items, v)
	}

	rowSets := make(map[string]interface{})
	rowSets["items"] = items
	helpers.Render(w, r, "compare.page.tmpl", &templates.TemplateData{
		RowSets: rowSets,
	})
}

// AllCreditApplications displays all credit applications
func AllCreditApplications(w http.ResponseWriter, r *http.Request) {
	helpers.Render(w, r, "all-credit-apps.page.tmpl", &templates.TemplateData{})
}

// OneCreditApp displays one credit application
func OneCreditApp(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	c, _ := vehicleModel.GetCreditApp(id)

	rowSet := make(map[string]interface{})
	rowSet["app"] = c
	helpers.Render(w, r, "one-credit-app.page.tmpl", &templates.TemplateData{
		RowSets: rowSet,
	})
}

// AllCreditAppsJSON returns json for credit apps
func AllCreditAppsJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

	dtinfo, err := datatables.ParseDatatablesRequest(r)
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	draw := dtinfo.Draw

	query, baseQuery, err := dtinfo.BuildQuery("credit_applications")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Do the queries and get back our data, the row count, and the filtered row count
	v, rowCount, filterCount, err := vehicleModel.CreditJSON(query, baseQuery)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	theData := CreditAppJSON{
		Draw:            int64(draw),
		RecordsTotal:    int64(rowCount),
		RecordsFiltered: int64(filterCount),
		DataRows:        v,
	}

	out, err := json.MarshalIndent(theData, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// AllQuickQuotes displays all credit applications
func AllQuickQuotes(w http.ResponseWriter, r *http.Request) {
	helpers.Render(w, r, "all-quick-quotes.page.tmpl", &templates.TemplateData{})
}

// OneQuickQuote displays one quick quote
func OneQuickQuote(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	c, _ := vehicleModel.GetQuickQuote(id)

	rowSet := make(map[string]interface{})
	rowSet["app"] = c
	helpers.Render(w, r, "one-quick-quote.page.tmpl", &templates.TemplateData{
		RowSets: rowSet,
	})
}

// AllQuickQuotesJSON returns json for credit apps
func AllQuickQuotesJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

	dtinfo, err := datatables.ParseDatatablesRequest(r)
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	draw := dtinfo.Draw

	query, baseQuery, err := dtinfo.BuildQuery("quick_quotes")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Do the queries and get back our data, the row count, and the filtered row count
	v, rowCount, filterCount, err := vehicleModel.QuickQuotesJSON(query, baseQuery)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	theData := QuickQuoteJSON{
		Draw:            int64(draw),
		RecordsTotal:    int64(rowCount),
		RecordsFiltered: int64(filterCount),
		DataRows:        v,
	}

	out, err := json.MarshalIndent(theData, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// AllTestDrives displays all test drives
func AllTestDrives(w http.ResponseWriter, r *http.Request) {
	helpers.Render(w, r, "all-test-drives.page.tmpl", &templates.TemplateData{})
}

// OneTestDrive displays one test drive
func OneTestDrive(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	c, _ := vehicleModel.GetTestDrive(id)

	rowSet := make(map[string]interface{})
	rowSet["app"] = c
	helpers.Render(w, r, "one-test-drive.page.tmpl", &templates.TemplateData{
		RowSets: rowSet,
	})
}

// AllTestDrivesJSON returns json for test drives
func AllTestDrivesJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

	dtinfo, err := datatables.ParseDatatablesRequest(r)
	if err != nil {
		app.ErrorLog.Print(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	draw := dtinfo.Draw

	query, baseQuery, err := dtinfo.BuildQuery("test_drives")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Do the queries and get back our data, the row count, and the filtered row count
	v, rowCount, filterCount, err := vehicleModel.TestDrivesJSON(query, baseQuery)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	theData := TestDriveJSON{
		Draw:            int64(draw),
		RecordsTotal:    int64(rowCount),
		RecordsFiltered: int64(filterCount),
		DataRows:        v,
	}

	out, err := json.MarshalIndent(theData, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// OptionsAll lists all options
func OptionsAll(w http.ResponseWriter, r *http.Request) {
	rowSets := make(map[string]interface{})
	options, err := vehicleModel.GetOptions()
	if err != nil {
		errorLog.Println(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	rowSets["options"] = options

	helpers.Render(w, r, "options-all.page.tmpl", &templates.TemplateData{
		RowSets: rowSets,
	})
}

// DisplayOneOption displays option for add/edit
func DisplayOneOption(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	rowSets := make(map[string]interface{})

	var o clientmodels.Option

	if id > 0 {
		op, err := vehicleModel.GetOneOption(id)
		if err != nil {
			errorLog.Println(err)
			helpers.ClientError(w, http.StatusBadRequest)
			return
		}
		o = op
	}
	rowSets["option"] = o

	helpers.Render(w, r, "option.page.tmpl", &templates.TemplateData{
		RowSets: rowSets,
		Form:    forms.New(nil),
	})
}

// DisplayOneOptionPost updates an option
func DisplayOneOptionPost(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	active := 0

	form := forms.New(r.PostForm, app.Database)
	if form.Has("active", r) {
		active = 1
	}

	o := clientmodels.Option{
		ID:         id,
		OptionName: form.Get("option_name"),
		Active:     active,
		UpdatedAt:  time.Now(),
		CreatedAt:  time.Now(),
	}

	if id > 0 {
		_ = vehicleModel.UpdateOption(o)
	} else {
		_ = vehicleModel.InsertOption(o)
	}

	session.Put(r.Context(), "flash", "Changes saved")
	http.Redirect(w, r, "/admin/inventory/options/all", http.StatusSeeOther)
}

// StaffAll lists all staff
func StaffAll(w http.ResponseWriter, r *http.Request) {
	rowSets := make(map[string]interface{})
	s, err := vehicleModel.GetStaff()
	if err != nil {
		errorLog.Println(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	rowSets["staff"] = s

	helpers.Render(w, r, "staff-all.page.tmpl", &templates.TemplateData{
		RowSets: rowSets,
	})
}

// SortStaff displays staff for sorting
func SortStaff(w http.ResponseWriter, r *http.Request) {
	rowSets := make(map[string]interface{})
	s, err := vehicleModel.GetStaffForSorting()
	if err != nil {
		errorLog.Println(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	rowSets["staff"] = s

	helpers.Render(w, r, "staff-sort.page.tmpl", &templates.TemplateData{
		RowSets: rowSets,
	})
}

// SortStaffPost sorts staff
func SortStaffPost(w http.ResponseWriter, r *http.Request) {
	sortList := r.Form.Get("sort_list")
	var sorted []SortOrder
	err := json.Unmarshal([]byte(sortList), &sorted)
	if err != nil {
		app.ErrorLog.Println(err)
	}

	for _, v := range sorted {
		imageID, _ := strconv.Atoi(v.ImageID)
		err := vehicleModel.UpdateSortOrderForStaff(imageID, v.StepNumber)
		if err != nil {
			app.ErrorLog.Println(err)
		}
	}

	session.Put(r.Context(), "flash", "Sort order saved")
	http.Redirect(w, r, "/admin/staff/sort-order", http.StatusSeeOther)
}

// DisplayOneStaff displays staff for add/edit
func DisplayOneStaff(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	rowSets := make(map[string]interface{})

	var o clientmodels.Employee

	if id > 0 {
		op, err := vehicleModel.GetOneStaff(id)
		if err != nil {
			errorLog.Println(err)
			helpers.ClientError(w, http.StatusBadRequest)
			return
		}
		o = op
	}
	rowSets["staff"] = o

	helpers.Render(w, r, "staff-one.page.tmpl", &templates.TemplateData{
		RowSets: rowSets,
		Form:    forms.New(nil),
	})
}

// DisplayOneStaffPost updates staff
func DisplayOneStaffPost(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	active := 0

	form := forms.New(r.PostForm, app.Database)
	if form.Has("active", r) {
		active = 1
	}

	var o clientmodels.Employee

	if id > 0 {
		o, _ = vehicleModel.GetOneStaff(id)
	} else {
		o.CreatedAt = time.Now()
	}

	o.FirstName = form.Get("first_name")
	o.LastName = form.Get("last_name")
	o.Email = form.Get("email")
	o.Position = form.Get("position")
	o.Description = form.Get("description")
	o.Active = active
	o.UpdatedAt = time.Now()

	if id > 0 {
		err := vehicleModel.UpdateStaff(o)
		if err != nil {
			errorLog.Println(err)
		}
	} else {
		newID, err := vehicleModel.InsertStaff(o)
		if err != nil {
			errorLog.Println(err)
		}
		id = newID
	}

	if form.HasFile("image", r) {
		// have an image
		fileName, fileNameDisplay, err := helpers.UploadOneFile(r, "./tmp/")
		if err != nil {
			errorLog.Println(err)
		} else {
			sourceDir := "./tmp"
			destDir := "./ui/static/site-content/staff"
			err = images.MakeThumbFromStaticFile(sourceDir, destDir, fileName, 600, 500)
			if err != nil {
				app.ErrorLog.Println("error making image")
			}
			oldLocation := fmt.Sprintf("./ui/static/site-content/staff/%s", fileName)
			newLocation := fmt.Sprintf("./ui/static/site-content/staff/%s", fileNameDisplay)
			err = MoveFile(oldLocation, newLocation)
			if err != nil {
				app.ErrorLog.Println("could not move from", oldLocation, "to", newLocation)
			}
			err = images.MakeThumbFromStaticFile(destDir, fmt.Sprintf("%s/thumbs/", destDir), fileNameDisplay, 100, 100)
			if err != nil {
				app.ErrorLog.Println("error making thumbnail")
			}

			s, _ := vehicleModel.GetOneStaff(id)
			s.Image = fileNameDisplay
			_ = vehicleModel.UpdateStaff(s)
		}
	}

	session.Put(r.Context(), "flash", "Changes saved")
	http.Redirect(w, r, "/admin/staff/all", http.StatusSeeOther)
}

// SalesPeopleAll lists all sales staff
func SalesPeopleAll(w http.ResponseWriter, r *http.Request) {
	rowSets := make(map[string]interface{})
	s, err := vehicleModel.GetSalesPeople()
	if err != nil {
		errorLog.Println(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	rowSets["staff"] = s

	helpers.Render(w, r, "sales-people-all.page.tmpl", &templates.TemplateData{
		RowSets: rowSets,
	})
}

// DisplayOneSalesStaff displays staff for add/edit
func DisplayOneSalesStaff(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	rowSets := make(map[string]interface{})

	var o clientmodels.SalesStaff

	if id > 0 {
		op, err := vehicleModel.GetOneSalesStaff(id)
		if err != nil {
			errorLog.Println(err)
			helpers.ClientError(w, http.StatusBadRequest)
			return
		}
		o = op
	}
	rowSets["staff"] = o

	helpers.Render(w, r, "sales-people-one.page.tmpl", &templates.TemplateData{
		RowSets: rowSets,
		Form:    forms.New(nil),
	})
}

// DisplayOneSalesStaffPost updates sales staff
func DisplayOneSalesStaffPost(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	active := 0

	form := forms.New(r.PostForm, app.Database)
	if form.Has("active", r) {
		active = 1
	}

	var o clientmodels.SalesStaff

	if id > 0 {
		o, _ = vehicleModel.GetOneSalesStaff(id)
	} else {
		o.CreatedAt = time.Now()
	}

	slugified := slug.Make(form.Get("name"))

	o.Name = form.Get("name")
	o.Slug = slugified
	o.Email = ""
	o.Phone = form.Get("phone")
	o.Active = active
	o.UpdatedAt = time.Now()

	if id > 0 {
		err := vehicleModel.UpdateSalesStaff(o)
		if err != nil {
			errorLog.Println(err)
		}
	} else {
		newID, err := vehicleModel.InsertSalesStaff(o)
		if err != nil {
			errorLog.Println(err)
		}
		infoLog.Println("Setting id to", id)
		id = newID
	}

	if form.HasFile("image", r) {
		infoLog.Println("Uploading image")
		// have an image
		fileName, fileNameDisplay, err := helpers.UploadOneFile(r, "./tmp/")
		if err != nil {
			errorLog.Println(err)
		} else {
			fmt.Println("in else")
			sourceDir := "./tmp"
			destDir := "./ui/static/site-content/salesstaff"
			err = images.MakeThumbFromStaticFile(sourceDir, destDir, fileName, 500, 1074)
			if err != nil {
				app.ErrorLog.Println("error making image")
			}
			oldLocation := fmt.Sprintf("./ui/static/site-content/salesstaff/%s", fileName)
			newLocation := fmt.Sprintf("./ui/static/site-content/salesstaff/%s", fileNameDisplay)
			err = MoveFile(oldLocation, newLocation)
			if err != nil {
				app.ErrorLog.Println("could not move from", oldLocation, "to", newLocation)
			}

			infoLog.Print("getting id of", id)
			s, _ := vehicleModel.GetOneSalesStaff(id)
			s.Image = fileNameDisplay
			err = vehicleModel.UpdateSalesStaff(s)
			if err != nil {
				errorLog.Println(err)
			}
		}
	}

	session.Put(r.Context(), "flash", "Changes saved")
	http.Redirect(w, r, "/admin/sales-people/all", http.StatusSeeOther)
}

// DeleteSalesPerson deletes sales
func DeleteSalesPerson(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	_ = vehicleModel.DeleteSalesStaff(id)
	session.Put(r.Context(), "flash", "Changes saved")
	http.Redirect(w, r, "/admin/sales-people/all", http.StatusSeeOther)
}

// DeleteStaff deletes a staff
func DeleteStaff(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	_ = vehicleModel.DeleteStaff(id)
	session.Put(r.Context(), "flash", "Changes saved")
	http.Redirect(w, r, "/admin/staff/all", http.StatusSeeOther)
}

// TestimonialsAllAdmin lists all testimonials in admin
func TestimonialsAllAdmin(w http.ResponseWriter, r *http.Request) {
	rowSets := make(map[string]interface{})
	s, err := vehicleModel.GetAllTestimonials()
	if err != nil {
		errorLog.Println(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	rowSets["staff"] = s

	helpers.Render(w, r, "testimonials-all.page.tmpl", &templates.TemplateData{
		RowSets: rowSets,
	})
}

// DisplayOneTestimonial displays staff for add/edit
func DisplayOneTestimonial(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	rowSets := make(map[string]interface{})

	var o clientmodels.Testimonial

	if id > 0 {
		op, err := vehicleModel.GetOneTestimonial(id)
		if err != nil {
			errorLog.Println(err)
			helpers.ClientError(w, http.StatusBadRequest)
			return
		}
		o = op
	}
	rowSets["testimonial"] = o

	helpers.Render(w, r, "testimonial-one.page.tmpl", &templates.TemplateData{
		RowSets: rowSets,
		Form:    forms.New(nil),
	})
}

// DisplayOneTestimonialPost posts a testimonial
func DisplayOneTestimonialPost(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	active := 0

	form := forms.New(r.PostForm, app.Database)
	if form.Has("active", r) {
		active = 1
	}

	var o clientmodels.Testimonial

	if id > 0 {
		o, _ = vehicleModel.GetOneTestimonial(id)
	} else {
		o.CreatedAt = time.Now()
	}

	o.Label = form.Get("label")
	o.Url = form.Get("url")
	o.Active = active
	o.UpdatedAt = time.Now()

	if id > 0 {
		err := vehicleModel.UpdateTestimonial(o)
		if err != nil {
			errorLog.Println(err)
		}
	} else {
		err := vehicleModel.InsertTestimonial(o)
		if err != nil {
			errorLog.Println(err)
		}
	}

	session.Put(r.Context(), "flash", "Changes saved")
	http.Redirect(w, r, "/admin/testimonials/all", http.StatusSeeOther)
}

// TestimonialsAllAdmin lists all testimonials in admin
func WordAllAdmin(w http.ResponseWriter, r *http.Request) {
	rowSets := make(map[string]interface{})
	s, err := vehicleModel.GetAllWordOfMouth()
	if err != nil {
		errorLog.Println(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	rowSets["staff"] = s

	helpers.Render(w, r, "word-all.page.tmpl", &templates.TemplateData{
		RowSets: rowSets,
	})
}

// DisplayOneWord displays one
func DisplayOneWord(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	rowSets := make(map[string]interface{})

	var o clientmodels.Word

	if id > 0 {
		op, err := vehicleModel.GetOneWordOfMouth(id)
		if err != nil {
			errorLog.Println(err)
			helpers.ClientError(w, http.StatusBadRequest)
			return
		}
		o = op
	}
	rowSets["testimonial"] = o

	helpers.Render(w, r, "word-one.page.tmpl", &templates.TemplateData{
		RowSets: rowSets,
		Form:    forms.New(nil),
	})
}

// DisplayOneWordPost posts word of mouth
func DisplayOneWordPost(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	active := 0

	form := forms.New(r.PostForm, app.Database)
	if form.Has("active", r) {
		active = 1
	}

	var o clientmodels.Word

	if id > 0 {
		o, _ = vehicleModel.GetOneWordOfMouth(id)
	} else {
		o.CreatedAt = time.Now()
	}

	o.Title = form.Get("title")
	o.Content = form.Get("content")
	o.Active = active
	o.UpdatedAt = time.Now()

	if id > 0 {
		err := vehicleModel.UpdateWordOfMouth(o)
		if err != nil {
			errorLog.Println(err)
		}
	} else {
		err := vehicleModel.InsertWordOfMouth(o)
		if err != nil {
			errorLog.Println(err)
		}
	}

	session.Put(r.Context(), "flash", "Changes saved")
	http.Redirect(w, r, "/admin/testimonials/word-of-mouth/all", http.StatusSeeOther)
}

func AllFinders(w http.ResponseWriter, r *http.Request) {
	rowSets := make(map[string]interface{})
	s, err := vehicleModel.GetAllFinders()
	if err != nil {
		errorLog.Println(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}
	rowSets["finders"] = s

	helpers.Render(w, r, "finder-all.page.tmpl", &templates.TemplateData{
		RowSets: rowSets,
	})
}

// DisplayOneFinder displays staff for add/edit
func DisplayOneFinder(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	rowSets := make(map[string]interface{})

	var o clientmodels.Finder

	if id > 0 {
		op, err := vehicleModel.GetOneFinder(id)
		if err != nil {
			errorLog.Println(err)
			helpers.ClientError(w, http.StatusBadRequest)
			return
		}
		o = op
	}
	rowSets["app"] = o

	helpers.Render(w, r, "finder-one.page.tmpl", &templates.TemplateData{
		RowSets: rowSets,
		Form:    forms.New(nil),
	})
}
