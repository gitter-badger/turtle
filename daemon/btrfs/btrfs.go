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

package btrfs

import (
	"fmt"
	"strconv"

	"github.com/desertbit/turtle/daemon/config"
	"github.com/desertbit/turtle/utils"
)

// IsSubvolume checks if the directory is a btrfs subvolume.
func IsSubvolume(subvolumeDir string) bool {
	// Run the command.
	err := utils.RunCommand("btrfs", "subvolume", "show", subvolumeDir)
	if err != nil {
		return false
	}

	return true
}

// CreateSubvolume creates a btrfs subvolume.
func CreateSubvolume(subvolumeDir string) error {
	// Run the command.
	err := utils.RunCommand("btrfs", "subvolume", "create", subvolumeDir)
	if err != nil {
		return fmt.Errorf("failed to create the btrfs subvolume '%s': %v", subvolumeDir, err)
	}

	return nil
}

// DeleteSubvolume deletes a btrfs subvolume.
func DeleteSubvolume(subvolumeDir string) error {
	// Run the command.
	err := utils.RunCommand("btrfs", "subvolume", "delete", subvolumeDir)
	if err != nil {
		return fmt.Errorf("failed to delete the btrfs subvolume '%s': %v", subvolumeDir, err)
	}

	return nil
}

// SetSubvolumeReadonly sets the readonly flag on a btrfs subvolume.
func SetSubvolumeReadonly(subvolumeDir string, readonly bool) error {
	// Run the command.
	err := utils.RunCommand("btrfs", "property", "set", "-ts", subvolumeDir, "ro", strconv.FormatBool(readonly))
	if err != nil {
		return fmt.Errorf("failed to set readonly flag of btrfs subvolume '%s': %v", subvolumeDir, err)
	}

	return nil
}

// Snapshot create a btrfs snapshot of a subvolume.
func Snapshot(subvolumeDir string, snapshotDir string, readonly bool) error {
	// Check if the passed subvolume directory is a btrfs subvolume.
	if !IsSubvolume(subvolumeDir) {
		return fmt.Errorf("failed to create btrfs snapshot: the subvolume directory '%s' is not a btrfs subvolume!", subvolumeDir)
	}

	// The destination snapshot directory should not exist!
	e, err := utils.Exists(snapshotDir)
	if err != nil {
		return err
	} else if e {
		return fmt.Errorf("failed to create btrfs snapshot: the snapshot directory '%s' already exists!", snapshotDir)
	}

	// Create the snapshot directory.
	if readonly {
		err = utils.RunCommand("btrfs", "subvolume", "snapshot", "-r", subvolumeDir, snapshotDir)
	} else {
		err = utils.RunCommand("btrfs", "subvolume", "snapshot", subvolumeDir, snapshotDir)
	}
	if err != nil {
		return fmt.Errorf("failed to create btrfs snapshot '%s': %v", snapshotDir, err)
	}

	// Force changed blocks to disk, update the super block.
	err = utils.RunCommand("sync")
	if err != nil {
		return fmt.Errorf("failed to force changed blocks to disk, update the super block: %v", err)
	}

	return nil
}

// Balance a btrfs partition.
func Balance(path string) error {
	// Run the command.
	err := utils.RunCommand("btrfs", "fi", "balance", "start", "-dusage="+strconv.Itoa(config.Config.BtrfsBalanceDusage), path)
	if err != nil {
		return fmt.Errorf("failed to balance btrfs path '%s': %v", path, err)
	}

	return nil
}
