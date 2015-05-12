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
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/desertbit/turtle/daemon/config"
	"github.com/desertbit/turtle/utils"
)

func populateKnownHosts() error {
	// Get the turtle known_hosts file path.
	src := config.Config.KnownHostsFilePath()

	// Skip if it does not exists.
	e, err := utils.Exists(src)
	if err != nil {
		return err
	} else if !e {
		return nil
	}

	// Get the current user value.
	usr, err := user.Current()
	if err != nil {
		return err
	}

	// Create the user's known_hosts file path.
	dest := filepath.Clean(usr.HomeDir + "/.ssh/known_hosts")

	// Create the .ssh directory if not present.
	err = utils.MkDirIfNotExists(filepath.Dir(dest))
	if err != nil {
		return err
	}

	// Copy the turtle known_hosts file to the final destination.
	// If the destination exists, it will be overwritten.
	err = utils.CopyFile(src, dest)
	if err != nil {
		return err
	}

	return nil
}

// hostFingerprintExists checks whenever a host fingerprint exists.
func hostFingerprintExists(host string) (bool, error) {
	// Get the know_hosts file path.
	path := config.Config.KnownHostsFilePath()

	// Check if file exists.
	e, err := utils.Exists(path)
	if err != nil {
		return false, err
	} else if !e {
		return false, nil
	}

	// Open the file.
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Read the file line by line.
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Check if the host exists.
		if strings.HasPrefix(strings.TrimSpace(scanner.Text()), host) {
			return true, nil
		}
	}

	// Check if an error occurred.
	if err = scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}

// addHostFingerprint adds a host fingerprint to the known_hosts file.
func addHostFingerprint(fingerprint string) error {
	if len(fingerprint) == 0 {
		return fmt.Errorf("failed to add host fingerprint: empty fingerprint!")
	}

	// Get the know_hosts file path.
	path := config.Config.KnownHostsFilePath()

	err := func() error {
		// Open the file.
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer f.Close()

		// Append the fingerprint.
		if _, err = f.WriteString("\n" + fingerprint); err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		return err
	}

	// Populate the new host file.
	return populateKnownHosts()
}

// getSshHostFingerprint obtains and returns the SSH fingerprint.
func getSshHostFingerprint(host string) (string, error) {
	var stderr, stdout bytes.Buffer

	// Create the command.
	cmd := exec.Command("ssh-keyscan", host)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	// Start the command and wait for it to exit.
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf(strings.TrimSpace(stderr.String()))
	}

	// Get the fingerprint from stdout.
	f := strings.TrimSpace(stdout.String())
	if len(f) == 0 {
		return "", fmt.Errorf("failed to obtain host fingerprint: is this a valid host?!")
	}

	return f, nil
}
