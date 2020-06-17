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

	/*--------------------------------------------------------------------------
	| TUS web hook
	|--------------------------------------------------------------------------
	| Web hook for tusd (overrides default handler in goBlender).
	| Handles uploads of images, files, videos for standard goBlender
	| functionality, and additional functionality specific to this site.
	| Authentication is handled the same way it is for goBlender. */

	mux.Post("/tusd/hook", standardMiddleWare.ThenFunc(TusWebHook(app)))

	/*--------------------------------------------------------------------------
	| Public Routes
	|--------------------------------------------------------------------------*/

	// credit app
	mux.Get("/credit-application", standardMiddleWare.ThenFunc(CreditApp))
	mux.Post("/credit-application", standardMiddleWare.ThenFunc(PostCreditApp))

	mux.Post("/inventory/compare-vehicles", standardMiddleWare.ThenFunc(CompareVehicles))
	mux.Post("/wheels/quick-quote", standardMiddleWare.ThenFunc(QuickQuote))
	mux.Post("/wheels/test-drive", standardMiddleWare.ThenFunc(TestDrive))
	mux.Post("/wheels/send-to-friend", standardMiddleWare.ThenFunc(SendFriend))

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

	// show vehicle
	mux.Get("/:CATEGORY/view/:ID/:SLUG", dynamicMiddleware.ThenFunc(DisplayOneVehicle))

	/*--------------------------------------------------------------------------
	| Vehicle Administration routes
	|--------------------------------------------------------------------------
	| These routes require authentication and a specific role assigned to a
	| user before they can be accessed. Any attempt to access them without the
	| proper authentication/role results in an "Unauthorized" http response. */

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

	// options
	mux.Get("/admin/inventory/options/all", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(OptionsAll))
	mux.Get("/admin/inventory/options/:ID", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(DisplayOneOption))
	mux.Post("/admin/inventory/options/:ID", dynamicMiddleware.Append(mw.Auth).Append(InventoryRole).ThenFunc(DisplayOneOptionPost))

	/*---------------------------------------------------------------------------
	| Staff
	|--------------------------------------------------------------------------*/

	mux.Get("/admin/staff/all", dynamicMiddleware.Append(mw.Auth).Append(StaffRole).ThenFunc(StaffAll))
	mux.Get("/admin/staff/:ID", dynamicMiddleware.Append(mw.Auth).Append(StaffRole).ThenFunc(DisplayOneStaff))
	mux.Post("/admin/staff/:ID", dynamicMiddleware.Append(mw.Auth).Append(StaffRole).ThenFunc(DisplayOneStaffPost))

	/*--------------------------------------------------------------------------
	| Credit Applications, test drives, quick quotes
	|--------------------------------------------------------------------------*/

	mux.Get("/admin/credit/all", dynamicMiddleware.Append(mw.Auth).Append(CreditRole).ThenFunc(AllCreditApplications))
	mux.Get("/admin/credit/application/:ID", dynamicMiddleware.Append(mw.Auth).Append(CreditRole).ThenFunc(OneCreditApp))
	mux.Post("/admin/credit/all-credit-apps-json", dynamicMiddleware.Append(mw.Auth).Append(CreditRole).ThenFunc(AllCreditAppsJSON))

	mux.Get("/admin/credit/all-quick-quotes", dynamicMiddleware.Append(mw.Auth).Append(CreditRole).ThenFunc(AllQuickQuotes))
	mux.Get("/admin/credit/quick-quote/:ID", dynamicMiddleware.Append(mw.Auth).Append(CreditRole).ThenFunc(OneQuickQuote))
	mux.Post("/admin/credit/all-quick-quotes-json", dynamicMiddleware.Append(mw.Auth).Append(CreditRole).ThenFunc(AllQuickQuotesJSON))

	mux.Get("/admin/test-drives/all", dynamicMiddleware.Append(mw.Auth).Append(TestDriveRole).ThenFunc(AllTestDrives))
	mux.Get("/admin/test-drives/:ID", dynamicMiddleware.Append(mw.Auth).Append(TestDriveRole).ThenFunc(OneTestDrive))
	mux.Post("/admin/credit/all-test-drives-json", dynamicMiddleware.Append(mw.Auth).Append(TestDriveRole).ThenFunc(AllTestDrivesJSON))

	return mux, nil
}
