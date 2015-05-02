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
	AddCommand("error", new(CmdError))
}

type CmdError struct{}

func (c CmdError) Help() string {
	return "show the last error message of an app."
}

func (c CmdError) Run(args []string) error {
	// Check if an argument is passed.
	if len(args) != 1 {
		return fmt.Errorf("the error command requires an app name as first argument.")
	}

	// Obtain the app name.
	appName := strings.TrimSpace(args[0])
	if len(appName) == 0 {
		return fmt.Errorf("invalid app name passed.")
	}

	// Create a new request.
	request := api.RequestErrorMsg{
		Name: appName,
	}

	// Send the request to the daemon.
	response, err := sendRequest(api.TypeErrorMsg, request)
	if err != nil {
		return err
	}

	// Map the response data to the error message value.
	var data api.ResponseErrorMsg
	if err = response.MapTo(&data); err != nil {
		return err
	}

	// Set to hint color.
	fmt.Print(colorWarning)

	if len(data.ErrorMessage) == 0 {
		fmt.Println("No errors occurred :)")
	} else {
		fmt.Printf("Error message:\n%s\n", data.ErrorMessage)
	}

	// Set to output color.
	fmt.Print(colorOutput)

	return nil
}
