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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/desertbit/turtle/daemon/btrfs"
	"github.com/desertbit/turtle/daemon/config"
	"github.com/desertbit/turtle/daemon/turtlefile"
	"github.com/desertbit/turtle/utils"

	"github.com/BurntSushi/toml"
)

const (
	settingsFilename = "settings"
	sourceDirectory  = "source"
	volumesDirectory = "volumes"
)

//################//
//### App type ###//
//################//

type App struct {
	name     string       // The App name.
	path     string       // The App directory path.
	settings *appSettings // The App settings. They are saved in the app settings file.

	turtlefile      *turtlefile.Turtlefile // Don't access this directly. It might be nil. Use App.Turtlefile() instead,
	turtlefileMutex sync.Mutex

	task      taskType
	taskMutex sync.Mutex
	taskErr   error
	taskState string

	//##
	//## Run task values:
	//##
	containerIDs []string
	restartApp   bool

	checkRestartMutex   sync.Mutex
	checkRestartRunning bool
	checkRestartError   chan error

	stopRequested           chan struct{}
	stopRequestedChanExists bool
	stopRequestedMutex      sync.Mutex
}

//  newApp creates a new app and sets the app directory path.
// A turtle App name should not not contain any whitespaces.
func newApp(name string) (*App, error) {
	// Return an error if the name contains whitespaces.
	if strings.ContainsRune(name, ' ') ||
		strings.ContainsRune(name, '\t') ||
		strings.ContainsRune(name, '\n') ||
		strings.ContainsRune(name, '\v') ||
		strings.ContainsRune(name, '\f') ||
		strings.ContainsRune(name, '\r') {
		return nil, fmt.Errorf("the app name '%s' contains a whitespace!", name)
	}

	return &App{
		name:     name,
		path:     filepath.Clean(config.Config.AppPath + "/" + name),
		settings: newSettings(),

		task:      taskNone,
		taskState: stateIdle,
	}, nil
}

// loadApp loads the app with the given name from the apps directory.
func loadApp(name string) (*App, error) {
	// Create a new app value.
	a, err := newApp(name)
	if err != nil {
		return nil, err
	}

	// The app directory has to be a btrfs subvolume.
	if !btrfs.IsSubvolume(a.path) {
		return nil, fmt.Errorf("the app's directory '%s' is not a btrfs subvolume!", a.path)
	}

	// Load the app settings.
	err = a.loadSettings()
	if err != nil {
		return nil, err
	}

	return a, nil
}

// Name returns the App's name.
func (a *App) Name() string {
	return a.name
}

// SourceURL returns the app's source url.
func (a *App) SourceURL() string {
	return a.settings.SourceURL
}

// Branch returns the app's branch.
func (a *App) Branch() string {
	return a.settings.Branch
}

func (a *App) SettingsFilePath() string {
	return a.path + "/" + settingsFilename
}

func (a *App) SourceDirectoryPath() string {
	return a.path + "/" + sourceDirectory
}

func (a *App) VolumesDirectoryPath() string {
	return a.path + "/" + volumesDirectory
}

func (a *App) BackupDirectoryPath() string {
	return config.Config.BackupPath + "/" + a.name
}

// Turtlefile returns the app's turtlefile.
func (a *App) Turtlefile() (*turtlefile.Turtlefile, error) {
	// Lock the mutex.
	a.turtlefileMutex.Lock()
	defer a.turtlefileMutex.Unlock()

	// Only load the turtlefile if not already loaded.
	if a.turtlefile != nil {
		return a.turtlefile, nil
	}

	// Obtain the turtlefile path.
	turtlefilePath := a.SourceDirectoryPath() + "/" + turtlefile.TurtlefileFilename
	tStat, err := os.Stat(turtlefilePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("Turtlefile is missing in source directory!")
	} else if err != nil {
		return nil, fmt.Errorf("failed to obtain state of file '%s': %v", turtlefilePath, err)
	}

	// If the path is a directory, then check if the turtlefile exists in it.
	if tStat.IsDir() {
		turtlefilePath += "/" + turtlefile.TurtlefileFilename
		e, err := utils.Exists(turtlefilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to check if file exists '%s': %v", turtlefilePath, err)
		} else if !e {
			return nil, fmt.Errorf("Turtlefile is missing in source directory!")
		}
	}

	// Load the turtlefile.
	t, err := turtlefile.Load(turtlefilePath)
	if err != nil {
		return nil, err
	}

	// Check if the turtlefile is valid.
	if err = t.IsValid(); err != nil {
		return nil, fmt.Errorf("the turtlefile is invalid: %v", err)
	}

	// Set the app's turtlefile pointer.
	a.turtlefile = t

	return t, nil
}

// Remove the app.
func (a *App) Remove() error {
	// Lock the task mutex.
	// The app should not be started during a remove process.
	a.taskMutex.Lock()
	defer a.taskMutex.Unlock()

	// Abort if the app is running.
	if a.IsRunning() {
		return fmt.Errorf("the app is running!")
	}

	// Remove the app subvolume if it exists.
	e, err := utils.Exists(a.path)
	if err != nil {
		return err
	} else if e {
		if err = btrfs.DeleteSubvolume(a.path); err != nil {
			return err
		}
	}

	// Lock the apps mutex.
	appsMutex.Lock()
	defer appsMutex.Unlock()

	// Remove the app from the apps map.
	delete(apps, a.name)

	return nil
}

//###########################//
//### Private App methods ###//
//###########################//

// saveSettings saves the app settings to the app's settings file.
func (a *App) saveSettings() error {
	// Encode the app value to TOML.
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(a.settings)
	if err != nil {
		return fmt.Errorf("failed to encode app settings values to toml: %v", err)
	}

	// Write the result to the app settings file.
	err = ioutil.WriteFile(a.SettingsFilePath(), buf.Bytes(), 0600)
	if err != nil {
		return fmt.Errorf("failed to save app values to settings file: %v", err)
	}

	return nil
}

// loadSettings loads the app's settings from the app settings file.
func (a *App) loadSettings() error {
	// Load and decode the file.
	_, err := toml.DecodeFile(a.SettingsFilePath(), a.settings)
	if err != nil {
		return fmt.Errorf("failed to load app settings file '%s': %v", a.SettingsFilePath(), err)
	}

	return nil
}
