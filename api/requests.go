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

//#####################//
//### Request Types ###//
//#####################//

type Type string

const (
	TypeAdd                 Type = "add"
	TypeRemove              Type = "remove"
	TypeInfo                Type = "info"
	TypeList                Type = "list"
	TypeListBackups         Type = "list-backups"
	TypeStart               Type = "start"
	TypeStop                Type = "stop"
	TypeRestart             Type = "restart"
	TypeSetupGet            Type = "setup-get"
	TypeSetupSet            Type = "setup-set"
	TypeErrorMsg            Type = "error-msg"
	TypeUpdate              Type = "update"
	TypeBackup              Type = "backup"
	TypeRemoveBackup        Type = "remove-backup"
	TypeRestoreBackup       Type = "restore-backup"
	TypeAddHostFingerprint  Type = "add-host-fingerprint"
	TypeHostFingerprintInfo Type = "host-fingerprint-info"
)

//####################//
//### Public Types ###//
//####################//

type RequestAdd struct {
	Name      string
	SourceURL string
	Branch    string
}

type RequestRemove struct {
	Name          string // App name
	RemoveBackups bool   // Whenever all backups should be removed also.
}

type RequestInfo struct {
	Name string // App name
}

type RequestStart struct {
	Name string // App name
}

type RequestStop struct {
	Name string // App name
}

type RequestRestart struct {
	Name string // App name
}

type RequestSetupGet struct {
	Name string // App name
}

type RequestSetupSet struct {
	Name  string // App name
	Setup Setup  // New setup values.
}

type RequestErrorMsg struct {
	Name string // App name
}

type RequestUpdate struct {
	Name string // App name
}

type RequestBackup struct {
	Name string // App name
}

type RequestListBackups struct {
	Name string // App name
}

type RequestRemoveBackup struct {
	Name string // App name
	Unix string // Backup unix timestamp
}

type RequestRestoreBackup struct {
	Name string // App name
	Unix string // Backup unix timestamp
}

type RequestAddHostFingerprint struct {
	Fingerprint string
}

type RequestHostFingerprintInfo struct {
	Host string
}
