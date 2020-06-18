package clienthandlers

import (
	"github.com/tsawler/goblender/client/clienthandlers/clientdb"
	template_data "github.com/tsawler/goblender/client/clienthandlers/template-data"
	"github.com/tsawler/goblender/pkg/config"
	"github.com/tsawler/goblender/pkg/driver"
	"github.com/tsawler/goblender/pkg/handlers"
	"github.com/tsawler/goblender/pkg/helpers"
	"log"
)

var app config.AppConfig
var infoLog *log.Logger
var errorLog *log.Logger
var repo *handlers.DBRepo
var vehicleModel *clientdb.DBModel

// ClientInit gives client code access to goBlender configuration
func ClientInit(conf config.AppConfig, parentDriver *driver.DB, rep *handlers.DBRepo) {
	// make sure the directories we need are there
	_ = helpers.CreateDirIfNotExist("./ui/static/site-content/inventory/")
	_ = helpers.CreateDirIfNotExist("./ui/static/site-content/panoramas/")
	_ = helpers.CreateDirIfNotExist("./ui/static/site-content/salesstaff/")
	_ = helpers.CreateDirIfNotExist("./ui/static/site-content/staff/")
	_ = helpers.CreateDirIfNotExist("./ui/static/site-content/staff/thumbs")

	// conf is the application config, from goBlender
	app = conf
	repo = rep

	// If we have additional databases (external to this application) we set the connection here.
	// The connection is specified in goBlender preferences.
	//conn := app.AlternateConnection

	// loggers
	infoLog = app.InfoLog
	errorLog = app.ErrorLog

	// We can access handlers from goBlender, but need to initialize them first.
	if app.Database == "postgresql" {
		handlers.NewPostgresqlHandlers(parentDriver, app.ServerName, app.InProduction)
	} else {
		handlers.NewMysqlHandlers(parentDriver, app.ServerName, app.InProduction)
	}

	// Set a different template for home page, if needed.
	//repo.SetHomePageTemplate("client-sample.page.tmpl")

	// Set a different template for inside pages, if needed.
	//repo.SetDefaultPageTemplate("client-sample.page.tmpl")

	vehicleModel = &clientdb.DBModel{DB: parentDriver.SQL}

	// Create client middleware
	NewClientMiddleware(app)
	template_data.NewTemplateData(parentDriver.SQL)

	infoLog.Println("Scheduling PBS for every 3 hours....")
	_, _ = app.Scheduler.AddFunc("@every 3h", func() {
		PullFromPBS()
	})

}
