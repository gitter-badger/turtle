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
	"os"
	"strings"
	"text/tabwriter"
)

var (
	// Tab writer
	tabWriterStdout *tabwriter.Writer
)

func init() {
	// Create a new tab writer to format the output.
	tabWriterStdout = new(tabwriter.Writer)

	// Format in tab-separated columns with a tab stop of 8.
	tabWriterStdout.Init(os.Stdout, 0, 8, 0, '\t', 0)
}

func print(a ...interface{}) {
	fmt.Fprint(tabWriterStdout, a...)
}

func println(a ...interface{}) {
	fmt.Fprintln(tabWriterStdout, a...)
}

func printf(format string, a ...interface{}) {
	fmt.Fprintf(tabWriterStdout, format, a...)
}

// printc prints all the variadic arguments as columns.
func printc(a ...interface{}) {
	// Create the format string
	var f string
	for range a {
		f += "%v\t"
	}
	f += "\n"

	// Print the columns.
	printf(f, a...)
}

func flush() {
	tabWriterStdout.Flush()
}

// readline reads a line from stdin and trims the result.
// If the result is empty, then the default value is used if defined.
func readline(defaultValue ...string) (string, error) {
	// Readline.
	in, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	// Trim spaces.
	in = strings.TrimSpace(in)

	// Set the default value if empty.
	if len(in) == 0 && len(defaultValue) > 0 {
		in = defaultValue[0]
	}

	return in, nil
}

// confirmCommit asks the user to confirm his request.
func confirmCommit() bool {
	for {
		// Get the app name.
		fmt.Print("Continue and commit the request? (y/n): ")
		q, err := readline()
		if err != nil {
			fmt.Printf("error: %v", err)
			return false
		}

		if q == "y" {
			return true
		} else if q == "n" {
			return false
		}

		fmt.Println("invalid option!")
	}
}
