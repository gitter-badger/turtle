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
	"sort"

	"github.com/desertbit/turtle/api"
)

func init() {
	// Add this command.
	AddCommand("help", new(CmdHelp))
}

type CmdHelp struct{}

func (c CmdHelp) Help() string {
	return "Show all available turtle commands."
}

func (c CmdHelp) PrintUsage() {
	fmt.Println("Usage: help")
	fmt.Printf("\n%s\n", c.Help())
}

func (c CmdHelp) Run(args []string) error {
	// Check if any arguments are passed.
	if len(args) > 0 {
		return errInvalidUsage
	}

	// Get all command map keys.
	i := 0
	keys := make([]string, len(commands))
	for k, _ := range commands {
		keys[i] = k
		i++
	}

	// Sort the keys slice.
	sort.Strings(keys)

	// Header.
	fmt.Printf("\nTURTLE %s - Rock Solid Cluster Management\n", api.Version)
	printTurtle()
	fmt.Println("Available commands:\n")

	// Print all available commands with a description.
	for _, key := range keys {
		printc(cmdIndent+key, commands[key].Help())
	}

	// Flush the output.
	flush()

	// Print a new empty line.
	fmt.Println()

	return nil
}
