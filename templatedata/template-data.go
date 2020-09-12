package templatedata

import (
	"database/sql"
	"github.com/tsawler/goblender/client/clienthandlers/clientdb"
	"github.com/tsawler/goblender/pkg/templates"
)

var vehicleModel *clientdb.DBModel

func NewTemplateData(p *sql.DB) {
	vehicleModel = &clientdb.DBModel{DB: p}
}

// AddDefaultData
func AddDefaultData(td *templates.TemplateData) *templates.TemplateData {
	data := make(map[string]interface{})

	soldThisMonth := vehicleModel.CountSoldThisMonth()
	data["sold_this_month"] = soldThisMonth

	pending := vehicleModel.CountPending()
	data["pending"] = pending

	tradeIns := vehicleModel.CountTradeIns()
	data["trade_ins"] = tradeIns

	forSale := vehicleModel.CountForSale()
	data["for_sale"] = forSale

	forSalePowerSports := vehicleModel.CountForSalePowerSports()
	data["for_sale_powersports"] = forSalePowerSports

	soldThisMonthPowerSports := vehicleModel.CountSoldThisMonthPowerSports()
	data["sold_this_month_powersports"] = soldThisMonthPowerSports

	td.Data = data
	return td
}
