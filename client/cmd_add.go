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
	"fmt"

	"github.com/desertbit/turtle/api"
	"github.com/desertbit/turtle/utils"
)

func init() {
	// Add this command.
	AddCommand("add", new(CmdAdd))
}

type CmdAdd struct{}

func (c CmdAdd) Help() string {
	return "Add a new app."
}

func (c CmdAdd) PrintUsage() {
	fmt.Println("Usage: add")
	fmt.Printf("\n%s\n", c.Help())
}

func (c CmdAdd) Run(args []string) error {
	// Check if any arguments are passed.
	if len(args) > 0 {
		return errInvalidUsage
	}

	fmt.Println("Add a new app")

	// Get the app name.
	fmt.Print("Name: ")
	name, err := readline()
	if err != nil {
		return err
	}

	// Get the app source url.
	fmt.Print("Source URL: ")
	sourceURL, err := readline()
	if err != nil {
		return err
	}

	// Get the app branch.
	fmt.Print("Branch [master]: ")
	branch, err := readline("master")
	if err != nil {
		return err
	}

	// Valdiate.
	if len(name) == 0 || len(sourceURL) == 0 || len(branch) == 0 {
		return fmt.Errorf("invalid or emtpy option(s)!")
	}

	// Confirm the request.
	if !confirmCommit() {
		return nil
	}

	// Add the host fingerprint if not trusted yet.
	if err = c.HandleHostFingerprint(sourceURL); err != nil {
		return err
	}

	// Create a new add request.
	request := api.RequestAdd{
		Name:      name,
		SourceURL: sourceURL,
		Branch:    branch,
	}

	// Send the add request to the daemon.
	_, err = sendRequest(api.TypeAdd, request)
	if err != nil {
		return err
	}

	fmt.Println("Successfully added new app.")

	return nil
}

func (c CmdAdd) HandleHostFingerprint(sourceURL string) error {
	host := utils.GetHostFromUrl(sourceURL)

	// Create a new info request.
	request := api.RequestHostFingerprintInfo{
		Host: host,
	}

	// Send the info request to the daemon.
	response, err := sendRequest(api.TypeHostFingerprintInfo, request)
	if err != nil {
		return err
	}

	// Map the response data to the info value.
	var info api.ResponseHostFingerprintInfo
	if err = response.MapTo(&info); err != nil {
		return err
	}

	// Check if the host fingerprint is already trusted.
	if info.Trusted {
		return nil
	}

	// Ask the user if he trusts the given fingerprint.
	fmt.Printf("Host fingerprint is missing.\nFingerprint of host '%s':\n\n", host)
	fmt.Print(colorHint, info.Fingerprint, colorOutput)
	fmt.Println("\n\nTrust it and add it?")

	// Confirm the request.
	if !confirmCommit() {
		return fmt.Errorf("aborted: host fingerprint not trusted by user.")
	}

	// Create the request to add the fingerprint.
	rAdd := api.RequestAddHostFingerprint{
		Fingerprint: info.Fingerprint,
	}

	// Send the request to the daemon.
	response, err = sendRequest(api.TypeAddHostFingerprint, rAdd)
	if err != nil {
		return err
	}

	return nil
}
