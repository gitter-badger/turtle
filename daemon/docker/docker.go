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
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/desertbit/turtle/daemon/config"

	log "github.com/Sirupsen/logrus"
	docker "github.com/fsouza/go-dockerclient"
)

const (
	TurtlePrefix = "turtle."

	killAfterTimeout = 10 // in seconds

	imageBuildTag = "turtle-build"
	imageOldTag   = "turtle-old"
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

// Logs obtains the stderr and stdout log messages from the container.
func Logs(containerID string, getStdout bool, getStderr bool) (stdout string, stderr string, err error) {
	// Create the output buffers.
	outBuf := bytes.NewBuffer(nil)
	errBuf := bytes.NewBuffer(nil)

	// Create the logs options.
	opts := docker.LogsOptions{
		Container:    containerID,
		Follow:       false,
		OutputStream: outBuf,
		ErrorStream:  errBuf,
		Stdout:       getStdout,
		Stderr:       getStderr,
	}

	// Obtain the logs.
	err = Client.Logs(opts)
	if err != nil {
		err = fmt.Errorf("failed to get container '%s' logs: %v", containerID, err)
		return
	}

	// Set the strings.
	stdout = strings.TrimSpace(outBuf.String())
	stderr = strings.TrimSpace(errBuf.String())

	return
}

// Build a docker image from a local directory.
func Build(imageName, tag, dir string) error {
	if len(imageName) == 0 || len(tag) == 0 || len(dir) == 0 {
		return fmt.Errorf("build docker image: invalid arguments!")
	}

	// Create a buffer to write our archive to.
	buf := bytes.NewBuffer(nil)

	err := func() (err error) {
		// Create a new tar archive.
		tw := tar.NewWriter(buf)

		// Check the error on Close.
		defer func() {
			if err != nil {
				tw.Close()
			} else {
				err = tw.Close()
			}
		}()

		// Based on: https://stackoverflow.com/questions/13611100/how-to-write-a-directory-not-just-the-files-in-it-to-a-tar-gz-file-in-golang
		walkFn := func(path string, info os.FileInfo, err error) error {
			// Handle the filepath walk error.
			if err != nil {
				return err
			}
			// SKip hidden files and hidden directories.
			if strings.HasPrefix(info.Name(), ".") {
				if info.Mode().IsDir() {
					return filepath.SkipDir
				} else {
					return nil
				}
			}
			// Skip directories. The directory contents are added if there are any.
			if info.Mode().IsDir() {
				return nil
			}
			// Because of scoping we can reference the external dir variable
			new_path := path[len(dir):]
			if len(new_path) == 0 {
				return nil
			}
			fr, err := os.Open(path)
			if err != nil {
				return err
			}
			defer fr.Close()

			if h, err := tar.FileInfoHeader(info, new_path); err != nil {
				return err
			} else {
				h.Name = new_path
				if err = tw.WriteHeader(h); err != nil {
					return err
				}
			}
			if _, err = io.Copy(tw, fr); err != nil {
				return err
			}
			return nil
		}

		if err = filepath.Walk(dir, walkFn); err != nil {
			return err
		}

		return nil
	}()
	if err != nil {
		return fmt.Errorf("failed to tar files in '%s': %v", dir, err)
	}

	// Create the build image name with tag.
	buildImage := imageName + ":" + imageBuildTag

	// Function to remove the build image if present.
	removeBuildImage := func() error {
		// Check if a previous build image exists.
		// Remove it if present.
		_, err := Client.InspectImage(buildImage)
		if err == nil {
			err = Client.RemoveImage(buildImage)
			if err != nil {
				return fmt.Errorf("failed to remove previous build image '%s': %v", buildImage, err)
			}
		}
		return nil
	}

	// Remove the build image if present.
	err = removeBuildImage()
	if err != nil {
		return err
	}

	// Create the output buffer.
	outputbuf := bytes.NewBuffer(nil)

	// Create the build options.
	opts := docker.BuildImageOptions{
		Name:                buildImage,
		NoCache:             true,
		Pull:                true,
		RmTmpContainer:      true,
		ForceRmTmpContainer: true,
		InputStream:         buf,
		OutputStream:        outputbuf,
	}

	// Build the image.
	if err := Client.BuildImage(opts); err != nil {
		return fmt.Errorf("%v\nbuild output:\n%s", err, outputbuf.String())
	}

	// Remove the build image on defer.
	// The error is not important.
	defer func() {
		err := removeBuildImage()
		if err != nil {
			log.Errorf("failed to remove temporary build image '%s'!", buildImage)
		}
	}()

	// Check if the old image exists.
	// Remove it if present.
	oldImage := imageName + ":" + imageOldTag
	_, err = Client.InspectImage(oldImage)
	if err == nil {
		err = Client.RemoveImage(oldImage)
		if err != nil {
			return fmt.Errorf("failed to remove old image '%s': %v", oldImage, err)
		}
	}

	// Tag the current image to the old image tag if it exists.
	image := imageName + ":" + tag
	_, err = Client.InspectImage(image)
	if err == nil {
		opts := docker.TagImageOptions{
			Repo:  imageName,
			Tag:   imageOldTag,
			Force: true,
		}

		err = Client.TagImage(image, opts)
		if err != nil {
			return fmt.Errorf("failed to tag current image '%s': %v", image, err)
		}
	}

	// Tag the build image to the new current image.
	optsT := docker.TagImageOptions{
		Repo:  imageName,
		Tag:   tag,
		Force: true,
	}

	err = Client.TagImage(buildImage, optsT)
	if err != nil {
		return fmt.Errorf("failed to tag build image '%s': %v", buildImage, err)
	}

	return nil
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
