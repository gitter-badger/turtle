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

	"github.com/desertbit/turtle/utils"
)

//####################//
//### Clone Source ###//
//####################//

func (a *App) cloneSource() error {
	// Start the task to clone the source.
	return a.runTask(taskCloneSource, taskFuncCloneSource)
}

func taskFuncCloneSource(app *App) error {
	app.setState("cloning source")

	// Clone the source with git.
	err := utils.RunCommand("git", "clone", "-b", app.settings.Branch, "--single-branch", app.settings.SourceURL, app.SourceDirectoryPath())
	if err != nil {
		return fmt.Errorf("failed to clone application source with git: %v", err)
	}

	return nil
}
