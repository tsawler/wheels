package clienthandlers

import (
	"fmt"
	"github.com/tsawler/goblender/client/clienthandlers/clientmodels"
	"github.com/tsawler/goblender/pkg/config"
	"github.com/tsawler/goblender/pkg/images"
	"time"
)

type VehicleImageProcessingJob struct {
	Image VehicleImageData
}

type VehicleImageData struct {
	SourceDir   string
	DestDir     string
	Slugified   string
	OldLocation string
	Height      int
	Width       int
	SortOrder   int
	VehicleID   int
	UserID      int
}

// NewVehicleImageManagerWorker takes a numeric id and a channel w/ worker pool.
func NewVehicleImageManagerWorker(id int, workerPool chan chan VehicleImageProcessingJob) VehicleImageWorker {
	return VehicleImageWorker{
		id:         id,
		jobQueue:   make(chan VehicleImageProcessingJob),
		workerPool: workerPool,
		quitChan:   make(chan bool),
	}
}

// VehicleImageWorker holds info for a pool worker
type VehicleImageWorker struct {
	id         int
	jobQueue   chan VehicleImageProcessingJob
	workerPool chan chan VehicleImageProcessingJob
	quitChan   chan bool
}

// start starts the worker
func (w VehicleImageWorker) start() {
	go func() {
		for {
			// Add jobQueue to the worker pool.
			w.workerPool <- w.jobQueue

			select {
			case job := <-w.jobQueue:
				w.processImageJob(job.Image)
			case <-w.quitChan:
				fmt.Printf("worker%d stopping\n", w.id)
				return
			}
		}
	}()
}

// stop the worker
func (w VehicleImageWorker) stop() {
	go func() {
		w.quitChan <- true
	}()
}

// NewVehicleImageDispatcher creates, and returns a new Dispatcher object.
func NewVehicleImageDispatcher(jobQueue chan VehicleImageProcessingJob, maxWorkers int) *VehicleImageDispatcher {
	workerPool := make(chan chan VehicleImageProcessingJob, maxWorkers)

	return &VehicleImageDispatcher{
		jobQueue:   jobQueue,
		maxWorkers: maxWorkers,
		workerPool: workerPool,
	}
}

// VehicleImageDispatcher holds info for a dispatcher
type VehicleImageDispatcher struct {
	workerPool chan chan VehicleImageProcessingJob
	maxWorkers int
	jobQueue   chan VehicleImageProcessingJob
	app        config.AppConfig
}

// run runs the workers
func (d *VehicleImageDispatcher) run() {
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewVehicleImageManagerWorker(i+1, d.workerPool)
		worker.start()
	}

	go d.dispatch()
}

// dispatch dispatches worker
func (d *VehicleImageDispatcher) dispatch() {
	for {
		select {
		case job := <-d.jobQueue:
			go func() {
				workerJobQueue := <-d.workerPool
				workerJobQueue <- job
			}()
		}
	}
}

// imageJob processes the main queue job
func (w VehicleImageWorker) processImageJob(i VehicleImageData) {
	app.InfoLog.Println("HIt process image job")
	processUploadedImage(i.VehicleID, i.SortOrder, i.SourceDir, i.DestDir, i.Slugified)

	// send notification via websocket
	payload := "done"
	data := map[string]string{"message": payload}

	err := app.Client.Trigger(fmt.Sprintf("private-channel-%s-%d", app.Identifier, i.UserID), "vehicle-image-manager-upload-event", data)
	if err != nil {
		errorLog.Println(err)
	}
}

func processUploadedImage(vehicleID, so int, sourceDir, destDir, slugified string) {
	infoLog.Println("Processing image:", vehicleID, so, sourceDir, destDir)
	err := images.MakeThumbFromStaticFile(sourceDir, destDir, slugified, 1200, 900)
	if err != nil {
		app.ErrorLog.Println(err)
	}

	destDir = fmt.Sprintf("%s/thumbs", destDir)
	err = images.MakeThumbFromStaticFile(sourceDir, destDir, slugified, 320, 240)
	if err != nil {
		app.ErrorLog.Println(err)
	}

	// get current max for sort order
	curSort, err := vehicleModel.GetMaxSortOrderForVehicleID(vehicleID)
	so = so + curSort

	// write image to db
	vi := clientmodels.Image{
		VehicleID: vehicleID,
		Image:     slugified,
		SortOrder: so,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = vehicleModel.InsertVehicleImage(vi)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}
}
