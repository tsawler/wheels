package clienthandlers

import (
	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
	"net/http"
)

// ClientRoutes is used to handle custom routes for specific clients. Prepend some unique (and site wide) value
// to the start of each route in order to avoid clashes with pages, etc. Middleware can be applied by importing and
// using the middleware.* functions.
func ClientRoutes(mux *pat.PatternServeMux, standardMiddleWare, dynamicMiddleware alice.Chain) (*pat.PatternServeMux, error) {

	// example of overriding standard route
	//mux.Get("/", dynamicMiddleware.ThenFunc(CustomShowHome))

	// we can use any of the handlers in goBlender, e.g.
	//mux.Get("/client/yellow/submarine", standardMiddleWare.ThenFunc(handlers.Repo.ShowGalleryPage(app)))

	// this route requires both a goBlender middleware, and a custom client middleware
	//mux.Get("/client/some-handler", standardMiddleWare.Append(mw.Auth).Append(SomeRole).ThenFunc(SomeHandler))

	// public folder
	fileServer := http.FileServer(http.Dir("./client/clienthandlers/public/"))
	mux.Get("/client/static/", http.StripPrefix("/client/static", fileServer))

	mux.Get("/all-vehicles", dynamicMiddleware.ThenFunc(AllVehicles))

	return mux, nil
}
