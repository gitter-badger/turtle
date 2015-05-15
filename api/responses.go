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

//####################//
//### Public Types ###//
//####################//

type ResponseInfo struct {
	Name       string
	Turtlefile string
	Maintainer string
	State      string
	SourceURL  string
	Branch     string

	Setup *Setup
}

type ResponseList struct {
	Apps []ResponseListApp
}

type ResponseListApp struct {
	Name       string
	Turtlefile string
	State      string
}

type ResponseListBackups struct {
	Backups []ResponseListBackup
}

type ResponseListBackup struct {
	Date string
	Unix string
}

type ResponseErrorMsg struct {
	Name         string
	ErrorMessage string
}

type ResponseHostFingerprintInfo struct {
	Host        string
	Trusted     bool
	Fingerprint string
}

type ResponseLogs struct {
	Containers  []string // Empty if a specific continer is passed. Otherwise a list of available containers is set.
	LogMessages string
}
