package clienthandlers

import (
	"encoding/json"
	"fmt"
	"github.com/tsawler/goblender/client/clienthandlers/clientmodels"
	"github.com/tsawler/goblender/pkg/datatables"
	"github.com/tsawler/goblender/pkg/forms"
	"github.com/tsawler/goblender/pkg/helpers"
	"github.com/tsawler/goblender/pkg/templates"
	"net/http"
	"strconv"
	"time"
)

// DataTablesJSON holds the json for datatables
type DataTablesJSON struct {
	Draw            int64                       `json:"draw"`
	RecordsTotal    int64                       `json:"recordsTotal"`
	RecordsFiltered int64                       `json:"recordsFiltered"`
	DataRows        []*clientmodels.VehicleJSON `json:"data"`
}

func AllVehicles(w http.ResponseWriter, r *http.Request) {
	helpers.Render(w, r, "all-vehicles.page.tmpl", &templates.TemplateData{})
}

func AllVehiclesForSale(w http.ResponseWriter, r *http.Request) {
	helpers.Render(w, r, "all-vehicles-for-sale.page.tmpl", &templates.TemplateData{})
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
	v, rowCount, filterCount, err := vehicleModel.VehicleJSON(query, baseQuery, "where vehicle_type < 7")
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
	v, rowCount, filterCount, err := vehicleModel.VehicleJSON(query, baseQuery, "where vehicle_status = 1 and vehicle_type <  7")
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

// DisplayVehicleForAdmin shows vehicle for edit
func DisplayVehicleForAdmin(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get(":ID"))
	if err != nil {
		errorLog.Println(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

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
		RowSets: rowSets,
		IntMap:  intMap,
		Form:    forms.New(nil),
	})

}
