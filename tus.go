package clienthandlers

import (
	"encoding/json"
	"fmt"
	"github.com/gosimple/slug"
	channel_data "github.com/tsawler/goblender/pkg/channel-data"
	"github.com/tsawler/goblender/pkg/config"
	"github.com/tsawler/goblender/pkg/helpers"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// TusMetaData is the metadata
type TusMetaData struct {
	FileName     string `json:"filename"`
	Token        string `json:"token"`
	FileType     string `json:"file_type"`
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	UploadType   string `json:"upload_type"`
	UploadTo     string `json:"upload_to"`
	SortOrder    string `json:"sort_order"`
	ProcessVideo string `json:"process"`
}

// TusStorage is the storage
type TusStorage struct {
	Type string `json:"Type"`
	Path string `json:"Path"`
}

// TusUpload is the actual data
type TusUpload struct {
	ID             string      `json:"ID"`
	Size           int         `json:"Size"`
	SizeIsDeferred bool        `json:"SizeIsDeferred"`
	Offset         int         `json:"Offset"`
	IsFinal        bool        `json:"IsFinal"`
	IsPartial      bool        `json:"IsPartial"`
	MetaData       TusMetaData `json:"MetaData"`
	Storage        TusStorage  `json:"Storage"`
}

// Upload is the json post from tus
type Upload struct {
	Upload TusUpload `json:"Upload"`
}

// TusWebHook handles web hook events for tus uploads
func TusWebHook(app config.AppConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var payload Upload
		err = json.Unmarshal(b, &payload)
		if err != nil {
			app.ErrorLog.Println(fmt.Sprintf("Error parsing webhook JSON: %v\n", err))
			return
		}

		hookName := r.Header.Get("Hook-Name")
		if hookName == "pre-create" {
			// validate the request is coming from a user we know about
			// we have their user id and email, so make sure they match ones in the system
			userID, _ := strconv.Atoi(payload.Upload.MetaData.UserID)
			email := payload.Upload.MetaData.Token
			canUpload := repo.DB.ValidateUploadByEmailAndID(userID, email)
			if !canUpload {
				// invalid user, so delete the video and throw a 401 at them
				videoID, _ := strconv.Atoi(payload.Upload.MetaData.ID)
				_ = repo.DB.DeleteVideoById(videoID)
				http.Error(w, "Not authorized", http.StatusUnauthorized)
				return
			}
		} else if hookName == "post-finish" {

			if payload.Upload.MetaData.UploadType == "video" {
				// we'll wait one second just in case it is a really small file, and the os hasn't updated
				time.Sleep(time.Second)

				videoID, _ := strconv.Atoi(payload.Upload.MetaData.ID)
				userID, _ := strconv.Atoi(payload.Upload.MetaData.UserID)
				target := fmt.Sprintf("%s/%s", app.TusDir, payload.Upload.ID)
				processVideo := true
				if payload.Upload.MetaData.ProcessVideo == "0" {
					processVideo = false
				}

				jobData := channel_data.VideoData{
					ID:           videoID,
					InputPath:    target,
					VideoName:    payload.Upload.MetaData.FileName,
					UserID:       userID,
					ProcessVideo: processVideo,
				}

				job := channel_data.VideoProcessingJob{
					Video: jobData,
				}

				app.VideoQueue <- job
			} else if payload.Upload.MetaData.UploadType == "file-manager" {

				oldLocation := fmt.Sprintf("%s/%s", app.TusDir, payload.Upload.ID)
				fileName := payload.Upload.MetaData.FileName
				dot := strings.LastIndex(fileName, ".")
				rootName := fileName[0:dot]
				last4 := fileName[dot:len(fileName)]
				slugified := fmt.Sprintf("%s%s", slug.Make(rootName), last4)

				newLocation := fmt.Sprintf("%s/%s", payload.Upload.MetaData.UploadTo, slugified)

				err := MoveFile(oldLocation, newLocation)
				if err != nil {
					app.ErrorLog.Println("could not move from", oldLocation, "to", newLocation)
				}

				err = os.Remove(fmt.Sprintf("%s.info", oldLocation))
				if err != nil {
					app.ErrorLog.Println("Error deleting info file")
				}
			} else if payload.Upload.MetaData.UploadType == "image-manager" {
				userID, _ := strconv.Atoi(payload.Upload.MetaData.UserID)
				fileName := payload.Upload.MetaData.FileName
				dot := strings.LastIndex(fileName, ".")
				rootName := fileName[0:dot]
				last4 := fileName[dot:len(fileName)]
				slugified := fmt.Sprintf("%s%s", slug.Make(rootName), last4)
				oldLocation := fmt.Sprintf("%s/%s", app.TusDir, payload.Upload.ID)
				newLocation := fmt.Sprintf("%s/%s", payload.Upload.MetaData.UploadTo, slugified)

				err := MoveFile(oldLocation, newLocation)
				if err != nil {
					app.ErrorLog.Println("could not move from", oldLocation, "to", newLocation)
				}

				// make thumb
				sourceDir := payload.Upload.MetaData.UploadTo
				destDir := fmt.Sprintf("%s/.thumb", sourceDir)

				jobData := channel_data.ImageData{
					UserID:      userID,
					SourceDir:   sourceDir,
					DestDir:     destDir,
					Slugified:   slugified,
					OldLocation: oldLocation,
				}

				job := channel_data.ImageProcessingJob{
					Image: jobData,
				}

				app.ImageQueue <- job
			} else if payload.Upload.MetaData.UploadType == "inventory" {
				vehicleID, _ := strconv.Atoi(payload.Upload.MetaData.ID)
				userID, _ := strconv.Atoi(payload.Upload.MetaData.UserID)

				_ = helpers.CreateDirIfNotExist(fmt.Sprintf("./ui/static/site-content/inventory/%d", vehicleID))
				_ = helpers.CreateDirIfNotExist(fmt.Sprintf("./ui/static/site-content/inventory/%d/thumbs", vehicleID))

				fileName := payload.Upload.MetaData.FileName
				dot := strings.LastIndex(fileName, ".")
				rootName := fileName[0:dot]
				last4 := fileName[dot:len(fileName)]
				slugified := fmt.Sprintf("%s%s", slug.Make(rootName), last4)

				oldLocation := fmt.Sprintf("%s/%s", app.TusDir, payload.Upload.ID)
				newLocation := fmt.Sprintf("%s/%s", payload.Upload.MetaData.UploadTo, slugified)

				err := MoveFile(oldLocation, newLocation)
				if err != nil {
					app.ErrorLog.Println("could not move from", oldLocation, "to", newLocation)
				}
				so, _ := strconv.Atoi(payload.Upload.MetaData.SortOrder)

				// make thumb
				sourceDir := payload.Upload.MetaData.UploadTo
				destDir := payload.Upload.MetaData.UploadTo

				jobData := VehicleImageData{
					SourceDir:   sourceDir,
					DestDir:     destDir,
					Slugified:   slugified,
					OldLocation: oldLocation,
					SortOrder:   so,
					VehicleID:   vehicleID,
					UserID:      userID,
				}

				job := VehicleImageProcessingJob{
					Image: jobData,
				}
				vehicleImageQueue <- job
			}
		}
	}
}

// MoveFile moves the file, just in case source/dest are on different volumes
func MoveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s", err)
	}

	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s", err)
	}
	return nil
}
