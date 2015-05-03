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

	"github.com/BurntSushi/toml"
)

const (
	TurtlefileFilename = "TURTLE"
)

//#######################//
//### Turtlefile type ###//
//#######################//

type Turtlefile struct {
	Name       string
	Maintainer string

	Env        Env
	Containers Containers `toml:"Container"`
	Ports      Ports      `toml:"Port"`
}

// IsValid checks if required values are missing or invalid.
func (t *Turtlefile) IsValid() error {
	if len(t.Name) == 0 {
		return fmt.Errorf("turtlefile name is missing!")
	}

	// Check if the environment is valid.
	err := t.Env.IsValid()
	if err != nil {
		return err
	}

	// Check if the containers are valid.
	err = t.Containers.IsValid()
	if err != nil {
		return err
	}

	// Check if the ports are valid.
	err = t.Ports.IsValid()
	if err != nil {
		return err
	}

	return nil
}

//##############//
//### Public ###//
//##############//

// Load a turtlefile and return a Turtlefile value,
func Load(turtlefilePath string) (*Turtlefile, error) {
	// Load the app values from the app settings file.
	var t Turtlefile
	_, err := toml.DecodeFile(turtlefilePath, &t)
	if err != nil {
		return nil, fmt.Errorf("failed to load turtlefile '%s': %v", turtlefilePath, err)
	}

	// Sort the containers by their startup level.
	if err = t.Containers.Sort(); err != nil {
		return nil, err
	}

	// Prepare the containers.
	if err = t.Containers.Prepare(); err != nil {
		return nil, err
	}

	return &t, nil
}
