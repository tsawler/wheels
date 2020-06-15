package clienthandlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tsawler/goblender/client/clienthandlers/clientmodels"
	"github.com/tsawler/goblender/pkg/datatables"
	"github.com/tsawler/goblender/pkg/forms"
	"github.com/tsawler/goblender/pkg/helpers"
	"github.com/tsawler/goblender/pkg/templates"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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

	pdf, err := CreateWindowSticker(vehicleID)
	if err != nil {
		lastPage := app.Session.GetString(r.Context(), "last-page")
		session.Put(r.Context(), "error", "Unable to generate PDF!")
		http.Redirect(w, r, lastPage, http.StatusSeeOther)
		return
	}

	//w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.pdf", v.StockNo))

	out := &bytes.Buffer{}
	if err := pdf.Output(out); err != nil {
		lastPage := app.Session.GetString(r.Context(), "last-page")
		session.Put(r.Context(), "error", "Unable to generate PDF!")
		http.Redirect(w, r, lastPage, http.StatusSeeOther)
		return
	}
	b := out.Bytes()

	_, err = w.Write(b)
	if err != nil {
		errorLog.Println(err)
	}

}
