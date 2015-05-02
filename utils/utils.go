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

package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// RunCommand runs a command and waits for it to exit.
// The stderr error message is returned on error.
func RunCommand(name string, args ...string) error {
	return RunCommandInPath("", name, args...)
}

// RunCommandInPath runs a command and waits for it to exit.
// The working directory is also set.
// The stderr error message is returned on error.
func RunCommandInPath(dir, name string, args ...string) error {
	// Create the command.
	var stderr bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stderr = &stderr
	cmd.Dir = dir

	// Start the command and wait for it to exit.
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf(strings.TrimSpace(stderr.String()))
	}

	return nil
}
