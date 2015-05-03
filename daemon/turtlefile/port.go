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

const (
	maxPort = 65535

	ProtocolTCP = "tcp"
	ProtocolUDP = "udp"
)

//##################//
//### Ports type ###//
//##################//

type Ports []*Port

// IsValid checks if required values are missing or invalid.
func (p *Ports) IsValid() error {
	for _, port := range *p {
		// If the protocol is empty, use the default TCP protocol.
		if len(port.Protocol) == 0 {
			port.Protocol = ProtocolTCP
		}

		// Validate
		if port.Port <= 0 || port.Port > maxPort {
			return fmt.Errorf("Invalid port: %v", port.Port)
		} else if port.Protocol != ProtocolTCP && port.Protocol != ProtocolUDP {
			return fmt.Errorf("Invalid port protocol: %s", port.Protocol)
		}

		// The container name has to be specified.
		if len(port.Container) == 0 {
			return fmt.Errorf("Missing container name for port %v!", port.Port)
		}
	}

	return nil
}

//#################//
//### Port type ###//
//#################//

type Port struct {
	Container string
	Port      int

	// Optional
	Protocol    string // tcp or udp. Default is tcp.
	Description string
}
