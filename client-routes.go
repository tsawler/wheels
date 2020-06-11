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

	// Vehicle Administration
	mux.Get("/admin/inventory/all-vehicles", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllVehicles))
	mux.Post("/admin/inventory/all-vehicles-json", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllVehiclesJSON))
	mux.Get("/admin/inventory/vehicle/:ID", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(DisplayVehicleForAdmin))

	return mux, nil
}
