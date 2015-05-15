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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/desertbit/turtle/api"

	gnureadline "github.com/bobappleyard/readline"
	"github.com/mgutz/ansi"
)

const (
	cmdIndent = "  "
)

//#####################//
//### Access Groups ###//
//#####################//

type Group int
type Groups []Group

//####################//
//### Command type ###//
//####################//

var (
	commands        = make(map[string]Command)
	errInvalidUsage = errors.New("invalid usage.")
)

type Command interface {
	Run(args []string) error
	Help() string
	PrintUsage()
}

func AddCommand(key string, command Command) {
	commands[key] = command
}

//###############//
//### Private ###//
//###############//

var (
	// Terminal ANSI colors.
	colorInput  = ansi.ColorCode("white")
	colorPrompt = ansi.ColorCode("green+bh")
	colorOutput = ansi.ColorCode("white+h")
	colorError  = ansi.ColorCode("red+bh")
	colorHint   = ansi.ColorCode("blue+h")
	colorReset  = ansi.ColorCode("reset")

	// Create the reader from stdin.
	reader = bufio.NewReader(os.Stdin)

	// Network
	host string = "127.0.0.1"
	port string = "28239"

	// Prompt
	prompt string = colorPrompt + "[turtle]$ " + colorReset + colorInput
)

func main() {
	// Set the maximum number of CPUs that can be executing simultaneously.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Catch interrupts.
	go onInterrupt()

	// Print our cute turtle.
	printTurtle()

	for {
		readCommand()
	}
}

func onInterrupt() {
	// Wait for the signal
	sigchan := make(chan os.Signal, 10)
	signal.Notify(sigchan, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGKILL)
	<-sigchan

	// Exit the application
	cleanupAndExit()
}

func cleanupAndExit() {
	// Restore terminal attributes..
	gnureadline.Cleanup()

	// Reset the terminal color.
	fmt.Print(colorReset)

	fmt.Println(`Good Bye

  ___/')
 /= =\/
/= = =\
^--|--^
`)

	// Just wait for a moment before exiting to be
	// sure, all defers get called and the program
	// performs a clean exit.
	time.Sleep(150 * time.Millisecond)

	// Exit the application with exit code 1.
	os.Exit(1)
}

func readCommand() {
	// Reset the color on defer.
	defer fmt.Print(colorReset)

	// Read the input line.
	line, err := gnureadline.String(prompt)
	if err == io.EOF {
		// Print a new line.
		fmt.Println("")
		return
	} else if err != nil {
		fmt.Printf("%serror: %v\n", colorError, err)
		return
	}

	// Set to output color.
	fmt.Print(colorOutput)

	// Trim all spaces.
	line = strings.TrimSpace(line)

	// Split the input line.
	args := strings.Fields(line)

	// Skip if empty.
	if len(args) == 0 {
		return
	}

	// Get the command key.
	key := args[0]

	// Remove the first command key from the slice.
	args = args[1:]

	// Try to find the command in the commands map.
	cmd, ok := commands[key]
	if !ok {
		fmt.Printf("%serror: invalid command\n", colorError)
		return
	}

	// Add the command to the history.
	gnureadline.AddHistory(line)

	// Run the command.
	err = cmd.Run(args)
	if err != nil {
		fmt.Printf("%serror: %v\n", colorError, err)

		// Print the usage if this is an invalid usage error.
		if err == errInvalidUsage {
			// Set to output color.
			fmt.Print(colorReset, colorOutput, "\n")

			cmd.PrintUsage()
		}

		return
	}
}

func printTurtle() {
	fmt.Println(`      ___
 ,,  // \\
(_,\/ \_/ \
  \ \_/_\_/>
  /_/  /_/
`)
}

// sendRequest sends a request to the daemon server.
// If a remote error occurres, the error value will be extracted and returned as error.
func sendRequest(requestType api.Type, data interface{}) (*api.Response, error) {
	// Create a new request value.
	request := api.NewRequest(requestType, data)

	// Marshal the request to JSON.
	json, err := request.ToJSON()
	if err != nil {
		return nil, err
	}

	// Create a new io.Reader from the JSON.
	jsonReader := bytes.NewReader(json)

	// Create a new HTTP request.
	req, _ := http.NewRequest("POST", "http://"+host+":"+port, jsonReader)
	req.Header.Set("Content-Type", "application/json")

	// Get the HTTP client and perform the request.
	client := &http.Client{}
	httpResponse, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Marshal the received JSON response to a response value.
	response, err := api.NewResponseFromJSON(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	// The API versions have to match.
	if request.Version != api.Version {
		return nil, fmt.Errorf("API Versions don't match: client=%s server=%s", request.Version, api.Version)
	}

	// Check if an error occurred.
	if response.Status == api.StatusError {
		// Map the error response value.
		var rErr api.ResponseError
		if err = response.MapTo(&rErr); err != nil {
			return nil, err
		}

		// Create an error from the error message string.
		return nil, fmt.Errorf(rErr.ErrorMessage)
	}

	return response, nil
}
