package clienthandlers

import (
	"encoding/json"
	"fmt"
	"github.com/tsawler/goblender/pkg/helpers"
	"github.com/tsawler/goblender/pkg/templates"
	"net/http"
	"strconv"
)

func DisplayAllVehicleInventory(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["pager-url"] = "/used-vehicle-inventory"
	intMap := make(map[string]int)
	intMap["show-makes"] = 1
	vehicleType := 0
	templateName := "inventory.page.tmpl"

	renderInventory(r, stringMap, vehicleType, w, intMap, templateName, "used-vehicle-inventory")
}

func DisplaySUVInventory(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["pager-url"] = "/used-suvs-fredericton"
	intMap := make(map[string]int)
	intMap["show-makes"] = 1
	vehicleType := 5
	templateName := "inventory.page.tmpl"

	renderInventory(r, stringMap, vehicleType, w, intMap, templateName, "used-vehicle-inventory")
}

// renderInventory renders inventory for a product type
func renderInventory(r *http.Request, stringMap map[string]string, vehicleType int, w http.ResponseWriter, intMap map[string]int, templateName, slug string) {
	var offset int
	var selectedYear, selectedMake, selectedModel, selectedPrice int
	pagerSuffix := ""
	stringMap["item-link-prefix"] = "view"
	stringMap["pager-prefix"] = "powersports-inventory"

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

	vehicles, num, err := vehicleModel.AllVehiclesPaginated(vehicleType, perPage, offset, selectedYear, selectedMake, selectedModel, selectedPrice)
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

func GetModelsForMake(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	models, err := vehicleModel.ModelsForMakeID(id)
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

func GetMakesForYear(w http.ResponseWriter, r *http.Request) {
	year, _ := strconv.Atoi(r.URL.Query().Get(":YEAR"))
	makes, err := vehicleModel.MakesForYear(year)
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

func DisplayOneVehicle(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":ID"))
	infoLog.Println(id)
	category := r.URL.Query().Get(":CATEGORY")
	infoLog.Println(category)

	v, err := vehicleModel.GetVehicleByID(id)
	if err != nil {
		errorLog.Println(err)
		helpers.ClientError(w, http.StatusBadRequest)
	}
	w.Write([]byte(v.Description))
}
