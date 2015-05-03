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
	"path/filepath"
	"time"

	"github.com/desertbit/turtle/daemon/config"
	"github.com/desertbit/turtle/daemon/docker"
	"github.com/desertbit/turtle/utils"

	log "github.com/Sirupsen/logrus"
	d "github.com/fsouza/go-dockerclient"
)

const (
	defaultContainerWaitAfterStartup = 300 * time.Millisecond
	waitAfterAppStart                = 3 * time.Second

	maximumRestartRetryCount = 3
)

//##########################//
//### Public App methods ###//
//##########################//

// IsRunning returns a boolean whenever the app is running.
func (a *App) IsRunning() bool {
	return a.task == taskRun
}

// Start the app.
func (a *App) Start() error {
	if a.IsTaskRunning() {
		return fmt.Errorf("app is already running!")
	} else if !a.IsSetup() {
		return fmt.Errorf("you have to setup the app first!")
	}

	// Create a backup.
	err := a.Backup()
	if err != nil {
		return err
	}

	// Start the task to run the app.
	return a.runTask(taskRun, taskFuncRun)
}

// Stop the app.
func (a *App) Stop() error {
	if !a.IsRunning() {
		return fmt.Errorf("app is not running!")
	}

	// Lock the mutex.
	a.stopRequestedMutex.Lock()
	defer a.stopRequestedMutex.Unlock()

	// Check if the channel can be closed and is not already closed.
	if !a.stopRequestedChanExists {
		return nil
	}

	// Update the flag.
	a.stopRequestedChanExists = false

	// Trigger the stop event by closing the channel.
	close(a.stopRequested)

	return nil
}

// Restart the app.
func (a *App) Restart() error {
	// Set the restart flag.
	a.restartApp = true

	// Stop the app. It will restart automatically, due to the flag.
	return a.Stop()
}

//###############//
//### Private ###//
//###############//

func (a *App) createStopChannel() {
	// Lock the mutex.
	a.stopRequestedMutex.Lock()
	defer a.stopRequestedMutex.Unlock()

	// Create the stop channel and update the flag.
	a.stopRequested = make(chan struct{})
	a.stopRequestedChanExists = true
}

func taskFuncRun(app *App) (err error) {
	// Create a new  backup ticker
	ticker := time.NewTicker(config.Config.BackupInterval)
	stopBackupLoop := make(chan struct{})

	defer func() {
		// Stop the ticker
		ticker.Stop()

		// Stop the loop by triggering the quit trigger.
		close(stopBackupLoop)
	}()

	// Start a backup job in a new goroutine.
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Infof("creating automatic backup of app '%s'.", app.name)

				// Create a backup.
				err := app.Backup()
				if err != nil {
					log.Errorf("failed to create automatic backup of app '%s': %v", app.name, err)
				}
			case <-stopBackupLoop:
				// Just exit the loop
				return
			}
		}
	}()

	for {
		// Reset the restart flag.
		app.restartApp = false

		// Create the stop channel and update the flag.
		app.createStopChannel()

		// Stop and delete all containers if present.
		app.setState("cleaning up containers...")
		if err = stopContainers(app); err != nil {
			return fmt.Errorf("failed to stop and delete previous app containers: %v", err)
		}

		// Setup the app environment.
		app.setState("setting up app environment...")
		if err = setupRunEnvironment(app); err != nil {
			return fmt.Errorf("failed to setup app environment for start request: %v", err)
		}

		// Start the app containers.
		app.setState("starting app...")
		if err = startContainers(app); err != nil {
			return fmt.Errorf("failed to start app containers: %v", err)
		}

		// Watch the app containers.
		if err = watchRunState(app); err != nil {
			return fmt.Errorf("watch app run state: %v", err)
		}

		// Continue the loop and restart the app if the flag is set.
		if app.restartApp {
			// Short timeout.
			time.Sleep(300 * time.Millisecond)
			continue
		}

		// Break the loop
		break
	}

	return nil
}

func setupRunEnvironment(app *App) error {
	// Get the app's volumes directory path.
	volumesPath := app.VolumesDirectoryPath()

	// Create the volumes folder if not present.
	err := utils.MkDirIfNotExists(volumesPath)
	if err != nil {
		return err
	}

	// Get the turtlefile.
	turtlefile, err := app.Turtlefile()
	if err != nil {
		return err
	}

	// Setup the container volume directories if not present.
	var path string
	for _, c := range turtlefile.Containers {
		if len(c.Volumes) == 0 {
			continue
		}

		// Create the base container volume directory path.
		path = volumesPath + "/" + c.Name

		for _, v := range c.Volumes {
			err = utils.MkDirIfNotExists(filepath.Clean(path+"/"+v), 0750)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// stopContainers stops and removes the containers.
func stopContainers(app *App) error {
	if app.containerIDs == nil {
		return nil
	}

	var err error

	// Stop and remove all app containers.
	// Do this in the reverse order, because the container IDs are
	// sorted to the startup order.
	for i := len(app.containerIDs) - 1; i >= 0; i-- {
		err = docker.StopAndDeleteContainer(app.containerIDs[i])
		if err != nil {
			return fmt.Errorf("failed to stop and remove container '%s': %v", app.containerIDs[i], err)
		}

		app.containerIDs = app.containerIDs[:i]
	}

	// Clear the container slice completly.
	app.containerIDs = nil

	return nil
}

func startContainers(app *App) (err error) {
	// Stop already started containers on error.
	defer func() {
		if err != nil && len(app.containerIDs) > 0 {
			if errS := stopContainers(app); errS != nil {
				log.Errorf("failed to stop and delete previous app containers: %v", errS)
			}
		}
	}()

	// Clear the container IDs slice.
	app.containerIDs = nil

	// Get the app's volumes directory path.
	volumesPath := app.VolumesDirectoryPath()

	// Get the turtlefile.
	turtlefile, err := app.Turtlefile()
	if err != nil {
		return err
	}

	// Create the container name prefix.
	cNamePrefix := docker.TurtlePrefix + app.Name() + "."

	// Start each app container.
	// Hint: the containers are already sorted by the turtlefile Load method.
	for _, container := range turtlefile.Containers {
		// Create the docker container name.
		containerName := cNamePrefix + container.Name

		// Check if a container with the same name is present.
		c, err := docker.GetContainerByName(containerName)
		if err != nil {
			return err
		} else if c != nil {
			// Stop and remove it.
			err = docker.StopAndDeleteContainer(c.ID)
			if err != nil {
				return fmt.Errorf("failed to stop and remove container '%s': %v", c.ID, err)
			}
		}

		// Create the bind volumes slice.
		binds := make([]string, len(container.Volumes)+len(container.StaticVolumes))
		i := 0
		for _, v := range container.Volumes {
			binds[i] = filepath.Clean(volumesPath+"/"+container.Name+"/"+v) + ":" + v
			i++
		}

		// Add the static volume mount.
		for _, v := range container.StaticVolumes {
			binds[i] = v
			i++
		}

		// Create the port bindings.
		portBindings := make(map[d.Port][]d.PortBinding)
		for _, p := range app.settings.Ports {
			// Skip if this is not for this container or if disabled.
			if p.ContainerName != container.Name || p.HostPort <= 0 {
				continue
			}

			portProtocol := d.Port(fmt.Sprintf("%v/%s", p.ContainerPort, p.Protocol))

			portBindings[portProtocol] = []d.PortBinding{
				d.PortBinding{
					HostPort: fmt.Sprintf("%v", p.HostPort),
				},
			}
		}

		// Create the links.
		links := make([]string, len(container.Links))
		for i, l := range container.Links {
			links[i] = cNamePrefix + l + ":" + l
		}

		// Create the environment variables slice.
		env, err := app.getEnv()
		if err != nil {
			return err
		}

		// Add the static environment variables.
		env = append(container.Env, env...)

		// Create the host config.
		hostConfig := &d.HostConfig{
			RestartPolicy:   d.NeverRestart(), // the docker daemon will not restart the container automatically.
			Links:           links,
			Privileged:      false,
			PublishAllPorts: false,
			PortBindings:    portBindings,
			Binds:           binds,
		}

		// Create the container options.
		options := &d.CreateContainerOptions{
			Name: containerName,
			Config: &d.Config{
				Image: container.Image,
				Env:   env,
			},
			HostConfig: hostConfig,
		}

		// Check if the image exists.
		_, err = docker.Client.InspectImage(container.Image)
		if err != nil {
			app.setState("pulling docker image: " + container.Image)

			// Pull the image.
			err = docker.Client.PullImage(d.PullImageOptions{
				Repository: container.Image,
				Tag:        container.Tag,
			}, d.AuthConfiguration{})

			if err != nil {
				return fmt.Errorf("failed to pull docker image '%s': %v", container.Image, err)
			}
		}

		app.setState("starting container: " + containerName)

		// Create the docker container.
		c, err = docker.Client.CreateContainer(*options)
		if err != nil {
			return fmt.Errorf("failed to create docker container:\nName: %s\nConfig: %+v\nHost Config: %+v\nError: %v",
				options.Name, options.Config, options.HostConfig, err)
		}

		// Start the container.
		err = docker.Client.StartContainer(c.ID, hostConfig)
		if err != nil {
			return fmt.Errorf("failed to start docker container:\nHost Config: %+v\nError: %v", hostConfig, err)
		}

		// Add the continer ID to the slice.
		app.containerIDs = append(app.containerIDs, c.ID)

		// Wait x milliseconds after the container started.
		// This delays the next container startup.
		// If set to 0, use the default value.
		if container.WaitAfterStartup == 0 {
			time.Sleep(defaultContainerWaitAfterStartup)
		} else {
			time.Sleep(time.Duration(container.WaitAfterStartup) * time.Millisecond)
		}
	}

	// Set the app state.
	app.setState("running")

	// Wait, to be sure all containers started up.
	time.Sleep(waitAfterAppStart)

	return nil
}

func watchRunState(app *App) (err error) {
	// Add an event listener function.
	eventID := docker.OnEvent(func(event *d.APIEvents) {
		// Filter events which are related to container stop events.
		if event.Status != "destroy" &&
			event.Status != "die" &&
			event.Status != "kill" &&
			event.Status != "pause" &&
			event.Status != "restart" &&
			event.Status != "stop" {
			return
		}

		// Check if the event is from one of the apps containers.
		for _, id := range app.containerIDs {
			if id == event.ID {
				// Check all the app containers state and restart them if stopped.
				checkRestart(app)
				break
			}
		}
	})

	// Remove the event listener on defer again.
	defer func() {
		docker.OffEvent(eventID)
	}()

	// Create the error channel.
	app.checkRestartError = make(chan error, 1)

	// Perfom an initial check of all app containers.
	checkRestart(app)

	// Our watch loop.
	for {
		select {
		case err = <-app.checkRestartError:
			return err
		case <-app.stopRequested:
			// Remove the event listener again, before stopping the containers.
			docker.OffEvent(eventID)

			// Set the app state.
			app.setState("stopping...")

			// Log.
			if app.restartApp {
				log.Infof("restarting app '%s'", app.name)
			} else {
				log.Infof("stopping app '%s'", app.name)
			}

			// Stop and remove all app containers.
			if err = stopContainers(app); err != nil {
				return err
			}

			return nil
		}
	}

	return nil
}

// checkRestart checks the state of all app containers and tries to restart them if stopped.
// This is executed in a new goroutine.
// If called multiple times, then the request is executed only once.
func checkRestart(app *App) {
	// Lock the mutex.
	app.checkRestartMutex.Lock()
	defer app.checkRestartMutex.Unlock()

	// Check if already running.
	if app.checkRestartRunning {
		return
	}

	// If not, lock the flag.
	app.checkRestartRunning = true

	go func() {
		defer func() {
			// Lock the mutex.
			app.checkRestartMutex.Lock()
			defer app.checkRestartMutex.Unlock()

			// Unlock our flag.
			app.checkRestartRunning = false
		}()

		// Perform the actual request.
		err := _checkRestart(app, 0)
		if err != nil {
			// Send the error to the channel.
			app.checkRestartError <- err
		}
	}()
}

func _checkRestart(app *App, retryCount int) (err error) {
	// Increment the retry count.
	retryCount++

	var stoppedContainers []*d.Container

	// Get all containers which stopped running.
	for _, id := range app.containerIDs {
		// Obtain the container with its ID.
		c, err := docker.Client.InspectContainer(id)
		if err != nil {
			return err
		}

		// Check if not running.
		if !c.State.Running {
			// Append the stopped container to the slice.
			stoppedContainers = append(stoppedContainers, c)

			// Log
			log.WithFields(
				log.Fields{
					"app":            app.name,
					"container ID":   c.ID,
					"container name": c.Name,
				}).Warning("app container stopped running!")
		}
	}

	// Return if everything is ok.
	if len(stoppedContainers) == 0 {
		return nil
	}

	// Abort and return an error if the retry count is reached.
	if retryCount > maximumRestartRetryCount {
		return fmt.Errorf("failed to restart app: max restart retries reached! App '%s' stopped running!", app.name)
	}

	// Set the app state.
	app.setState("restarting app...")

	// Log.
	log.Infof("restarting app '%s'", app.name)

	// First stop and remove all app containers.
	if err = stopContainers(app); err != nil {
		return err
	}

	// Start the app containers again.
	if err = startContainers(app); err != nil {
		return err
	}

	// Recall this method to recheck the state.
	return _checkRestart(app, retryCount)
}
