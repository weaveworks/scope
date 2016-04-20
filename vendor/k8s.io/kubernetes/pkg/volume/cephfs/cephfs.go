/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package cephfs

import (
	"fmt"
	"os"
	"strings"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/types"
	"k8s.io/kubernetes/pkg/util/mount"
	utilstrings "k8s.io/kubernetes/pkg/util/strings"
	"k8s.io/kubernetes/pkg/volume"
)

// This is the primary entrypoint for volume plugins.
func ProbeVolumePlugins() []volume.VolumePlugin {
	return []volume.VolumePlugin{&cephfsPlugin{nil}}
}

type cephfsPlugin struct {
	host volume.VolumeHost
}

var _ volume.VolumePlugin = &cephfsPlugin{}

const (
	cephfsPluginName = "kubernetes.io/cephfs"
)

func (plugin *cephfsPlugin) Init(host volume.VolumeHost) error {
	plugin.host = host
	return nil
}

func (plugin *cephfsPlugin) Name() string {
	return cephfsPluginName
}

func (plugin *cephfsPlugin) CanSupport(spec *volume.Spec) bool {
	return (spec.Volume != nil && spec.Volume.CephFS != nil) || (spec.PersistentVolume != nil && spec.PersistentVolume.Spec.CephFS != nil)
}

func (plugin *cephfsPlugin) GetAccessModes() []api.PersistentVolumeAccessMode {
	return []api.PersistentVolumeAccessMode{
		api.ReadWriteOnce,
		api.ReadOnlyMany,
		api.ReadWriteMany,
	}
}

func (plugin *cephfsPlugin) NewMounter(spec *volume.Spec, pod *api.Pod, _ volume.VolumeOptions) (volume.Mounter, error) {
	cephvs := plugin.getVolumeSource(spec)
	secret := ""
	if cephvs.SecretRef != nil {
		kubeClient := plugin.host.GetKubeClient()
		if kubeClient == nil {
			return nil, fmt.Errorf("Cannot get kube client")
		}

		secretName, err := kubeClient.Core().Secrets(pod.Namespace).Get(cephvs.SecretRef.Name)
		if err != nil {
			err = fmt.Errorf("Couldn't get secret %v/%v err: %v", pod.Namespace, cephvs.SecretRef, err)
			return nil, err
		}
		for name, data := range secretName.Data {
			secret = string(data)
			glog.V(1).Infof("found ceph secret info: %s", name)
		}
	}
	return plugin.newMounterInternal(spec, pod.UID, plugin.host.GetMounter(), secret)
}

func (plugin *cephfsPlugin) newMounterInternal(spec *volume.Spec, podUID types.UID, mounter mount.Interface, secret string) (volume.Mounter, error) {
	cephvs := plugin.getVolumeSource(spec)
	id := cephvs.User
	if id == "" {
		id = "admin"
	}
	path := cephvs.Path
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	secret_file := cephvs.SecretFile
	if secret_file == "" {
		secret_file = "/etc/ceph/" + id + ".secret"
	}

	return &cephfsMounter{
		cephfs: &cephfs{
			podUID:      podUID,
			volName:     spec.Name(),
			mon:         cephvs.Monitors,
			path:        path,
			secret:      secret,
			id:          id,
			secret_file: secret_file,
			readonly:    cephvs.ReadOnly,
			mounter:     mounter,
			plugin:      plugin},
	}, nil
}

func (plugin *cephfsPlugin) NewUnmounter(volName string, podUID types.UID) (volume.Unmounter, error) {
	return plugin.newUnmounterInternal(volName, podUID, plugin.host.GetMounter())
}

func (plugin *cephfsPlugin) newUnmounterInternal(volName string, podUID types.UID, mounter mount.Interface) (volume.Unmounter, error) {
	return &cephfsUnmounter{
		cephfs: &cephfs{
			podUID:  podUID,
			volName: volName,
			mounter: mounter,
			plugin:  plugin},
	}, nil
}

func (plugin *cephfsPlugin) getVolumeSource(spec *volume.Spec) *api.CephFSVolumeSource {
	if spec.Volume != nil && spec.Volume.CephFS != nil {
		return spec.Volume.CephFS
	} else {
		return spec.PersistentVolume.Spec.CephFS
	}
}

// CephFS volumes represent a bare host file or directory mount of an CephFS export.
type cephfs struct {
	volName     string
	podUID      types.UID
	mon         []string
	path        string
	id          string
	secret      string
	secret_file string
	readonly    bool
	mounter     mount.Interface
	plugin      *cephfsPlugin
	volume.MetricsNil
}

type cephfsMounter struct {
	*cephfs
}

var _ volume.Mounter = &cephfsMounter{}

func (cephfsVolume *cephfsMounter) GetAttributes() volume.Attributes {
	return volume.Attributes{
		ReadOnly:        cephfsVolume.readonly,
		Managed:         false,
		SupportsSELinux: false,
	}
}

// SetUp attaches the disk and bind mounts to the volume path.
func (cephfsVolume *cephfsMounter) SetUp(fsGroup *int64) error {
	return cephfsVolume.SetUpAt(cephfsVolume.GetPath(), fsGroup)
}

// SetUpAt attaches the disk and bind mounts to the volume path.
func (cephfsVolume *cephfsMounter) SetUpAt(dir string, fsGroup *int64) error {
	notMnt, err := cephfsVolume.mounter.IsLikelyNotMountPoint(dir)
	glog.V(4).Infof("CephFS mount set up: %s %v %v", dir, !notMnt, err)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if !notMnt {
		return nil
	}
	os.MkdirAll(dir, 0750)

	err = cephfsVolume.execMount(dir)
	if err == nil {
		return nil
	}

	// cleanup upon failure
	cephfsVolume.cleanup(dir)
	// return error
	return err
}

type cephfsUnmounter struct {
	*cephfs
}

var _ volume.Unmounter = &cephfsUnmounter{}

// TearDown unmounts the bind mount
func (cephfsVolume *cephfsUnmounter) TearDown() error {
	return cephfsVolume.TearDownAt(cephfsVolume.GetPath())
}

// TearDownAt unmounts the bind mount
func (cephfsVolume *cephfsUnmounter) TearDownAt(dir string) error {
	return cephfsVolume.cleanup(dir)
}

// GatePath creates global mount path
func (cephfsVolume *cephfs) GetPath() string {
	name := cephfsPluginName
	return cephfsVolume.plugin.host.GetPodVolumeDir(cephfsVolume.podUID, utilstrings.EscapeQualifiedNameForDisk(name), cephfsVolume.volName)
}

func (cephfsVolume *cephfs) cleanup(dir string) error {
	noMnt, err := cephfsVolume.mounter.IsLikelyNotMountPoint(dir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("CephFS: Error checking IsLikelyNotMountPoint: %v", err)
	}
	if noMnt {
		return os.RemoveAll(dir)
	}

	if err := cephfsVolume.mounter.Unmount(dir); err != nil {
		return fmt.Errorf("CephFS: Unmounting failed: %v", err)
	}
	noMnt, mntErr := cephfsVolume.mounter.IsLikelyNotMountPoint(dir)
	if mntErr != nil {
		return fmt.Errorf("CephFS: IsMountpoint check failed: %v", mntErr)
	}
	if noMnt {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("CephFS: removeAll %s/%v", dir, err)
		}
	}

	return nil
}

func (cephfsVolume *cephfs) execMount(mountpoint string) error {
	// cephfs mount option
	ceph_opt := ""
	// override secretfile if secret is provided
	if cephfsVolume.secret != "" {
		ceph_opt = "name=" + cephfsVolume.id + ",secret=" + cephfsVolume.secret
	} else {
		ceph_opt = "name=" + cephfsVolume.id + ",secretfile=" + cephfsVolume.secret_file
	}
	// build option array
	opt := []string{}
	if cephfsVolume.readonly {
		opt = append(opt, "ro")
	}
	opt = append(opt, ceph_opt)

	// build src like mon1:6789,mon2:6789,mon3:6789:/
	hosts := cephfsVolume.mon
	l := len(hosts)
	// pass all monitors and let ceph randomize and fail over
	i := 0
	src := ""
	for i = 0; i < l-1; i++ {
		src += hosts[i] + ","
	}
	src += hosts[i] + ":" + cephfsVolume.path

	if err := cephfsVolume.mounter.Mount(src, mountpoint, "ceph", opt); err != nil {
		return fmt.Errorf("CephFS: mount failed: %v", err)
	}

	return nil
}
