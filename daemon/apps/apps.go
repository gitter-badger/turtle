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

package apps

import (
	"fmt"
	"io/ioutil"
	"sort"
	"sync"

	"github.com/desertbit/turtle/daemon/btrfs"
	"github.com/desertbit/turtle/daemon/config"
	"github.com/desertbit/turtle/utils"

	log "github.com/Sirupsen/logrus"
)

var (
	apps      map[string]*App = make(map[string]*App)
	appsMutex sync.Mutex
)

//##############//
//### Public ###//
//##############//

// Release this package and lock the Apps access mutex.
// Once released, the mutex is locked forever.
func Release() {
	// Lock the mutex.
	appsMutex.Lock()
}

// LoadApps reads all apps from the turtle apps directory and fills the apps map.
func LoadApps() error {
	// Lock the mutex.
	appsMutex.Lock()
	defer appsMutex.Unlock()

	// Clear the map if required.
	if apps != nil {
		apps = make(map[string]*App)
	}

	// Get all files in the apps directory.
	files, err := ioutil.ReadDir(config.Config.AppPath)
	if err != nil {
		return err
	}

	// Load all the apps.
	for _, f := range files {
		// Skip if not a directory.
		if !f.IsDir() {
			continue
		}

		// Load the app.
		// The directory name is the app name.
		a, err := loadApp(f.Name())
		if err != nil {
			return err
		}

		// Add the app to the map.
		apps[a.name] = a
	}

	return nil
}

// Apps returns a slice of all apps.
func Apps() []*App {
	// Lock the mutex.
	appsMutex.Lock()
	defer appsMutex.Unlock()

	// Get all map keys.
	i := 0
	keys := make([]string, len(apps))
	for k, _ := range apps {
		keys[i] = k
		i++
	}

	// Sort the keys slice.
	sort.Strings(keys)

	// Create the slice
	s := make([]*App, len(apps))

	// Add all map items sorted to the slice.
	for i, k := range keys {
		s[i] = apps[k]
	}

	return s
}

// Get an app by its name.
// An error is returned if this failed.
func Get(name string) (*App, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("get app: empty name passed.")
	}

	// Lock the mutex.
	appsMutex.Lock()
	defer appsMutex.Unlock()

	// Obtain the app with the given name.
	app, ok := apps[name]
	if !ok {
		return nil, fmt.Errorf("an App with the name '%s' does not exists.", name)
	}

	return app, nil
}

// Add and register a new turtle App.
// A turtle App name should not not contain any whitespaces.
func Add(name, sourceURL, branch string) (err error) {
	var a *App

	// Cleanup on any error.
	defer func() {
		if err == nil || a == nil {
			return
		}

		// Remove the app subvolume if it exists.
		e, errC := utils.Exists(a.path)
		if errC != nil {
			log.Errorf("failed to cleanup failed add app action: %v", errC)
		} else if e {
			if errC = btrfs.DeleteSubvolume(a.path); errC != nil {
				log.Errorf("failed to cleanup failed add app action: %v", errC)
			}
		}
	}()

	// Lock the mutex.
	appsMutex.Lock()
	defer appsMutex.Unlock()

	// Check if an app with the same name already exists.
	_, ok := apps[name]
	if ok {
		return fmt.Errorf("an App with the name '%s' already exists!", name)
	}

	// Create a new app value.
	a, err = newApp(name)
	if err != nil {
		return fmt.Errorf("failed to create add: %v", err)
	}

	// Set the app's settings.
	a.settings.SourceURL = sourceURL
	a.settings.Branch = branch

	// Create the app's subvolume.
	if err = btrfs.CreateSubvolume(a.path); err != nil {
		return fmt.Errorf("failed to prepare app's environment: %v", err)
	}

	// Save the app's settings to the app settings file.
	if err = a.saveSettings(); err != nil {
		return fmt.Errorf("failed to prepare app's environment: %v", err)
	}

	// Create the source directory.
	err = utils.MkDirIfNotExists(a.SourceDirectoryPath())
	if err != nil {
		return fmt.Errorf("failed to prepare app's environment: %v", err)
	}

	// Clone the source in a new task.
	if err = a.cloneSource(); err != nil {
		return err
	}

	// Finally add the app to the map.
	apps[name] = a

	return nil
}
