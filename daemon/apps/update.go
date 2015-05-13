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

	"github.com/desertbit/turtle/daemon/docker"
	"github.com/desertbit/turtle/utils"

	log "github.com/Sirupsen/logrus"
	d "github.com/fsouza/go-dockerclient"
)

//####################//
//### Clone Source ###//
//####################//

// Update the app.
func (a *App) Update() error {
	// Abort if any app task is running.
	if a.IsTaskRunning() {
		return fmt.Errorf("the app is running!")
	}

	// Create a backup.
	err := a.Backup()
	if err != nil {
		return err
	}

	// Start the task to update the source.
	return a.runTask(taskUpdate, taskFuncUpdate)
}

func taskFuncUpdate(app *App) error {
	app.setState("updating")

	// Pull the source with git.
	err := utils.RunCommandInPath(app.SourceDirectoryPath(), "git", "pull")
	if err != nil {
		return fmt.Errorf("failed to pull latest application source with git: %v", err)
	}

	// Get and update the turtlefile.
	app.turtlefile = nil
	turtlefile, err := app.Turtlefile()
	if err != nil {
		return err
	}

	// Get the container name prefix.
	cNamePrefix := app.ContainerNamePrefix()

	// Get the app's source path.
	sourcePath := app.SourceDirectoryPath()

	// Update all docker images.
	for _, container := range turtlefile.Containers {
		isLocalBuild := container.IsLocalBuild()

		// Create the docker container name.
		containerName := cNamePrefix + container.Name

		// Create the container image and image name.
		imageName := container.Image
		if isLocalBuild {
			imageName = containerName
		}
		image := imageName + ":" + container.Tag

		// Check whenever to build or pull the image.
		if isLocalBuild {
			app.setState("building local docker image: " + image)
			log.Infof("building local docker image: %s", image)

			// Build the local image.
			err = docker.Build(imageName, container.Tag, container.BuildPath(sourcePath))
			if err != nil {
				return fmt.Errorf("failed to build image '%s': %v", image, err)
			}
		} else {
			app.setState("pulling docker image: " + image)
			log.Infof("pulling docker image: %s", image)

			// Pull the image.
			err = docker.Client.PullImage(d.PullImageOptions{
				Repository: container.Image,
				Tag:        container.Tag,
			}, d.AuthConfiguration{})

			if err != nil {
				return fmt.Errorf("failed to pull docker image '%s': %v", image, err)
			}
		}
	}

	return nil
}
