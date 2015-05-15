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
	AddCommand("listb", new(CmdListBackups))
}

type CmdListBackups struct{}

func (c CmdListBackups) Help() string {
	return "Print a list of all backups of an app."
}

func (c CmdListBackups) PrintUsage() {
	fmt.Println("Usage: listb APP")
	fmt.Printf("\n%s\n", c.Help())
}

func (c CmdListBackups) Run(args []string) error {
	// Check if an argument is passed.
	if len(args) != 1 {
		return errInvalidUsage
	}

	// Obtain the app name.
	appName := strings.TrimSpace(args[0])
	if len(appName) == 0 {
		return fmt.Errorf("invalid app name passed.")
	}

	// Create a new request.
	request := api.RequestListBackups{
		Name: appName,
	}

	// Send the list request to the daemon.
	response, err := sendRequest(api.TypeListBackups, request)
	if err != nil {
		return err
	}

	// Map the response data to the list value.
	var list api.ResponseListBackups
	if err = response.MapTo(&list); err != nil {
		return err
	}

	// Check if no backups are present.
	if len(list.Backups) == 0 {
		fmt.Println("There are no backups.")
		return nil
	}

	// Print a new empty line.
	fmt.Println()

	// Print the column header.
	println("DATE\tUNIX TIMESTAMP")

	// Print all the backups.
	for _, b := range list.Backups {
		printc(b.Date, b.Unix)
	}

	// Flush the output.
	flush()

	// Print a new empty line.
	fmt.Println()

	return nil
}
