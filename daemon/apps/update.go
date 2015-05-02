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

	d "github.com/fsouza/go-dockerclient"
)

//####################//
//### Clone Source ###//
//####################//

// Update the app.
func (a *App) Update() error {
	// Abort if the app is running.
	if a.IsRunning() {
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

	// Update all docker images.
	for _, container := range turtlefile.Containers {
		// Pull the image.
		err = docker.Client.PullImage(d.PullImageOptions{
			Repository: container.Image,
		}, d.AuthConfiguration{})

		if err != nil {
			return fmt.Errorf("failed to pull docker image '%s': %v", container.Image, err)
		}
	}

	return nil
}
