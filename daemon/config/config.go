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

package config

import (
	"time"
)

const (
	TurtleRoot = "/turtle"
)

// TODO: Load the config from the turtle directory.

var (
	Config = config{
		ListenAddress:  ":28239",
		DockerEndPoint: "unix:///var/run/docker.sock",

		AppPath:    TurtleRoot + "/apps",
		BackupPath: TurtleRoot + "/backups",
		TurtlePath: TurtleRoot + "/turtle",

		BtrfsBalanceInterval: 3 * time.Hour,
		BtrfsBalanceDusage:   20,

		BackupInterval: 4 * time.Hour,
	}
)

type config struct {
	ListenAddress  string
	DockerEndPoint string

	AppPath    string
	BackupPath string
	TurtlePath string

	BtrfsBalanceInterval time.Duration
	BtrfsBalanceDusage   int // In percent

	BackupInterval time.Duration // Create backups of running apps in this interval.
}

// StateFilePath returns the turtle state file path.
func (c *config) StateFilePath() string {
	return c.TurtlePath + "/state"
}
