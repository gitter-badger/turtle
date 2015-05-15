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
)

func init() {
	// Add this command.
	AddCommand("exit", new(CmdExit))
}

type CmdExit struct{}

func (c CmdExit) Help() string {
	return "Exit the turtle shell."
}

func (c CmdExit) PrintUsage() {
	fmt.Println("Usage: exit")
	fmt.Printf("\n%s\n", c.Help())
}

func (c CmdExit) Run(args []string) error {
	// Check if any arguments are passed.
	if len(args) > 0 {
		return errInvalidUsage
	}

	// Exit the application.
	cleanupAndExit()

	return nil
}
