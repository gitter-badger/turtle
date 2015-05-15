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
	"strings"

	"github.com/desertbit/turtle/api"
)

func init() {
	// Add this command.
	AddCommand("rm", new(CmdRm))
}

type CmdRm struct{}

func (c CmdRm) Help() string {
	return "Remove an app with all data."
}

func (c CmdRm) PrintUsage() {
	fmt.Println("Usage: rm APP")
	fmt.Printf("\n%s\n", c.Help())
}

func (c CmdRm) Run(args []string) error {
	// Check if an argument is passed.
	if len(args) != 1 {
		return errInvalidUsage
	}

	// Obtain the app name.
	appName := strings.TrimSpace(args[0])
	if len(appName) == 0 {
		return fmt.Errorf("invalid app name passed.")
	}

	// Check whenever to remove all backups also.
	var removeBackups bool
	for {
		fmt.Print("Remove all backups also? (y/N) ")
		input, err := readline("n")
		if err != nil {
			return err
		}
		input = strings.ToLower(input)

		if input == "y" {
			removeBackups = true
			break
		} else if input == "n" {
			removeBackups = false
			break
		}

		continue
	}

	if removeBackups {
		fmt.Println("All backups will be deleted!")
	} else {
		fmt.Println("A new backup will be created before deletion.")
	}

	fmt.Printf("Remove app '%s'?\n", appName)

	// Confirm the request.
	if !confirmCommit() {
		return nil
	}

	// Create a new remove request.
	request := api.RequestRemove{
		Name:          appName,
		RemoveBackups: removeBackups,
	}

	// Send the remove request to the daemon.
	_, err := sendRequest(api.TypeRemove, request)
	if err != nil {
		return err
	}

	fmt.Println("Successfully removed app.")

	return nil
}
