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

package apps

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
)

const (
	stateError        = "error"
	stateIdle         = "stopped"
	stateStartingTask = "starting task..."
)

//##################//
//### task types ###//
//##################//

type taskType int

const (
	taskNone        taskType = 1 << iota
	taskCloneSource taskType = 1 << iota
	taskRun         taskType = 1 << iota
	taskUpdate      taskType = 1 << iota
	taskBackup      taskType = 1 << iota
)

//########################//
//### App task methods ###//
//########################//

// IsTaskRunning returns a boolean whenever a task is active and running.
func (a *App) IsTaskRunning() bool {
	return a.task != taskNone
}

// Error returns the last error of the last task.
// nil is returned, if no error occurred.
func (a *App) Error() error {
	return a.taskErr
}

// State returns the app's current state.
func (a *App) State() string {
	return a.taskState
}

// setState sets the task state.
func (a *App) setState(s string) {
	a.taskState = s
}

// runTask runs a task function in a new goroutine and waits for it to exit.
func (a *App) runTask(t taskType, f func(*App) error) error {
	// Lock the mutex.
	a.taskMutex.Lock()
	defer a.taskMutex.Unlock()

	// Check if another task is already set.
	if a.task != taskNone {
		return fmt.Errorf("Another task is already running...")
	}

	// Set the task and reset the task error.
	a.task = t
	a.taskErr = nil

	// Set the initial state.
	a.setState(stateStartingTask)

	// Run it in a new goroutine.
	go func() {
		// Recover panics and log the error.
		defer func() {
			if e := recover(); e != nil {
				log.Errorf("task panic: type %v: %v", t, e)
			}
		}()

		// Run the task.
		err := f(a)

		// Lock the mutex.
		a.taskMutex.Lock()
		defer a.taskMutex.Unlock()

		// Reset the app's task.
		a.task = taskNone

		// Check the error.
		if err != nil {
			// Set the error state.
			a.setState(stateError)

			// Set the error.
			a.taskErr = fmt.Errorf("task failed: %v", err)

			// Log the error.
			log.Errorf("app '%s': %v", a.name, a.taskErr)
		} else {
			// Reset the state.
			a.setState(stateIdle)
		}
	}()

	return nil
}
