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
	AddCommand("info", new(CmdInfo))
}

type CmdInfo struct{}

func (c CmdInfo) Help() string {
	return "Show detailed information about an app."
}

func (c CmdInfo) PrintUsage() {
	fmt.Println("Usage: info APP")
	fmt.Printf("\n%s\n", c.Help())
}

func (c CmdInfo) Run(args []string) error {
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
	request := api.RequestInfo{
		Name: appName,
	}

	// Send the request to the daemon.
	response, err := sendRequest(api.TypeInfo, request)
	if err != nil {
		return err
	}

	// Map the response data to the info value.
	var d api.ResponseInfo
	if err = response.MapTo(&d); err != nil {
		return err
	}

	// Print new lines and a header.
	println("\nGeneral:\n========")

	// Print the general information.
	printc("Name", d.Name)
	printc("State", d.State)
	printc("Turtlefile", d.Turtlefile)
	printc("Maintainer", d.Maintainer)
	printc("SourceURL", d.SourceURL)
	printc("Branch", d.Branch)

	// Print new lines and a header.
	println("\nExposed Ports:\n==============")

	// Print the ports.
	var hp string
	for _, p := range d.Setup.Ports {
		if p.HostPort > 0 {
			hp = strconv.Itoa(p.HostPort)
		} else {
			hp = "DISABLED"
		}

		printf("%v/%s => %s\t%s\n", p.Port, p.Protocol, hp, p.Description)
	}

	// Print new lines and a header.
	println("\nEnvironment variables:\n=====================")

	// Print the environement variables.
	for _, env := range d.Setup.Env {
		if len(env.Description) > 0 {
			printf("%s = %s\t(%s)\n", env.Name, env.Value, env.Description)
		} else {
			printf("%s = %s\t\n", env.Name, env.Value)
		}
	}

	// Flush the output.
	flush()

	// Print a new empty line.
	fmt.Println()

	return nil
}
