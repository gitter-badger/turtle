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
	AddCommand("list", new(CmdList))
}

type CmdList struct{}

func (c CmdList) Help() string {
	return "print a list of all apps including their state."
}

func (c CmdList) Run(args []string) error {
	// Check if any arguments are passed.
	if len(args) > 0 {
		return fmt.Errorf("no arguments expected!")
	}

	// Send the list request to the daemon.
	response, err := sendRequest(api.TypeList, nil)
	if err != nil {
		return err
	}

	// Map the response data to the list value.
	var list api.ResponseList
	if err = response.MapTo(&list); err != nil {
		return err
	}

	// Print a new empty line.
	fmt.Println()

	// Print the column header.
	println("NAME\tTURTLEFILE\tSTATE")

	// Print all the apps.
	for _, app := range list.Apps {
		printc(app.Name, app.Turtlefile, app.State)
	}

	// Flush the output.
	flush()

	// Print a new empty line.
	fmt.Println()

	return nil
}
