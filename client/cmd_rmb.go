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
	AddCommand("rmb", new(CmdRmb))
}

type CmdRmb struct{}

func (c CmdRmb) Help() string {
	return "remove an app's backup."
}

func (c CmdRmb) Run(args []string) error {
	// Check if an argument is passed.
	if len(args) != 2 {
		return fmt.Errorf("the rmb command requires the app name as first argument and the backup timestamp as second argument.")
	}

	// Obtain the app name.
	name := strings.TrimSpace(args[0])
	if len(name) == 0 {
		return fmt.Errorf("invalid app name passed.")
	}

	// Obtain the timestamp.
	unix := strings.TrimSpace(args[1])
	if len(unix) == 0 {
		return fmt.Errorf("invalid backup timestamp passed.")
	}

	fmt.Printf("Remove backup '%s'?\n", unix)

	// Confirm the request.
	if !confirmCommit() {
		return nil
	}

	// Create a new remove request.
	request := api.RequestRemoveBackup{
		Name: name,
		Unix: unix,
	}

	// Send the remove request to the daemon.
	_, err := sendRequest(api.TypeRemoveBackup, request)
	if err != nil {
		return err
	}

	fmt.Println("Successfully removed backup.")

	return nil
}
