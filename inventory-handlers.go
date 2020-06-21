package clienthandlers

import (
	"encoding/json"
	"fmt"
	"github.com/tsawler/goblender/pkg/helpers"
	"github.com/tsawler/goblender/pkg/templates"
	"net/http"
	"strconv"
)

const (
	SOLD    = 0
	FORSALE = 1
	PENDING = 2
	TRADEIN = 3
)

// DisplayAllVehicleInventory shows all vehicle inventory
func DisplayAllVehicleInventory(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["pager-url"] = "/used-vehicle-inventory"
	intMap := make(map[string]int)
	intMap["show-makes"] = 1
	vehicleType := All
	templateName := "inventory.page.tmpl"

	renderInventory(r, stringMap, vehicleType, w, intMap, templateName, "used-vehicle-inventory", false)
}

// DisplaySUVInventory shows suv inventory
func DisplaySUVInventory(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["pager-url"] = "/used-suvs-fredericton"
	intMap := make(map[string]int)
	intMap["show-makes"] = 1
	vehicleType := SUV
	templateName := "inventory.page.tmpl"

	renderInventory(r, stringMap, vehicleType, w, intMap, templateName, "used-suvs-fredericton", false)
}

// DisplayCarInventory shows car inventory
func DisplayCarInventory(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["pager-url"] = "/used-cars-fredericton"
	intMap := make(map[string]int)
	intMap["show-makes"] = 1
	vehicleType := Car
	templateName := "inventory.page.tmpl"

	renderInventory(r, stringMap, vehicleType, w, intMap, templateName, "used-cars-fredericton", false)
}

// DisplayTruckInventory shows truck inventory
func DisplayTruckInventory(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["pager-url"] = "/used-trucks-fredericton"
	intMap := make(map[string]int)
	intMap["show-makes"] = 1
	vehicleType := Truck
	templateName := "inventory.page.tmpl"

	renderInventory(r, stringMap, vehicleType, w, intMap, templateName, "used-trucks-fredericton", false)
}

// DisplayMinivanInventory shows minivan inventory
func DisplayMinivanInventory(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["pager-url"] = "/used-minivans-fredericton"
	intMap := make(map[string]int)
	intMap["show-makes"] = 1
	vehicleType := MiniVan
	templateName := "inventory.page.tmpl"

	renderInventory(r, stringMap, vehicleType, w, intMap, templateName, "used-minivans-fredericton", false)
}

// MVI Select shows budget priced used cars inventory
func DisplayMVISelect(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["pager-url"] = "/budget-priced-used-cars"
	intMap := make(map[string]int)
	intMap["show-makes"] = 1
	vehicleType := All
	templateName := "inventory.page.tmpl"

	renderInventory(r, stringMap, vehicleType, w, intMap, templateName, "budget-priced-used-cars", true)
}

// renderInventory renders inventory for a product type
func renderInventory(r *http.Request, stringMap map[string]string, vehicleType int, w http.ResponseWriter, intMap map[string]int, templateName, slug string, handPicked bool) {
	var offset int
	var selectedYear, selectedMake, selectedModel, selectedPrice int
	pagerSuffix := ""
	stringMap["item-link-prefix"] = "view"

	pageIndex, err := strconv.Atoi(r.URL.Query().Get(":pageIndex"))
	if err != nil {
		pageIndex = 1
	}

	searching, ok := r.URL.Query()["year"]
	if !ok || len(searching[0]) < 1 {
		selectedYear = 0
		selectedMake = 0
		selectedModel = 0
		selectedPrice = 0
	} else {
		selectedYear, _ = strconv.Atoi(r.URL.Query()["year"][0])
		selectedMake, _ = strconv.Atoi(r.URL.Query()["make"][0])
		selectedModel, _ = strconv.Atoi(r.URL.Query()["model"][0])
		selectedPrice, _ = strconv.Atoi(r.URL.Query()["price"][0])
		pagerSuffix = fmt.Sprintf("?year=%d&make=%d&model=%d&price=%d", selectedYear, selectedMake, selectedModel, selectedPrice)
		stringMap["pager-suffix"] = pagerSuffix
	}

	perPage := 10
	offset = (pageIndex - 1) * perPage

	vehicles, num, err := vehicleModel.AllVehiclesPaginated(vehicleType, perPage, offset, selectedYear, selectedMake, selectedModel, selectedPrice, handPicked)
	if err != nil {
		errorLog.Println(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

	rowSets := make(map[string]interface{})
	rowSets["vehicles"] = vehicles

	intMap["num-vehicles"] = num
	intMap["current-page"] = pageIndex
	intMap["year"] = selectedYear
	intMap["make"] = selectedMake
	intMap["model"] = selectedModel
	intMap["price"] = selectedPrice
	intMap["vehicle-type"] = vehicleType

	pg, err := repo.DB.GetPageBySlug(slug)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// get makes
	makes, err := vehicleModel.GetMakesForVehicleType(vehicleType)
	if err != nil {
		errorLog.Println(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

	// get models
	models, err := vehicleModel.GetModelsForVehicleType(vehicleType)
	if err != nil {
		errorLog.Println(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

	// get years
	years, err := vehicleModel.GetYearsForVehicleType(vehicleType)
	if err != nil {
		errorLog.Println(err)
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

	rowSets["years"] = years
	rowSets["models"] = models
	rowSets["makes"] = makes

	helpers.Render(w, r, templateName, &templates.TemplateData{
		Page:      pg,
		RowSets:   rowSets,
		IntMap:    intMap,
		StringMap: stringMap,
	})
}

// GetModelsForMake gets models for make
func GetModelsForMake(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	vehicleTypeID, _ := strconv.Atoi(r.URL.Query().Get(":type"))
	models, err := vehicleModel.ModelsForMakeID(id, vehicleTypeID)
	if err != nil {
		return
	}

	out, err := json.MarshalIndent(models, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// GetModelsForMakeAdmin gets models for make
func GetModelsForMakeAdmin(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	models, err := vehicleModel.ModelsForMakeIDAdmin(id)
	if err != nil {
		return
	}

	out, err := json.MarshalIndent(models, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// GetMakesForYear gets makes for year
func GetMakesForYear(w http.ResponseWriter, r *http.Request) {
	year, _ := strconv.Atoi(r.URL.Query().Get(":YEAR"))
	vehicleTypeID, _ := strconv.Atoi(r.URL.Query().Get(":type"))
	makes, err := vehicleModel.MakesForYear(year, vehicleTypeID)
	if err != nil {
		return
	}

	out, err := json.MarshalIndent(makes, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)

}

// DisplayOneVehicle shows one vehicle
func DisplayOneVehicle(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))

	pg, err := repo.DB.GetPageBySlug("display-one-item")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	pg.PageNotEditable = 1

	item, err := vehicleModel.GetVehicleByID(id)
	if err != nil {
		fmt.Fprint(w, "custom 404")
		return
	}

	if item.Status != FORSALE {
		// item is sold, or whatever
		http.Redirect(w, r, "/used-vehicle-inventory", http.StatusMovedPermanently)
		return
	}

	rowSets := make(map[string]interface{})
	rowSets["item"] = item

	staff, err := vehicleModel.GetSales()
	if err != nil {
		errorLog.Println(err)
	}

	rowSets["sales"] = staff

	helpers.Render(w, r, "item.page.tmpl", &templates.TemplateData{
		Page:    pg,
		RowSets: rowSets,
	})
}
