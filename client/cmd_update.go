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
	AddCommand("update", new(CmdUpdate))
}

type CmdUpdate struct{}

func (c CmdUpdate) Help() string {
	return "update sources and docker images of an app."
}

func (c CmdUpdate) Run(args []string) error {
	// Check if an argument is passed.
	if len(args) != 1 {
		return fmt.Errorf("the update command requires an app name as first argument.")
	}

	// Obtain the app name.
	appName := strings.TrimSpace(args[0])
	if len(appName) == 0 {
		return fmt.Errorf("invalid app name passed.")
	}

	fmt.Printf("Started process: update app %s.\n", appName)

	// Create a new request.
	request := api.RequestUpdate{
		Name: appName,
	}

	// Send the request to the daemon.
	_, err := sendRequest(api.TypeUpdate, request)
	if err != nil {
		return err
	}

	return nil
}
