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

package turtlefile

import (
	"fmt"
)

//################//
//### Env type ###//
//################//

type Env []*EnvValue

// IsValid checks if required values are missing or invalid.
func (e *Env) IsValid() error {
	for _, env := range *e {
		if len(env.Name) == 0 {
			return fmt.Errorf("Environment name is empty!")
		}
	}

	return nil
}

//#####################//
//### EnvValue type ###//
//#####################//

type EnvValue struct {
	Name string // Environment variable name.

	// Optional
	Alias       []string // A list of aliases. They will be all set as environment variables.
	Required    bool     // If this field is required. Default: false
	Description string   // Setup text.
	Default     string   // Default value of the environment variable.
	Containers  []string // Only set the environment variable in the specific container(s).
}
