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
	"path/filepath"
	"strings"
)

const (
	DefaultImageTag    = "latest"
	DefaultNetworkMode = "bridge"

	maxContainerWaitAfterStartup = 20000 // seconds
	readonlyVolumeSuffix         = ":ro"
)

//#######################//
//### Containers type ###//
//#######################//

type Containers []*Container

// IsValid checks if required values are missing or invalid.
func (cc Containers) IsValid() error {
	for _, c := range cc {
		if len(c.Name) == 0 {
			return fmt.Errorf("Container name is empty!")
		} else if len(c.Image) == 0 {
			return fmt.Errorf("Container '%s' image is empty!", c.Name)
		} else if strings.Contains(c.Image, ":") || strings.Contains(c.Image, "..") {
			return fmt.Errorf("Container '%s' image '%s' contains an invalid character!", c.Name, c.Image)
		} else if c.WaitAfterStartup < 0 || c.WaitAfterStartup > maxContainerWaitAfterStartup {
			return fmt.Errorf("Container '%s' WaitAfterStartup '%v' value has an invalid range!", c.Name, c.WaitAfterStartup)
		}

		for _, v := range c.Volumes {
			if strings.Contains(strings.TrimSuffix(v, readonlyVolumeSuffix), ":") {
				return fmt.Errorf("Container '%s': volume '%s' contains invalid character ':'!", c.Name, v)
			} else if strings.Contains(v, "..") {
				return fmt.Errorf("Container '%s': volume '%s' contains invalid character '..'!", c.Name, v)
			}
		}
	}

	return nil
}

func (cc Containers) Prepare() error {
	for _, c := range cc {
		// Set the default image tag if not set.
		if len(c.Tag) == 0 {
			c.Tag = DefaultImageTag
		}
		if len(c.NetworkMode) == 0 {
			c.NetworkMode = DefaultNetworkMode
		}
	}

	return nil
}

// Sort the containers slice by their startup level.
// The first container in the sorted slice should be started first and stopped last.
func (cc Containers) Sort() error {
	// Create a dependency graph.
	// The dependencies are the links.
	graph := make(map[string][]string)

	// Fill the graph.
	for _, c := range cc {
		// Check if the linked containers exists.
		for _, l := range c.Links {
			found := false
			for _, lc := range cc {
				if l == lc.Name {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("container '%s': linked container '%s' does not exists!", c.Name, l)
			}
		}

		// Add the container to the graph.
		graph[c.Name] = c.Links
	}

	// Do the topological sort.
	order, cyclic := topSortDFS(graph)
	if len(cyclic) > 0 {
		// Append the first cycle element to visualize the cycle.
		cyclic = append(cyclic, cyclic[0])
		return fmt.Errorf("container link cycle detected: %+v", cyclic)
	}

	// Sort the containers.
	i := 0
	sorted := make(Containers, len(cc))
	for _, o := range order {
		for _, c := range cc {
			if o == c.Name {
				sorted[i] = c
				i++
				break
			}
		}
	}

	// Set the sorted container slice to the current container.
	// Reverse the order, because the first container should be the container without dependencies.
	for x, i := 0, len(sorted)-1; i >= 0; x, i = x+1, i-1 {
		cc[x] = sorted[i]
	}

	return nil
}

//######################//
//### Container type ###//
//######################//

type Container struct {
	Name  string // Container name.
	Image string // Docker image name. If the image name starts with a dot, then it will be build from the source directory. Subdirectories can be specified after the dot.

	// Optional
	Tag              string   // The image tag.
	WaitAfterStartup int      // Wait x milliseconds after the container started. This delays the next container startup.
	Links            []string // List of linked container names.
	Volumes          []string // List of volume mount points. A valid suffix for read-only mount is ':ro'.
	StaticVolumes    []string // List of static predefined volume mount points.
	Env              []string // A list of static predefined environment variables in the form of VAR=value.
	Cmd              []string // Specifiy a custom container command.
	Entrypoint       []string // Specifiy a custom container entrypoint.
	WorkingDir       string   // A string value containing the working dir for commands to run in.
	DNS              []string // A list of dns servers for the container to use.
	Hostname         string   // A string value containing the desired hostname to use for the container.
	Domainname       string   // A string value containing the desired domain name to use for the container.
	NetworkDisabled  bool     // Boolean value, when true disables neworking for the container
	NetworkMode      string   `toml:"Net"` // Set the Network mode for the container. Default: bridge
}

// IsLocalBuild returns a boolean whenever this container image should be build from the local source.
func (c *Container) IsLocalBuild() bool {
	return strings.HasPrefix(c.Image, ".")
}

// BuildPath returns the local build path for the container image.
// Pass the app's source path.
func (c *Container) BuildPath(sourcePath string) string {
	return filepath.Clean(sourcePath + "/" + strings.TrimPrefix(c.Image, "."))
}

// GetVolumeBinds returns a valid docker list of all container volume binds.
func (c *Container) GetVolumeBinds(volumesPath string) []string {
	binds := make([]string, len(c.Volumes)+len(c.StaticVolumes))

	// Create the bind volumes slice.
	i := 0
	for _, v := range c.Volumes {
		binds[i] = filepath.Clean(volumesPath+"/"+c.Name+"/"+strings.TrimSuffix(v, readonlyVolumeSuffix)) + ":" + v
		i++
	}

	// Add the static volume mount.
	for _, v := range c.StaticVolumes {
		binds[i] = v
		i++
	}

	return binds
}

//###############//
//### Private ###//
//###############//

// Topological sort.
// Source -> http://rosettacode.org/wiki/Topological_sort#Go
func topSortDFS(g map[string][]string) (order, cyclic []string) {
	L := make([]string, len(g))
	i := len(L)
	temp := map[string]bool{}
	perm := map[string]bool{}
	var cycleFound bool
	var cycleStart string
	var visit func(string)
	visit = func(n string) {
		switch {
		case temp[n]:
			cycleFound = true
			cycleStart = n
			return
		case perm[n]:
			return
		}
		temp[n] = true
		for _, m := range g[n] {
			visit(m)
			if cycleFound {
				if cycleStart > "" {
					cyclic = append(cyclic, n)
					if n == cycleStart {
						cycleStart = ""
					}
				}
				return
			}
		}
		delete(temp, n)
		perm[n] = true
		i--
		L[i] = n
	}
	for n := range g {
		if perm[n] {
			continue
		}
		visit(n)
		if cycleFound {
			return nil, cyclic
		}
	}
	return L, nil
}
