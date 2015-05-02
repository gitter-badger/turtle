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
	"strconv"
	"strings"

	"github.com/desertbit/turtle/api"
)

func init() {
	// Add this command.
	AddCommand("setup", new(CmdSetup))
}

type CmdSetup struct{}

func (c CmdSetup) Help() string {
	return "setup an app."
}

func (c CmdSetup) Run(args []string) error {
	// Check if an argument is passed.
	if len(args) != 1 {
		return fmt.Errorf("the setup command requires an app name as first argument.")
	}

	// Obtain the app name.
	appName := strings.TrimSpace(args[0])
	if len(appName) == 0 {
		return fmt.Errorf("invalid app name passed.")
	}

	// Create a new setup request.
	request := api.RequestSetupGet{
		Name: appName,
	}

	// Send the list request to the daemon.
	response, err := sendRequest(api.TypeSetupGet, request)
	if err != nil {
		return err
	}

	// Map the response data to the setup value.
	var setup api.Setup
	if err = response.MapTo(&setup); err != nil {
		return err
	}

	// Get the environement values from the user.
	for _, env := range setup.Env {
		var defaultValue string

		// Set to hint color.
		fmt.Print(colorHint)

		// Print the description if present.
		if len(env.Description) > 0 {
			fmt.Printf("Description of %s: %s\n", env.Name, env.Description)
		}
		// Print the default value if present.
		if len(env.Default) > 0 {
			fmt.Printf("Default value: %s\n", env.Default)
			defaultValue = env.Default
		}
		// Print the current set value if present.
		if len(env.Value) > 0 {
			fmt.Printf("Current value: %s\n", env.Value)
			defaultValue = env.Value
		}

		// Set to output color.
		fmt.Print(colorOutput)

		for {
			fmt.Printf("> %s [%s]: ", env.Name, defaultValue)

			// Get the user value.
			value, err := readline(defaultValue)
			if err != nil {
				return err
			}

			if env.Required && len(value) == 0 {
				fmt.Println("This value is required!")
				continue
			}

			// Set the new value.
			env.Value = value

			// Print a new line.
			fmt.Println()

			break
		}
	}

	// Get the exposed port values from the user.
	for _, p := range setup.Ports {
		defaultValue := "!"

		// Set to hint color.
		fmt.Print(colorHint)

		// Print the description if present.
		if len(p.Description) > 0 {
			fmt.Printf("Description of container port %v/%s: %s\n", p.Port, p.Protocol, p.Description)
		}
		// Print the current set host port if present.
		if p.HostPort > 0 {
			fmt.Printf("Current host port: %v\n", p.HostPort)
			defaultValue = strconv.Itoa(p.HostPort)
		} else {
			fmt.Println("Host Port is currently disabled.")
		}

		fmt.Println("Enter ! to disable a port.")

		// Set to output color.
		fmt.Print(colorOutput)

		for {
			fmt.Printf("> Expose port %v/%s to host port [%v]: ", p.Port, p.Protocol, defaultValue)

			// Get the user value.
			value, err := readline(defaultValue)
			if err != nil {
				return err
			}

			// If disabled, then disable the port (=0).
			if value == "!" {
				p.HostPort = 0
			} else {
				// Convert to int.
				port, err := strconv.Atoi(value)
				if err != nil || port < 0 {
					fmt.Println("invalid port value!")
					continue
				}

				// Set the new host port value.
				p.HostPort = port
			}

			// Print a new line.
			fmt.Println()

			break
		}
	}

	// Confirm the request.
	if !confirmCommit() {
		return nil
	}

	// Create a new setup request.
	sRequest := api.RequestSetupSet{
		Name:  appName,
		Setup: setup,
	}

	// Send the setup request to the daemon.
	_, err = sendRequest(api.TypeSetupSet, sRequest)
	if err != nil {
		return err
	}

	fmt.Println("Successfully configured app.")

	return nil
}
