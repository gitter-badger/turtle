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
	"github.com/desertbit/turtle/api"

	log "github.com/Sirupsen/logrus"
)

//#########################//
//### App setup methods ###//
//#########################//

// IsSetup returns a boolean if the app is setup.
func (a *App) IsSetup() bool {
	// Get the turtlefile.
	t, err := a.Turtlefile()
	if err != nil {
		log.Warningf("app '%s': isSetup: %v", a.name, err)
		return false
	}

	// Check if a required environment variable is unset.
	for _, env := range t.Env {
		// Skip not optional variables.
		if !env.Required {
			continue
		}

		// Check if it is set.
		v, ok := a.settings.Env[env.Name]
		if !ok || len(v) == 0 {
			return false
		}
	}

	return true
}

// GetSetup returns a API setup value.
func (a *App) GetSetup() (*api.Setup, error) {
	// Get the turtlefile.
	t, err := a.Turtlefile()
	if err != nil {
		return nil, err
	}

	// Create a new setup value.
	setup := &api.Setup{
		Env:   make(api.Env, len(t.Env)),
		Ports: make(api.Ports, len(t.Ports)),
	}

	// Fill the setup environment values,
	i := 0
	for _, env := range t.Env {
		// Create a new setup environment value.
		e := &api.EnvValue{
			Name:        env.Name,
			Required:    env.Required,
			Description: env.Description,
			Default:     env.Default,
		}

		// Get the environment value from the settings.
		for name, value := range a.settings.Env {
			if e.Name == name {
				e.Value = value
				break
			}
		}

		// Add it to the setup.
		setup.Env[i] = e
		i++
	}

	// Fill the setup ports.
	i = 0
	for _, port := range t.Ports {
		// Create a new port value.
		p := &api.Port{
			Port:        port.Port,
			Protocol:    port.Protocol,
			Description: port.Description,
		}

		// Get the port value from the settings.
		for _, sp := range a.settings.Ports {
			if sp.ContainerPort == p.Port && sp.Protocol == p.Protocol {
				p.HostPort = sp.HostPort
				break
			}
		}

		// Add it to the setup.
		setup.Ports[i] = p
		i++
	}

	return setup, nil
}

// Setup the app and save the values.
func (a *App) Setup(setup *api.Setup) error {
	// Create a backup first.
	err := a.Backup()
	if err != nil {
		return err
	}

	// Remove the previous set values first.
	a.settings.Env = make(map[string]string)
	a.settings.Ports = make(appSettingsPorts, len(setup.Ports))

	// Set the environment values to the settings.
	for _, env := range setup.Env {
		a.settings.Env[env.Name] = env.Value
	}

	// Set the port values to the settings.
	for i, p := range setup.Ports {
		a.settings.Ports[i] = &appSettingsPort{
			ContainerPort: p.Port,
			HostPort:      p.HostPort,
			Protocol:      p.Protocol,
		}
	}

	// Save the settings.
	return a.saveSettings()
}
