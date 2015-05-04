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

	"github.com/desertbit/turtle/utils"
)

func getKnownHostsFilePath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	return filepath.Clean(usr.HomeDir + "/.ssh/known_hosts"), nil
}

// hostFingerprintExists checks whenever a host fingerprint exists.
func hostFingerprintExists(host string) (bool, error) {
	// Get the know_hosts file path.
	path, err := getKnownHostsFilePath()
	if err != nil {
		return false, err
	}

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
	path, err := getKnownHostsFilePath()
	if err != nil {
		return err
	}

	// Create the .ssh directory if not present.
	err = utils.MkDirIfNotExists(filepath.Dir(path))
	if err != nil {
		return err
	}

	// Open the file.
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	// Append the fingerprint.
	if _, err = f.WriteString("\n" + fingerprint); err != nil {
		return err
	}

	return nil
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
