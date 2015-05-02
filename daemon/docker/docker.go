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

package docker

import (
	"fmt"
	"strings"
	"sync"

	"github.com/desertbit/turtle/daemon/config"

	log "github.com/Sirupsen/logrus"
	docker "github.com/fsouza/go-dockerclient"
)

const (
	TurtlePrefix = "turtle."

	killAfterTimeout = 10 // in seconds
)

var (
	Client *docker.Client

	eventFuncs        = make(map[int64]func(*docker.APIEvents))
	eventFuncsCounter int64
	eventFuncsMutex   sync.Mutex
)

//##############//
//### Public ###//
//##############//

// Init creates and connects the docker client.
func Init() error {
	var err error

	log.Info("Preparing docker environment...")

	// Connect to the docker server.
	Client, err = docker.NewClient(config.Config.DockerEndPoint)
	if err != nil {
		return fmt.Errorf("failed to create docker client: %v", err)
	}

	// Stop and remove all turtle containers.
	if err = CleanupTurtleContainers(); err != nil {
		return fmt.Errorf("failed to cleanup turtle containers: %v", err)
	}

	// Start the event listener.
	if err = startEventListener(); err != nil {
		return fmt.Errorf("failed to start docker event listener: %v", err)
	}

	return nil
}

// CleanupTurtleContainers stops and removes all turtle containers.
func CleanupTurtleContainers() error {
	// Get all containers.
	containers, err := Client.ListContainers(docker.ListContainersOptions{
		All: true,
	})
	if err != nil {
		return err
	}

	// Stop and remove all containers which start with the turtle prefix.
	var isTurtleContainer bool
	for _, c := range containers {
		// Check if this is a turtle container.
		isTurtleContainer = false
		for _, name := range c.Names {
			if strings.HasPrefix(name, "/"+TurtlePrefix) {
				isTurtleContainer = true
				break
			}
		}

		// Skip if this is not a turtle container.
		if !isTurtleContainer {
			continue
		}

		// Stop and delete the container.
		err = StopAndDeleteContainer(c.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

// StopAndDeleteContainer stops the container and deletes it.
func StopAndDeleteContainer(id string) error {
	// Inspect the container.
	c, err := Client.InspectContainer(id)
	if err != nil {
		return err
	}

	// Stop the container if it is running.
	if c.State.Running || c.State.Paused || c.State.Restarting {
		err = Client.StopContainer(id, killAfterTimeout)
		if err != nil {
			return err
		}
	}

	// Remove the container.
	err = Client.RemoveContainer(docker.RemoveContainerOptions{
		ID:            id,
		RemoveVolumes: false,
		Force:         true,
	})
	if err != nil {
		return err
	}

	return nil
}

// GetContainerByName obtains a container by its name.
func GetContainerByName(name string) (*docker.Container, error) {
	// Get all containers.
	containers, err := Client.ListContainers(docker.ListContainersOptions{
		All: true,
	})
	if err != nil {
		return nil, err
	}

	// Prepend a slash to the name. This is docker specific.
	name = "/" + name

	// Search for the container.
	var found bool
	for _, c := range containers {
		// Check if this is the desired container.
		found = false
		for _, n := range c.Names {
			if n == name {
				found = true
				break
			}
		}

		// Skip if this is not the container.
		if !found {
			continue
		}

		// Inspect the container.
		cc, err := Client.InspectContainer(c.ID)
		if err != nil {
			return nil, err
		}

		// Return the container
		return cc, nil
	}

	return nil, nil
}

// GetContainersByName obtains the containers by their name.
// If no containers were found, no error will be returned. Just a nil slice.
func GetContainersByName(names ...string) ([]*docker.Container, error) {
	// Get all containers.
	containers, err := Client.ListContainers(docker.ListContainersOptions{
		All: true,
	})
	if err != nil {
		return nil, err
	}

	// Prepend a slash to all names. This is docker specific.
	for i := 0; i < len(names); i++ {
		names[i] = "/" + names[i]
	}

	var result []*docker.Container

	// Search for the containers.
	var found bool
	for _, c := range containers {
		// Check if this is the desired container.
		found = false
		for _, name := range names {
			for _, n := range c.Names {
				if n == name {
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		// Skip if this is not a result.
		if !found {
			continue
		}

		// Inspect the container.
		cc, err := Client.InspectContainer(c.ID)
		if err != nil {
			return nil, err
		}

		// Add the container to the result slice.
		result = append(result, cc)
	}

	return result, nil
}

// OnEvent adds the function to the events map and returns its unique ID.
// Use this ID to remove the event again.
func OnEvent(f func(*docker.APIEvents)) int64 {
	// Lock the mutex.
	eventFuncsMutex.Lock()
	defer eventFuncsMutex.Unlock()

	// Increment the counter.
	eventFuncsCounter++

	// Add the event function to the map.
	eventFuncs[eventFuncsCounter] = f

	return eventFuncsCounter
}

// OffEvent removes the event function with the specific ID.
func OffEvent(id int64) {
	// Lock the mutex.
	eventFuncsMutex.Lock()
	defer eventFuncsMutex.Unlock()

	// Remove the event again.
	delete(eventFuncs, id)
}

//###############//
//### Private ###//
//###############//

func startEventListener() error {
	// Create a docker event listener.
	listener := make(chan *docker.APIEvents)
	err := Client.AddEventListener(listener)
	if err != nil {
		return err
	}

	// Start the watch loop in a new goroutine.
	go func() {
		// Remove the event listener on defer again.
		// Currently not triggered...
		defer func() {
			err = Client.RemoveEventListener(listener)
			if err != nil {
				log.Errorf("failed to remove docker event listener: %v", err)
			}
		}()

		for {
			select {
			case event := <-listener:
				func() {
					// Recover panics and log the error.
					defer func() {
						if e := recover(); e != nil {
							log.Errorf("docker event listener panic: event '%v': %v", event.ID, e)
						}
					}()

					// Lock the mutex.
					eventFuncsMutex.Lock()
					defer eventFuncsMutex.Unlock()

					// Call all the event functions.
					for _, f := range eventFuncs {
						f(event)
					}
				}()
			}
		}
	}()

	return nil
}
