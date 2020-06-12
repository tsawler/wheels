package clienthandlers

import (
	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
	mw "github.com/tsawler/goblender/pkg/middleware"
	"net/http"
)

// ClientRoutes is used to handle custom routes for specific clients. Prepend some unique (and site wide) value
// to the start of each route in order to avoid clashes with pages, etc. Middleware can be applied by importing and
// using the middleware.* functions.
func ClientRoutes(mux *pat.PatternServeMux, standardMiddleWare, dynamicMiddleware alice.Chain) (*pat.PatternServeMux, error) {
	// public folder
	fileServer := http.FileServer(http.Dir("./client/clienthandlers/public/"))
	mux.Get("/client/static/", http.StripPrefix("/client/static", fileServer))

	// webhook from tusd
	mux.Post("/tusd/hook", standardMiddleWare.ThenFunc(TusWebHook(app)))

	// Vehicle Administration
	mux.Get("/admin/inventory/vehicles/all-vehicles", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllVehicles))
	mux.Post("/admin/inventory/all-vehicles-json", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllVehiclesJSON))
	mux.Get("/admin/:CATEGORY/:TYPE/:SRC/:ID", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(DisplayVehicleForAdmin))

	mux.Get("/admin/inventory/vehicles/all-vehicles-for-sale", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllVehiclesForSale))
	mux.Post("/admin/inventory/all-vehicles-for-sale-json", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllVehiclesForSaleJSON))

	mux.Get("/admin/inventory/vehicles/all-vehicles-sold", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllSold))
	mux.Post("/admin/inventory/all-vehicles-sold-json", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllSoldJSON))

	mux.Get("/admin/inventory/vehicles/all-vehicles-sold-this-month", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllSoldThisMonth))
	mux.Post("/admin/inventory/all-vehicles-sold-this-month-json", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllSoldThisMonthJSON))

	mux.Get("/admin/powersports-inventory/powersports/all-powersports-for-sale", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllPowerSportsForSale))
	mux.Post("/admin/powersports/all-powersports-for-sale-json", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllPowerSportsForSaleJSON))

	mux.Get("/admin/powersports-inventory/powersports/all-powersports-sold", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllPowerSportsSold))
	mux.Post("/admin/powersports/all-powersports-sold-json", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllPowerSportsSoldJSON))

	mux.Get("/admin/powersports-inventory/powersports/all-powersports-sold-this-month", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllPowerSportsSoldThisMonth))
	mux.Post("/admin/powersports/all-powersports-sold-this-month-json", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllPowerSportsSoldThisMonthJSON))

	mux.Get("/admin/inventory/vehicles/all-vehicles-pending", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllVehiclesPending))
	mux.Post("/admin/inventory/all-vehicles-pending-json", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllVehiclesPendingJSON))

	return mux, nil
}
