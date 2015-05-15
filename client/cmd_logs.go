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
	AddCommand("logs", new(CmdLogs))
}

type CmdLogs struct{}

func (c CmdLogs) Help() string {
	return "Fetch the logs of an app."
}

func (c CmdLogs) PrintUsage() {
	fmt.Println("Usage: logs APP [CONTAINER] [STREAM]")
	fmt.Printf("\n%s\n\n", c.Help())
	fmt.Println("Available streams:")
	printc(cmdIndent+"combined", "Combined standard streams. Default option.")
	printc(cmdIndent+"stdout", "Fetch only standard output messages.")
	printc(cmdIndent+"stderr", "Fetch only standard error messages.")
	flush()
}

func (c CmdLogs) Run(args []string) error {
	// Check if an argument is passed.
	if len(args) < 1 || len(args) > 3 {
		return errInvalidUsage
	}

	// Obtain the app name.
	appName := strings.TrimSpace(args[0])
	if len(appName) == 0 {
		return fmt.Errorf("invalid app name passed.")
	}

	// Obtain the container.
	var container string
	if len(args) >= 2 {
		container = strings.TrimSpace(args[1])
	}

	// Obtain the stream.
	var stream string
	if len(args) >= 3 {
		stream = strings.TrimSpace(args[2])
		if stream == "combined" {
			stream = ""
		} else if stream != "stderr" && stream != "stdout" {
			return fmt.Errorf("invalid stream option passed: %s", stream)
		}
	}

	// Create a new request.
	request := api.RequestLogs{
		Name:      appName,
		Container: container,
		Stream:    stream,
	}

	// Send the request to the daemon.
	response, err := sendRequest(api.TypeLogs, request)
	if err != nil {
		return err
	}

	// Map the response data to the logs value.
	var data api.ResponseLogs
	if err = response.MapTo(&data); err != nil {
		return err
	}

	if len(data.Containers) > 0 {
		fmt.Println("Available app containers:\n")

		for i, c := range data.Containers {
			printc(cmdIndent+strconv.Itoa(i+1)+")", c)
		}
		flush()

		fmt.Println()
		c.PrintUsage()

		return nil
	}

	if len(data.LogMessages) == 0 {
		fmt.Println("No log messages available.")
	} else {
		lines := strings.Split(data.LogMessages, "\n")

		for _, l := range lines {
			fmt.Println(cmdIndent + l)
		}
	}

	return nil
}
