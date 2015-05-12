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
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/desertbit/turtle/daemon/apps"
	"github.com/desertbit/turtle/daemon/btrfs"
	"github.com/desertbit/turtle/daemon/config"
	"github.com/desertbit/turtle/daemon/docker"
	"github.com/desertbit/turtle/utils"

	log "github.com/Sirupsen/logrus"
)

const (
	InterruptExitCode = -1
)

func onInterrupt() {
	// Wait for the signal
	sigchan := make(chan os.Signal, 10)
	signal.Notify(sigchan, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGKILL)
	<-sigchan

	log.Info("Exiting...")

	// First cleanup
	release()

	// Just wait for a moment before exiting to be
	// sure, all defers get called and the program
	// performs a clean exit.
	time.Sleep(250 * time.Millisecond)

	// Exit the application
	os.Exit(InterruptExitCode)
}

// release is called as soon as the daemon application is terminating.
func release() {
	// Block the requests http handler method.
	// Don't handle any further requests,
	requestRWLock.Lock()

	// Block the remove old backups job.
	removeOldBackupsMutex.Lock()

	// Save the current state of all running apps...
	err := saveCurrentState()
	if err != nil {
		log.Errorf("failed to save current turtle state: %v", err)
	}

	// Release the app package.
	apps.Release()
}

// prepareEnv prepares the turtle environment.
func prepareEnv() (err error) {
	// Create the directories if they don't exists.
	createDirs := []string{
		config.Config.AppPath,
		config.Config.BackupPath,
		config.Config.TurtlePath,
	}

	for _, dir := range createDirs {
		if err = utils.MkDirIfNotExists(dir); err != nil {
			return err
		}
	}

	// Populate the saved host file.
	if err = populateKnownHosts(); err != nil {
		return err
	}

	// Initialize and load the apps.
	if err = apps.LoadApps(); err != nil {
		return err
	}

	return nil
}

func btrfsBalanceJob() {
	for {
		// Sleep.
		time.Sleep(config.Config.BtrfsBalanceInterval / 2)

		log.Infof("Balancing btrfs path '%s'...", config.TurtleRoot)

		// Balance the turtle root btrfs partition.
		err := btrfs.Balance(config.TurtleRoot)
		if err != nil {
			log.Errorf("Balancing of btrfs path '%s' failed: %v", config.TurtleRoot, err)
		} else {
			log.Infof("Balancing of btrfs path '%s' done.", config.TurtleRoot)
		}

		// Sleep.
		time.Sleep(config.Config.BtrfsBalanceInterval / 2)
	}
}

func main() {
	// Set the maximum number of CPUs that can be executing simultaneously.
	runtime.GOMAXPROCS(runtime.NumCPU())

	log.Infof("Initializing...")

	// Prepare the turtle environment.
	err := prepareEnv()
	if err != nil {
		log.Fatalf("failed to prepare turtle environment: %v", err)
	}

	// Initialize the docker client.
	if err = docker.Init(); err != nil {
		log.Fatalln(err)
	}

	// Catch interrupts.
	go onInterrupt()

	// Start the btrfs balance job.
	go btrfsBalanceJob()

	// Restore the previous state.
	// Start all apps which where running during the last daemon shutdown...
	err = restoreState()
	if err != nil {
		log.Warningf("failed to restore previous turtle state: %v", err)
	}

	// Start the loop to remove old backups.
	go autoRemoveOldBackupsLoop()

	// Log
	log.Infof("Turtle server listening on '%s'", config.Config.ListenAddress)

	// Start the http server.
	log.Fatal(http.ListenAndServe(config.Config.ListenAddress, nil))
}
