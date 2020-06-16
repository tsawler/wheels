package clienthandlers

import (
	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
	mw "github.com/tsawler/goblender/pkg/middleware"
	"net/http"
)

// ClientRoutes are the client specific routes
func ClientRoutes(mux *pat.PatternServeMux, standardMiddleWare, dynamicMiddleware alice.Chain) (*pat.PatternServeMux, error) {
	// client public folder
	fileServer := http.FileServer(http.Dir("./client/clienthandlers/public/"))
	mux.Get("/client/static/", http.StripPrefix("/client/static", fileServer))

	/*
		|--------------------------------------------------------------------------
		| TUS web hook
		|--------------------------------------------------------------------------
		| Web hook for tusd (overrides default handler in goBlender).
		| Handles uploads of images, files, videos for standard goBlender
		| functionality, and additional functionality specific to this site.
		| Authentication is handled the same way it is for goBlender.
		|
	*/
	mux.Post("/tusd/hook", standardMiddleWare.ThenFunc(TusWebHook(app)))

	/*
		|--------------------------------------------------------------------------
		| Public Routes // TODO
		|--------------------------------------------------------------------------
		|
	*/

	// credit app
	mux.Get("/credit-application", standardMiddleWare.ThenFunc(CreditApp))
	mux.Post("/credit-application", standardMiddleWare.ThenFunc(PostCreditApp))

	// inventory filters
	mux.Get("/inventory-filter/makes/:YEAR", dynamicMiddleware.ThenFunc(GetMakesForYear))
	mux.Get("/inventory-filter/models/:ID", dynamicMiddleware.ThenFunc(GetModelsForMake))

	// all used vehicles
	mux.Get("/used-vehicle-inventory", dynamicMiddleware.ThenFunc(DisplayAllVehicleInventory))
	mux.Get("/used-vehicle-inventory/:pageIndex", dynamicMiddleware.ThenFunc(DisplayAllVehicleInventory))

	// suvs
	mux.Get("/used-suvs-fredericton", dynamicMiddleware.ThenFunc(DisplaySUVInventory))
	mux.Get("/used-suvs-fredericton/:pageIndex", dynamicMiddleware.ThenFunc(DisplaySUVInventory))

	// cars
	mux.Get("/used-cars-fredericton", dynamicMiddleware.ThenFunc(DisplayCarInventory))
	mux.Get("/used-cars-fredericton/:pageIndex", dynamicMiddleware.ThenFunc(DisplayCarInventory))

	// trucks
	mux.Get("/used-trucks-fredericton", dynamicMiddleware.ThenFunc(DisplayTruckInventory))
	mux.Get("/used-trucks-fredericton/:pageIndex", dynamicMiddleware.ThenFunc(DisplayTruckInventory))

	// minivans
	mux.Get("/used-minivans-fredericton", dynamicMiddleware.ThenFunc(DisplayMinivanInventory))
	mux.Get("/used-minivans-fredericton/:pageIndex", dynamicMiddleware.ThenFunc(DisplayMinivanInventory))

	mux.Get("/:CATEGORY/view/:ID/:SLUG", dynamicMiddleware.ThenFunc(DisplayOneVehicle))
	/*
		|--------------------------------------------------------------------------
		| Vehicle Administration routes
		|--------------------------------------------------------------------------
		| These routes require authentication and a specific role assigned to a
		| user before they can be accessed. Any attempt to access them without the
		| proper authentication/role results in an "Unauthorized" http response.
		|
	*/

	// json for vehicle admin
	mux.Post("/admin/vehicle-images-json/:ID", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(VehicleImagesJSON))
	mux.Post("/admin/delete-vehicle-image-json/:ID", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(VehicleImageDelete))

	// pbs update
	mux.Get("/admin/inventory/refresh-from-pbs", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(RefreshFromPBS))

	// print window sticker
	mux.Get("/admin/inventory/print-window-sticker/:ID", dynamicMiddleware.ThenFunc(PrintWindowSticker))

	// display and edit vehicle/item
	mux.Get("/admin/:CATEGORY/:TYPE/:SRC/:ID", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(DisplayVehicleForAdmin))
	mux.Post("/admin/:CATEGORY/:TYPE/:SRC/:ID", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(DisplayVehicleForAdminPost))

	// all cars/trucks
	mux.Get("/admin/inventory/vehicles/all-vehicles", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllVehicles))
	mux.Post("/admin/inventory/all-vehicles-json", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllVehiclesJSON))

	//		all cars/trucks for sale
	mux.Get("/admin/inventory/vehicles/all-vehicles-for-sale", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllVehiclesForSale))
	mux.Post("/admin/inventory/all-vehicles-for-sale-json", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllVehiclesForSaleJSON))

	// all cars/trucks sold
	mux.Get("/admin/inventory/vehicles/all-vehicles-sold", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllSold))
	mux.Post("/admin/inventory/all-vehicles-sold-json", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllSoldJSON))

	// all cars/trucks sold this month
	mux.Get("/admin/inventory/vehicles/all-vehicles-sold-this-month", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllSoldThisMonth))
	mux.Post("/admin/inventory/all-vehicles-sold-this-month-json", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllSoldThisMonthJSON))

	// all power sports for sale
	mux.Get("/admin/powersports-inventory/powersports/all-powersports-for-sale", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllPowerSportsForSale))
	mux.Post("/admin/powersports/all-powersports-for-sale-json", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllPowerSportsForSaleJSON))

	// all power sports sold
	mux.Get("/admin/powersports-inventory/powersports/all-powersports-sold", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllPowerSportsSold))
	mux.Post("/admin/powersports/all-powersports-sold-json", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllPowerSportsSoldJSON))

	// all power sports sold this month
	mux.Get("/admin/powersports-inventory/powersports/all-powersports-sold-this-month", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllPowerSportsSoldThisMonth))
	mux.Post("/admin/powersports/all-powersports-sold-this-month-json", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllPowerSportsSoldThisMonthJSON))

	// all vehicles pending
	mux.Get("/admin/inventory/vehicles/all-vehicles-pending", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllVehiclesPending))
	mux.Post("/admin/inventory/all-vehicles-pending-json", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllVehiclesPendingJSON))

	// all trade ins
	mux.Get("/admin/inventory/vehicles/all-vehicles-trade-ins", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllVehiclesTradeIns))
	mux.Post("/admin/inventory/all-vehicles-trade-ins-json", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(AllVehiclesTradeInsJSON))

	return mux, nil
}
