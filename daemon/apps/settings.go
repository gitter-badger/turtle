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

//##########################//
//### App settings types ###//
//##########################//

type appSettings struct {
	SourceURL string            // The git source URL.
	Branch    string            // Main stable branch.
	Env       map[string]string // The environment values. The key is the name and the value is the variable value.
	Ports     appSettingsPorts
}

// newSettings creates and initializes a new app settings value,
func newSettings() *appSettings {
	return &appSettings{
		Env: make(map[string]string),
	}
}

type appSettingsPorts []*appSettingsPort

type appSettingsPort struct {
	ContainerPort int
	HostPort      int // 0 if disabled.
	Protocol      string
}
