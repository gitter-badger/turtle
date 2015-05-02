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

package api

//###################//
//### setup types ###//
//###################//

type Setup struct {
	Env   Env
	Ports Ports
}

type Env []*EnvValue

type EnvValue struct {
	Name string // Environment variable name.

	// Optional
	Value       string // Already set value if present.
	Required    bool   // If this field is required. Default: false
	Description string // Setup text.
	Default     string // Default value of the environment variable.
}

type Ports []*Port

type Port struct {
	Port     int
	Protocol string // tcp or udp. Default is tcp.

	// Optional
	HostPort    int
	Description string
}
