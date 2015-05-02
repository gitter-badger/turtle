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
)

func init() {
	// Add this command.
	AddCommand("add", new(CmdAdd))
}

type CmdAdd struct{}

func (c CmdAdd) Help() string {
	return "add a new app."
}

func (c CmdAdd) Run(args []string) error {
	// Check if any arguments are passed.
	if len(args) > 0 {
		return fmt.Errorf("no arguments expected!")
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
	fmt.Print("branch [master]: ")
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
