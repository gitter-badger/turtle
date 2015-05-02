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
	"strconv"
	"strings"
	"time"

	"github.com/desertbit/turtle/daemon/apps"
	"github.com/desertbit/turtle/daemon/config"

	log "github.com/Sirupsen/logrus"
)

const (
	removeBackupInterval = 5 * time.Hour
)

func autoRemoveOldBackupsLoop() {
	for {
		// Sleep.
		time.Sleep(removeBackupInterval / 2)

		// Start the actual job.
		err := removeOldBackups()
		if err != nil {
			log.Errorf("failed to automatically remove some old backups:\n%v", err)
		}

		// Sleep.
		time.Sleep(removeBackupInterval / 2)
	}
}

func removeOldBackups() error {
	var allErr string

	addErr := func(err error) {
		allErr += err.Error() + "\n"
	}

	// Create the expiration unix timestamp.
	expire := time.Now().Unix() - config.Config.KeepBackupsDuration

	// Get all apps.
	curApps := apps.Apps()

	for _, app := range curApps {
		// Get all backups of the app.
		backups, err := app.Backups()
		if err != nil {
			addErr(err)
			continue
		}

		// Remove all old backups.
		for _, b := range backups {
			u, err := strconv.ParseInt(b, 10, 64)
			if err != nil {
				addErr(err)
				continue
			}

			// Skip if this is not an old backup.
			if (expire - u) <= 0 {
				continue
			}

			log.Infof("Removing old backup '%s' of app '%s'.", b, app.Name())

			// Remove the backup.
			err = app.RemoveBackup(b)
			if err != nil {
				addErr(err)
				continue
			}
		}
	}

	// Trim the all error messages string.
	allErr = strings.TrimSpace(allErr)

	// Return the error(s) if present.
	if len(allErr) > 0 {
		return fmt.Errorf(allErr)
	}

	return nil
}
