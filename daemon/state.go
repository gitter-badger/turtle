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
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/desertbit/turtle/daemon/apps"
	"github.com/desertbit/turtle/daemon/config"
	"github.com/desertbit/turtle/utils"

	"github.com/BurntSushi/toml"
)

type state struct {
	RunningApps []string // List of all running app names.
}

// saveCurrentState saves the current running turtle apps.
func saveCurrentState() error {
	// Create a new state value.
	var s state

	// Get all apps.
	curApps := apps.Apps()

	for _, app := range curApps {
		// Skip if not running.
		if !app.IsRunning() {
			continue
		}

		// Add the app name to the running apps slice.
		s.RunningApps = append(s.RunningApps, app.Name())
	}

	// Encode the state value to TOML.
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(&s)
	if err != nil {
		return fmt.Errorf("failed to encode turtle state value to toml: %v", err)
	}

	// Write the result to the app settings file.
	err = ioutil.WriteFile(config.Config.StateFilePath(), buf.Bytes(), 0600)
	if err != nil {
		return fmt.Errorf("failed to save turtle state value: %v", err)
	}

	return nil
}

// restoreState loads the turtle state file if present and
// restores the previous running apps.
func restoreState() error {
	var s state

	// Set the turtle state file path.
	statePath := config.Config.StateFilePath()

	// Skip if it does not exists.
	e, err := utils.Exists(statePath)
	if err != nil {
		return err
	} else if !e {
		return nil
	}

	// Load and decode the file.
	_, err = toml.DecodeFile(statePath, &s)
	if err != nil {
		return fmt.Errorf("failed to load turtle state file '%s': %v", statePath, err)
	}

	// Get all apps.
	curApps := apps.Apps()

	// Collect all errors during startup.
	var allErr string

	// Start the apps which where running during the last turtle daemon shutdown.
	for _, rApp := range s.RunningApps {
		for _, app := range curApps {
			// Skip if this is not our searched app or if it is already running.
			if rApp != app.Name() || app.IsRunning() {
				continue
			}

			// Start the app.
			err = app.Start()
			if err != nil {
				allErr += err.Error() + "\n"
			}
		}
	}

	// Trim the new linew at the end.
	allErr = strings.TrimSpace(allErr)

	// Return the error(s) if present.
	if len(allErr) > 0 {
		return fmt.Errorf(allErr)
	}

	return nil
}
