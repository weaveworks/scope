/*
Copyright 2014 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gce_pd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/glog"
	gcecloud "k8s.io/kubernetes/pkg/cloudprovider/providers/gce"
	"k8s.io/kubernetes/pkg/util/exec"
	"k8s.io/kubernetes/pkg/util/keymutex"
	"k8s.io/kubernetes/pkg/util/runtime"
	"k8s.io/kubernetes/pkg/util/sets"
	"k8s.io/kubernetes/pkg/volume"
)

const (
	diskByIdPath         = "/dev/disk/by-id/"
	diskGooglePrefix     = "google-"
	diskScsiGooglePrefix = "scsi-0Google_PersistentDisk_"
	diskPartitionSuffix  = "-part"
	diskSDPath           = "/dev/sd"
	diskSDPattern        = "/dev/sd*"
	maxChecks            = 60
	maxRetries           = 10
	checkSleepDuration   = time.Second
	errorSleepDuration   = 5 * time.Second
)

// Singleton key mutex for keeping attach/detach operations for the same PD atomic
var attachDetachMutex = keymutex.NewKeyMutex()

type GCEDiskUtil struct{}

// Attaches a disk specified by a volume.GCEPersistentDisk to the current kubelet.
// Mounts the disk to it's global path.
func (diskUtil *GCEDiskUtil) AttachAndMountDisk(b *gcePersistentDiskMounter, globalPDPath string) error {
	glog.V(5).Infof("AttachAndMountDisk(...) called for PD %q. Will block for existing operations, if any. (globalPDPath=%q)\r\n", b.pdName, globalPDPath)

	// Block execution until any pending detach operations for this PD have completed
	attachDetachMutex.LockKey(b.pdName)
	defer attachDetachMutex.UnlockKey(b.pdName)

	glog.V(5).Infof("AttachAndMountDisk(...) called for PD %q. Awake and ready to execute. (globalPDPath=%q)\r\n", b.pdName, globalPDPath)

	sdBefore, err := filepath.Glob(diskSDPattern)
	if err != nil {
		glog.Errorf("Error filepath.Glob(\"%s\"): %v\r\n", diskSDPattern, err)
	}
	sdBeforeSet := sets.NewString(sdBefore...)

	devicePath, err := attachDiskAndVerify(b, sdBeforeSet)
	if err != nil {
		return err
	}

	// Only mount the PD globally once.
	notMnt, err := b.mounter.IsLikelyNotMountPoint(globalPDPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(globalPDPath, 0750); err != nil {
				return err
			}
			notMnt = true
		} else {
			return err
		}
	}
	options := []string{}
	if b.readOnly {
		options = append(options, "ro")
	}
	if notMnt {
		err = b.diskMounter.FormatAndMount(devicePath, globalPDPath, b.fsType, options)
		if err != nil {
			os.Remove(globalPDPath)
			return err
		}
	}
	return nil
}

// Unmounts the device and detaches the disk from the kubelet's host machine.
func (util *GCEDiskUtil) DetachDisk(c *gcePersistentDiskUnmounter) error {
	glog.V(5).Infof("DetachDisk(...) for PD %q\r\n", c.pdName)

	if err := unmountPDAndRemoveGlobalPath(c); err != nil {
		glog.Errorf("Error unmounting PD %q: %v", c.pdName, err)
	}

	// Detach disk asynchronously so that the kubelet sync loop is not blocked.
	go detachDiskAndVerify(c)
	return nil
}

func (util *GCEDiskUtil) DeleteVolume(d *gcePersistentDiskDeleter) error {
	cloud, err := getCloudProvider(d.gcePersistentDisk.plugin)
	if err != nil {
		return err
	}

	if err = cloud.DeleteDisk(d.pdName); err != nil {
		glog.V(2).Infof("Error deleting GCE PD volume %s: %v", d.pdName, err)
		return err
	}
	glog.V(2).Infof("Successfully deleted GCE PD volume %s", d.pdName)
	return nil
}

// CreateVolume creates a GCE PD.
// Returns: volumeID, volumeSizeGB, labels, error
func (gceutil *GCEDiskUtil) CreateVolume(c *gcePersistentDiskProvisioner) (string, int, map[string]string, error) {
	cloud, err := getCloudProvider(c.gcePersistentDisk.plugin)
	if err != nil {
		return "", 0, nil, err
	}

	name := volume.GenerateVolumeName(c.options.ClusterName, c.options.PVName, 63) // GCE PD name can have up to 63 characters
	requestBytes := c.options.Capacity.Value()
	// GCE works with gigabytes, convert to GiB with rounding up
	requestGB := volume.RoundUpSize(requestBytes, 1024*1024*1024)

	// The disk will be created in the zone in which this code is currently running
	// TODO: We should support auto-provisioning volumes in multiple/specified zones
	zone, err := cloud.GetZone()
	if err != nil {
		glog.V(2).Infof("error getting zone information from GCE: %v", err)
		return "", 0, nil, err
	}

	err = cloud.CreateDisk(name, zone.FailureDomain, int64(requestGB), *c.options.CloudTags)
	if err != nil {
		glog.V(2).Infof("Error creating GCE PD volume: %v", err)
		return "", 0, nil, err
	}
	glog.V(2).Infof("Successfully created GCE PD volume %s", name)

	labels, err := cloud.GetAutoLabelsForPD(name)
	if err != nil {
		// We don't really want to leak the volume here...
		glog.Errorf("error getting labels for volume %q: %v", name, err)
	}

	return name, int(requestGB), labels, nil
}

// Attaches the specified persistent disk device to node, verifies that it is attached, and retries if it fails.
func attachDiskAndVerify(b *gcePersistentDiskMounter, sdBeforeSet sets.String) (string, error) {
	devicePaths := getDiskByIdPaths(b.gcePersistentDisk)
	var gceCloud *gcecloud.GCECloud
	for numRetries := 0; numRetries < maxRetries; numRetries++ {
		var err error
		if gceCloud == nil {
			gceCloud, err = getCloudProvider(b.gcePersistentDisk.plugin)
			if err != nil || gceCloud == nil {
				// Retry on error. See issue #11321
				glog.Errorf("Error getting GCECloudProvider while detaching PD %q: %v", b.pdName, err)
				time.Sleep(errorSleepDuration)
				continue
			}
		}

		if numRetries > 0 {
			glog.Warningf("Retrying attach for GCE PD %q (retry count=%v).", b.pdName, numRetries)
		}

		if err := gceCloud.AttachDisk(b.pdName, b.plugin.host.GetHostName(), b.readOnly); err != nil {
			glog.Errorf("Error attaching PD %q: %v", b.pdName, err)
			time.Sleep(errorSleepDuration)
			continue
		}

		for numChecks := 0; numChecks < maxChecks; numChecks++ {
			path, err := verifyDevicePath(devicePaths, sdBeforeSet)
			if err != nil {
				// Log error, if any, and continue checking periodically. See issue #11321
				glog.Errorf("Error verifying GCE PD (%q) is attached: %v", b.pdName, err)
			} else if path != "" {
				// A device path has successfully been created for the PD
				glog.Infof("Successfully attached GCE PD %q.", b.pdName)
				return path, nil
			}

			// Sleep then check again
			glog.V(3).Infof("Waiting for GCE PD %q to attach.", b.pdName)
			time.Sleep(checkSleepDuration)
		}
	}

	return "", fmt.Errorf("Could not attach GCE PD %q. Timeout waiting for mount paths to be created.", b.pdName)
}

// Returns the first path that exists, or empty string if none exist.
func verifyDevicePath(devicePaths []string, sdBeforeSet sets.String) (string, error) {
	if err := udevadmChangeToNewDrives(sdBeforeSet); err != nil {
		// udevadm errors should not block disk detachment, log and continue
		glog.Errorf("udevadmChangeToNewDrives failed with: %v", err)
	}

	for _, path := range devicePaths {
		if pathExists, err := pathExists(path); err != nil {
			return "", fmt.Errorf("Error checking if path exists: %v", err)
		} else if pathExists {
			return path, nil
		}
	}

	return "", nil
}

// Detaches the specified persistent disk device from node, verifies that it is detached, and retries if it fails.
// This function is intended to be called asynchronously as a go routine.
func detachDiskAndVerify(c *gcePersistentDiskUnmounter) {
	glog.V(5).Infof("detachDiskAndVerify(...) for pd %q. Will block for pending operations", c.pdName)
	defer runtime.HandleCrash()

	// Block execution until any pending attach/detach operations for this PD have completed
	attachDetachMutex.LockKey(c.pdName)
	defer attachDetachMutex.UnlockKey(c.pdName)

	glog.V(5).Infof("detachDiskAndVerify(...) for pd %q. Awake and ready to execute.", c.pdName)

	devicePaths := getDiskByIdPaths(c.gcePersistentDisk)
	var gceCloud *gcecloud.GCECloud
	for numRetries := 0; numRetries < maxRetries; numRetries++ {
		var err error
		if gceCloud == nil {
			gceCloud, err = getCloudProvider(c.gcePersistentDisk.plugin)
			if err != nil || gceCloud == nil {
				// Retry on error. See issue #11321
				glog.Errorf("Error getting GCECloudProvider while detaching PD %q: %v", c.pdName, err)
				time.Sleep(errorSleepDuration)
				continue
			}
		}

		if numRetries > 0 {
			glog.Warningf("Retrying detach for GCE PD %q (retry count=%v).", c.pdName, numRetries)
		}

		if err := gceCloud.DetachDisk(c.pdName, c.plugin.host.GetHostName()); err != nil {
			glog.Errorf("Error detaching PD %q: %v", c.pdName, err)
			time.Sleep(errorSleepDuration)
			continue
		}

		for numChecks := 0; numChecks < maxChecks; numChecks++ {
			allPathsRemoved, err := verifyAllPathsRemoved(devicePaths)
			if err != nil {
				// Log error, if any, and continue checking periodically.
				glog.Errorf("Error verifying GCE PD (%q) is detached: %v", c.pdName, err)
			} else if allPathsRemoved {
				// All paths to the PD have been successfully removed
				unmountPDAndRemoveGlobalPath(c)
				glog.Infof("Successfully detached GCE PD %q.", c.pdName)
				return
			}

			// Sleep then check again
			glog.V(3).Infof("Waiting for GCE PD %q to detach.", c.pdName)
			time.Sleep(checkSleepDuration)
		}

	}

	glog.Errorf("Failed to detach GCE PD %q. One or more mount paths was not removed.", c.pdName)
}

// Unmount the global PD mount, which should be the only one, and delete it.
func unmountPDAndRemoveGlobalPath(c *gcePersistentDiskUnmounter) error {
	globalPDPath := makeGlobalPDName(c.plugin.host, c.pdName)

	err := c.mounter.Unmount(globalPDPath)
	os.Remove(globalPDPath)
	return err
}

// Returns the first path that exists, or empty string if none exist.
func verifyAllPathsRemoved(devicePaths []string) (bool, error) {
	allPathsRemoved := true
	for _, path := range devicePaths {
		if err := udevadmChangeToDrive(path); err != nil {
			// udevadm errors should not block disk detachment, log and continue
			glog.Errorf("%v", err)
		}
		if exists, err := pathExists(path); err != nil {
			return false, fmt.Errorf("Error checking if path exists: %v", err)
		} else {
			allPathsRemoved = allPathsRemoved && !exists
		}
	}

	return allPathsRemoved, nil
}

// Returns list of all /dev/disk/by-id/* paths for given PD.
func getDiskByIdPaths(pd *gcePersistentDisk) []string {
	devicePaths := []string{
		path.Join(diskByIdPath, diskGooglePrefix+pd.pdName),
		path.Join(diskByIdPath, diskScsiGooglePrefix+pd.pdName),
	}

	if pd.partition != "" {
		for i, path := range devicePaths {
			devicePaths[i] = path + diskPartitionSuffix + pd.partition
		}
	}

	return devicePaths
}

// Checks if the specified path exists
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

// Return cloud provider
func getCloudProvider(plugin *gcePersistentDiskPlugin) (*gcecloud.GCECloud, error) {
	if plugin == nil {
		return nil, fmt.Errorf("Failed to get GCE Cloud Provider. plugin object is nil.")
	}
	if plugin.host == nil {
		return nil, fmt.Errorf("Failed to get GCE Cloud Provider. plugin.host object is nil.")
	}

	cloudProvider := plugin.host.GetCloudProvider()
	gceCloudProvider, ok := cloudProvider.(*gcecloud.GCECloud)
	if !ok || gceCloudProvider == nil {
		return nil, fmt.Errorf("Failed to get GCE Cloud Provider. plugin.host.GetCloudProvider returned %v instead", cloudProvider)
	}

	return gceCloudProvider, nil
}

// Calls "udevadm trigger --action=change" for newly created "/dev/sd*" drives (exist only in after set).
// This is workaround for Issue #7972. Once the underlying issue has been resolved, this may be removed.
func udevadmChangeToNewDrives(sdBeforeSet sets.String) error {
	sdAfter, err := filepath.Glob(diskSDPattern)
	if err != nil {
		return fmt.Errorf("Error filepath.Glob(\"%s\"): %v\r\n", diskSDPattern, err)
	}

	for _, sd := range sdAfter {
		if !sdBeforeSet.Has(sd) {
			return udevadmChangeToDrive(sd)
		}
	}

	return nil
}

// Calls "udevadm trigger --action=change" on the specified drive.
// drivePath must be the the block device path to trigger on, in the format "/dev/sd*", or a symlink to it.
// This is workaround for Issue #7972. Once the underlying issue has been resolved, this may be removed.
func udevadmChangeToDrive(drivePath string) error {
	glog.V(5).Infof("udevadmChangeToDrive: drive=%q", drivePath)

	// Evaluate symlink, if any
	drive, err := filepath.EvalSymlinks(drivePath)
	if err != nil {
		return fmt.Errorf("udevadmChangeToDrive: filepath.EvalSymlinks(%q) failed with %v.", drivePath, err)
	}
	glog.V(5).Infof("udevadmChangeToDrive: symlink path is %q", drive)

	// Check to make sure input is "/dev/sd*"
	if !strings.Contains(drive, diskSDPath) {
		return fmt.Errorf("udevadmChangeToDrive: expected input in the form \"%s\" but drive is %q.", diskSDPattern, drive)
	}

	// Call "udevadm trigger --action=change --property-match=DEVNAME=/dev/sd..."
	_, err = exec.New().Command(
		"udevadm",
		"trigger",
		"--action=change",
		fmt.Sprintf("--property-match=DEVNAME=%s", drive)).CombinedOutput()
	if err != nil {
		return fmt.Errorf("udevadmChangeToDrive: udevadm trigger failed for drive %q with %v.", drive, err)
	}
	return nil
}
