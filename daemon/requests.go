/*
 *  Turtle - Rock Solid Cluster Management
 *  Copyright DesertBit
 *  Author: Roland Singer
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/desertbit/turtle/api"
	"github.com/desertbit/turtle/daemon/apps"
	"github.com/desertbit/turtle/utils"

	log "github.com/Sirupsen/logrus"
)

var (
	// This mutex is locked in read mode by all requests.
	// A write lock is perfomed to block the requests method
	// during application shutdown.
	requestRWLock sync.RWMutex
)

func init() {
	// Set the request HTTP handler.
	http.HandleFunc("/", handleRequest)
}

//############################//
//### Main request handler ###//
//############################//

// handleRequest main HTTP handler.
func handleRequest(rw http.ResponseWriter, req *http.Request) {
	// Lock the mutex in read mode.
	requestRWLock.RLock()
	defer requestRWLock.RUnlock()

	var err error
	var request *api.Request

	// Get the remote address from the client.
	remoteAddr, _ := utils.RemoteAddress(req)

	// The error function.
	handleError := func(err error) {
		// Construct a new response value.
		response := api.NewResponse()

		// Set the status and the error data.
		response.Status = api.StatusError
		response.Data = api.ResponseError{
			ErrorMessage: err.Error(),
		}

		// Log the error.
		if request != nil {
			log.Warningf("Request from client '%s': %s: %v", remoteAddr, request.Type, err)
		} else {
			log.Warningf("Request from client '%s': %v", remoteAddr, err)
		}

		// Marshal the reponse to JSON.
		resJSON, err := response.ToJSON()
		if err != nil {
			log.Error("handleRequest: %v", err)
			http.Error(rw, "Internal Server Error", 500)
			return
		}

		// Send the result to the client.
		rw.Write(resJSON)
	}

	// Recover panics and send the error.
	defer func() {
		if e := recover(); e != nil {
			handleError(fmt.Errorf("panic: %v", e))
		}
	}()

	// Create the request value from the http JSON body.
	request, err = api.NewRequestFromJSON(req.Body)
	if err != nil {
		handleError(err)
		return
	}

	// The API versions have to match.
	if request.Version != api.Version {
		handleError(fmt.Errorf("API Versions don't match: client=%s server=%s", request.Version, api.Version))
		return
	}

	// Log the request.
	log.Infof("Request from client '%s': %s: %+v", remoteAddr, request.Type, request.Data)

	// The response data interface.
	var data interface{}

	// Handle the specific request.
	switch request.Type {
	case api.TypeAdd:
		data, err = handleAdd(request)
	case api.TypeRemove:
		data, err = handleRemove(request)
	case api.TypeInfo:
		data, err = handleInfo(request)
	case api.TypeList:
		data, err = handleList(request)
	case api.TypeStart:
		data, err = handleStart(request)
	case api.TypeStop:
		data, err = handleStop(request)
	case api.TypeRestart:
		data, err = handleRestart(request)
	case api.TypeSetupGet:
		data, err = handleSetupGet(request)
	case api.TypeSetupSet:
		data, err = handleSetupSet(request)
	case api.TypeErrorMsg:
		data, err = handleErrorMsg(request)
	case api.TypeUpdate:
		data, err = handleUpdate(request)
	case api.TypeBackup:
		data, err = handleBackup(request)
	case api.TypeListBackups:
		data, err = handleListBackups(request)
	case api.TypeRemoveBackup:
		data, err = handleRemoveBackup(request)
	case api.TypeRestoreBackup:
		data, err = handleRestoreBackup(request)
	default:
		handleError(fmt.Errorf("unkown request type '%v'", request.Type))
		return
	}

	// Handle the error.
	if err != nil {
		handleError(err)
		return
	}

	// Construct a new response value.
	response := api.NewResponse()

	// Set the status and the data.
	response.Status = api.StatusSuccess
	response.Data = data

	// Marshal the reponse to JSON.
	resJSON, err := response.ToJSON()
	if err != nil {
		handleError(err)
		return
	}

	// Send the result to the client.
	rw.Write(resJSON)
}

// handleAdd handles the add App request.
func handleAdd(request *api.Request) (interface{}, error) {
	// Map the data to the custom type.
	var data api.RequestAdd
	err := request.MapTo(&data)
	if err != nil {
		return nil, err
	}

	// Validate.
	if len(data.Name) == 0 || len(data.SourceURL) == 0 || len(data.Branch) == 0 {
		return nil, fmt.Errorf("missing or invalid data: %+v", data)
	}

	// Add a new app.
	err = apps.Add(data.Name, data.SourceURL, data.Branch)
	if err != nil {
		return nil, fmt.Errorf("failed to add app: %v", err)
	}

	return nil, nil
}

// handleInfo returns information about an app.
func handleInfo(request *api.Request) (interface{}, error) {
	// Map the data to the custom type.
	var data api.RequestInfo
	err := request.MapTo(&data)
	if err != nil {
		return nil, err
	}

	// Validate.
	if len(data.Name) == 0 {
		return nil, fmt.Errorf("missing or invalid data: %+v", data)
	}

	// Obtain the app with the given name.
	a, err := apps.Get(data.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get information of app: %v", err)
	}

	// Get the setup data from the app.
	setup, err := a.GetSetup()
	if err != nil {
		return nil, fmt.Errorf("failed to get information of app: %v", err)
	}

	// Get the turtlefile.
	// Continue also on error.
	t, err := a.Turtlefile()
	if err != nil {
		return nil, err
	}

	// Create the response value.
	res := api.ResponseInfo{
		Name:       a.Name(),
		Maintainer: t.Maintainer,
		Turtlefile: t.Name,
		State:      a.State(),
		SourceURL:  a.SourceURL(),
		Branch:     a.Branch(),

		Setup: setup,
	}

	return res, nil
}

// handleList sends a list of all Apps.
func handleList(request *api.Request) (interface{}, error) {
	// Get all apps.
	curApps := apps.Apps()

	// Create the response value.
	res := api.ResponseList{
		Apps: make([]api.ResponseListApp, len(curApps)),
	}

	// Add all the apps to the response value.
	for i, app := range curApps {
		// Get the turtlefile.
		// Continue also on error.
		var turtlefileName string
		t, err := app.Turtlefile()
		if err != nil {
			turtlefileName = "ERROR In FILE"
		} else {
			turtlefileName = t.Name
		}

		res.Apps[i] = api.ResponseListApp{
			Name:       app.Name(),
			Turtlefile: turtlefileName,
			State:      app.State(),
		}
	}

	return res, nil
}

// handleRemove handles the remove App request.
func handleRemove(request *api.Request) (interface{}, error) {
	// Map the data to the custom type.
	var data api.RequestRemove
	err := request.MapTo(&data)
	if err != nil {
		return nil, err
	}

	// Validate.
	if len(data.Name) == 0 {
		return nil, fmt.Errorf("missing or invalid data: %+v", data)
	}

	// Obtain the app with the given name.
	a, err := apps.Get(data.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to remove app: %v", err)
	}

	// Remove the app.
	err = a.Remove(data.RemoveBackups)
	if err != nil {
		return nil, fmt.Errorf("failed to remove app: %v", err)
	}

	return nil, nil
}

// handleStart handles the start App request.
func handleStart(request *api.Request) (interface{}, error) {
	// Map the data to the custom type.
	var data api.RequestStart
	err := request.MapTo(&data)
	if err != nil {
		return nil, err
	}

	// Validate.
	if len(data.Name) == 0 {
		return nil, fmt.Errorf("missing or invalid data: %+v", data)
	}

	// Obtain the app with the given name.
	a, err := apps.Get(data.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to start app: %v", err)
	}

	// Start the app.
	err = a.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start app: %v", err)
	}

	return nil, nil
}

// handleStop handles the stop App request.
func handleStop(request *api.Request) (interface{}, error) {
	// Map the data to the custom type.
	var data api.RequestStop
	err := request.MapTo(&data)
	if err != nil {
		return nil, err
	}

	// Validate.
	if len(data.Name) == 0 {
		return nil, fmt.Errorf("missing or invalid data: %+v", data)
	}

	// Obtain the app with the given name.
	a, err := apps.Get(data.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to stop app: %v", err)
	}

	// Stop the app.
	err = a.Stop()
	if err != nil {
		return nil, fmt.Errorf("failed to stop app: %v", err)
	}

	return nil, nil
}

// handleRestart handles the restart App request.
func handleRestart(request *api.Request) (interface{}, error) {
	// Map the data to the custom type.
	var data api.RequestRestart
	err := request.MapTo(&data)
	if err != nil {
		return nil, err
	}

	// Validate.
	if len(data.Name) == 0 {
		return nil, fmt.Errorf("missing or invalid data: %+v", data)
	}

	// Obtain the app with the given name.
	a, err := apps.Get(data.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to restart app: %v", err)
	}

	// Stop the app.
	err = a.Restart()
	if err != nil {
		return nil, fmt.Errorf("failed to restart app: %v", err)
	}

	return nil, nil
}

// handleSetup sends the setup values of an app.
func handleSetupGet(request *api.Request) (interface{}, error) {
	// Map the data to the custom type.
	var data api.RequestSetupGet
	err := request.MapTo(&data)
	if err != nil {
		return nil, err
	}

	// Validate.
	if len(data.Name) == 0 {
		return nil, fmt.Errorf("missing or invalid data: %+v", data)
	}

	// Obtain the app with the given name.
	a, err := apps.Get(data.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to setup app: %v", err)
	}

	// Get the response setup data from the app.
	setup, err := a.GetSetup()
	if err != nil {
		return nil, fmt.Errorf("failed to setup app: %v", err)
	}

	return setup, nil
}

// handleSetupSet obtains the setup values and configures the app.
func handleSetupSet(request *api.Request) (interface{}, error) {
	// Map the data to the custom type.
	var data api.RequestSetupSet
	err := request.MapTo(&data)
	if err != nil {
		return nil, err
	}

	// Validate.
	if len(data.Name) == 0 {
		return nil, fmt.Errorf("missing or invalid data: %+v", data)
	}

	// Obtain the app with the given name.
	a, err := apps.Get(data.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to setup app: %v", err)
	}

	// Get the response setup data from the app.
	err = a.Setup(&data.Setup)
	if err != nil {
		return nil, fmt.Errorf("failed to setup app: %v", err)
	}

	return nil, nil
}

// handleErrorMsg send the error message of an app if present.
func handleErrorMsg(request *api.Request) (interface{}, error) {
	// Map the data to the custom type.
	var data api.RequestErrorMsg
	err := request.MapTo(&data)
	if err != nil {
		return nil, err
	}

	// Validate.
	if len(data.Name) == 0 {
		return nil, fmt.Errorf("missing or invalid data: %+v", data)
	}

	// Obtain the app with the given name.
	a, err := apps.Get(data.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get error message: %v", err)
	}

	// Create the response value.
	res := api.ResponseErrorMsg{
		Name: data.Name,
	}

	// Get the app error if present.
	appErr := a.Error()
	if appErr != nil {
		res.ErrorMessage = strings.TrimSpace(appErr.Error())
	}

	return res, nil
}

// handleUpdate handles the update App request.
func handleUpdate(request *api.Request) (interface{}, error) {
	// Map the data to the custom type.
	var data api.RequestUpdate
	err := request.MapTo(&data)
	if err != nil {
		return nil, err
	}

	// Validate.
	if len(data.Name) == 0 {
		return nil, fmt.Errorf("missing or invalid data: %+v", data)
	}

	// Obtain the app with the given name.
	a, err := apps.Get(data.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to update app: %v", err)
	}

	// Update the app.
	err = a.Update()
	if err != nil {
		return nil, fmt.Errorf("failed to update app: %v", err)
	}

	return nil, nil
}

// handleBackup creates a hot backup.
func handleBackup(request *api.Request) (interface{}, error) {
	// Map the data to the custom type.
	var data api.RequestBackup
	err := request.MapTo(&data)
	if err != nil {
		return nil, err
	}

	// Validate.
	if len(data.Name) == 0 {
		return nil, fmt.Errorf("missing or invalid data: %+v", data)
	}

	// Obtain the app with the given name.
	a, err := apps.Get(data.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to backup app: %v", err)
	}

	// Backup the app.
	err = a.Backup()
	if err != nil {
		return nil, fmt.Errorf("failed to backup app: %v", err)
	}

	return nil, nil
}

// handleListBackups returns a list of all app backups.
func handleListBackups(request *api.Request) (interface{}, error) {
	// Map the data to the custom type.
	var data api.RequestListBackups
	err := request.MapTo(&data)
	if err != nil {
		return nil, err
	}

	// Validate.
	if len(data.Name) == 0 {
		return nil, fmt.Errorf("missing or invalid data: %+v", data)
	}

	// Obtain the app with the given name.
	a, err := apps.Get(data.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %v", err)
	}

	// Get all backup timestamps of the app.
	list, err := a.Backups()
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %v", err)
	}

	// Create the response value.
	res := api.ResponseListBackups{
		Backups: make([]api.ResponseListBackup, len(list)),
	}

	// Add all the backup timestamps to the response value.
	for i, u := range list {
		unix, err := strconv.ParseInt(u, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to list backups: failed to parse unix timestamp: %v", err)
		}

		res.Backups[i] = api.ResponseListBackup{
			Unix: u,
			Date: time.Unix(unix, 0).String(),
		}
	}

	return res, nil
}

// handleRemoveBackup handles the remove backup request.
func handleRemoveBackup(request *api.Request) (interface{}, error) {
	// Map the data to the custom type.
	var data api.RequestRemoveBackup
	err := request.MapTo(&data)
	if err != nil {
		return nil, err
	}

	// Validate.
	if len(data.Name) == 0 || len(data.Unix) == 0 {
		return nil, fmt.Errorf("missing or invalid data: %+v", data)
	}

	// Obtain the app with the given name.
	a, err := apps.Get(data.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to remove backup: %v", err)
	}

	// Remove the backup.
	err = a.RemoveBackup(data.Unix)
	if err != nil {
		return nil, fmt.Errorf("failed to remove backup: %v", err)
	}

	return nil, nil
}

// handleRestoreBackup handles the restore backup request.
func handleRestoreBackup(request *api.Request) (interface{}, error) {
	// Map the data to the custom type.
	var data api.RequestRestoreBackup
	err := request.MapTo(&data)
	if err != nil {
		return nil, err
	}

	// Validate.
	if len(data.Name) == 0 || len(data.Unix) == 0 {
		return nil, fmt.Errorf("missing or invalid data: %+v", data)
	}

	// Obtain the app with the given name.
	a, err := apps.Get(data.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to restore backup: %v", err)
	}

	// Restore the backup.
	err = a.RestoreBackup(data.Unix)
	if err != nil {
		return nil, fmt.Errorf("failed to restore backup: %v", err)
	}

	return nil, nil
}
