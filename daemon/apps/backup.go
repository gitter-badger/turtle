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
	"os"
	"strconv"
	"time"

	"github.com/desertbit/turtle/daemon/btrfs"
	"github.com/desertbit/turtle/utils"

	log "github.com/Sirupsen/logrus"
)

//##########################//
//### App Backup methods ###//
//##########################//

// Backup the app data.
func (a *App) Backup() error {
	// Lock the task mutex.
	// The app should not be started during a backup process.
	a.taskMutex.Lock()
	defer a.taskMutex.Unlock()

	// Perform the actual backup.
	return a.backup()
}

// backup the app data.
// This method won't lock the taskMutex. You have to handle it!
func (a *App) backup() error {
	// Don't backup during some special app tasks.
	if a.task == taskCloneSource ||
		a.task == taskUpdate {
		return fmt.Errorf("can't backup app '%s' during an update task!", a.name)
	}

	// Get the app's base backup folder.
	backupPath := a.BackupDirectoryPath()

	// Create the base app backup folder if not present.
	err := utils.MkDirIfNotExists(backupPath)
	if err != nil {
		return fmt.Errorf("failed to backup app '%s': %v", a.name, err)
	}

	// Create a new backup directory with the current timestamp.
	backupPath += "/" + strconv.FormatInt(time.Now().Unix(), 10)

	// Log
	log.Infof("creating backup of app '%s': %s", a.name, backupPath)

	// Create a snapshot of the complete app subvolume.
	err = btrfs.Snapshot(a.path, backupPath, true)
	if err != nil {
		return fmt.Errorf("failed to backup app '%s': %v", a.name, err)
	}

	return nil
}

// Backups returns a slice of all app backup timestamps.
func (a *App) Backups() ([]string, error) {
	dir := a.BackupDirectoryPath()

	// If the backup folder does not exists, then return nil.
	e, err := utils.Exists(dir)
	if err != nil {
		return nil, err
	} else if !e {
		return nil, nil
	}

	// Get all files in the apps directory.
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	backups := make([]string, len(files))

	// Get all the backup timestampts.
	for i, f := range files {
		// Skip if not a directory.
		if !f.IsDir() {
			continue
		}

		backups[i] = f.Name()
	}

	return backups, nil
}

// RemoveBackup removes the given backup.
func (a *App) RemoveBackup(timestamp string) error {
	// Create the backup directory path.
	path := a.BackupDirectoryPath() + "/" + timestamp

	// Check if the backup exists.
	if !btrfs.IsSubvolume(path) {
		return fmt.Errorf("no backup '%s' found!", timestamp)
	}

	log.Infof("Removing backup '%s' of app '%s'.", timestamp, a.name)

	// Remove the backup subvolume.
	err := btrfs.DeleteSubvolume(path)
	if err != nil {
		return fmt.Errorf("failed to delete backup subvolume '%s': %v", timestamp, err)
	}

	return nil
}

// RemoveAllBackups removes all backups of the app.
// Don't call this during any other backup method call.
func (a *App) RemoveAllBackups() error {
	// First get all backups.
	backups, err := a.Backups()
	if err != nil {
		return err
	}

	// Remove all backups.
	for _, b := range backups {
		if err = a.RemoveBackup(b); err != nil {
			return err
		}
	}

	// Delete the base backup directory.
	err = os.RemoveAll(a.BackupDirectoryPath())
	if err != nil {
		return fmt.Errorf("failed to remove app base backup directory: %v", err)
	}

	return nil
}

// RestoreBackup restores the given app backup.
func (a *App) RestoreBackup(timestamp string) (err error) {
	// Lock the task mutex.
	// The app should not be started during a backup process.
	a.taskMutex.Lock()
	defer a.taskMutex.Unlock()

	// Abort if any app ask is running.
	if a.IsTaskRunning() {
		return fmt.Errorf("the app is running!")
	}

	// Create the backup directory path.
	backupPath := a.BackupDirectoryPath() + "/" + timestamp

	// Check if the backup exists.
	if !btrfs.IsSubvolume(backupPath) {
		return fmt.Errorf("no backup '%s' found!", timestamp)
	}

	// Create the apps backup path for the current data.
	newAppBackupPath := a.BackupDirectoryPath() + "/" + strconv.FormatInt(time.Now().Unix(), 10)

	// Log
	log.Infof("restoring backup of app '%s': %s", a.name, timestamp)

	// Reload the turtlefile and settings on defer.
	defer func() {
		err = a.reload()
	}()

	// Move the current subvolume to the backup location with the current timestamp.
	err = os.Rename(a.path, newAppBackupPath)
	if err != nil {
		return fmt.Errorf("failed to move apps current subvolume: %v", err)
	}

	// Restore the current backup subvolume on failure.
	defer func() {
		if err != nil {
			errR := os.Rename(newAppBackupPath, a.path)
			if errR != nil {
				log.Errorf("failed to restore apps current subvolume backup: %v", errR)
			} else {
				// Remove the readonly flag again.
				errR = btrfs.SetSubvolumeReadonly(a.path, false)
				if errR != nil {
					log.Errorf("failed to restore apps subvolume flag: %v", errR)
				}
			}
		}
	}()

	// Set the readonly flag on the moved subvolume.
	err = btrfs.SetSubvolumeReadonly(newAppBackupPath, true)
	if err != nil {
		return err
	}

	// Create a snapshot of the backup and restore it to the app path.
	err = btrfs.Snapshot(backupPath, a.path, false)
	if err != nil {
		return fmt.Errorf("failed to restore app '%s': %v", a.name, err)
	}

	return nil
}
