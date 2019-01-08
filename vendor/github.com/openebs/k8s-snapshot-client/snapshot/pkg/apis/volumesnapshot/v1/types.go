/*
Copyright 2017 The Kubernetes Authors.

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

package v1

import (
	core_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// VolumeSnapshotDataResourcePlural is "volumesnapshotdatas"
	VolumeSnapshotDataResourcePlural = "volumesnapshotdatas"
	// VolumeSnapshotResourcePlural is "volumesnapshots"
	VolumeSnapshotResourcePlural = "volumesnapshots"
)

// VolumeSnapshotStatus is the status of the VolumeSnapshot
type VolumeSnapshotStatus struct {
	// The time the snapshot was successfully created
	// +optional
	CreationTimestamp metav1.Time `json:"creationTimestamp" protobuf:"bytes,1,opt,name=creationTimestamp"`

	// Represent the latest available observations about the volume snapshot
	Conditions []VolumeSnapshotCondition `json:"conditions" protobuf:"bytes,2,rep,name=conditions"`
}

// VolumeSnapshotConditionType is the type of VolumeSnapshot conditions
type VolumeSnapshotConditionType string

// These are valid conditions of a volume snapshot.
const (
	// VolumeSnapshotConditionPending means the snapshot is cut and the application
	// can resume accessing data if core_v1.ConditionStatus is True. It corresponds
	// to "Uploading" in GCE PD or "Pending" in AWS and core_v1.ConditionStatus is True.
	// It also corresponds to "Creating" in OpenStack Cinder and core_v1.ConditionStatus
	// is Unknown.
	VolumeSnapshotConditionPending VolumeSnapshotConditionType = "Pending"
	// VolumeSnapshotConditionReady is added when the snapshot has been successfully created and is ready to be used.
	VolumeSnapshotConditionReady VolumeSnapshotConditionType = "Ready"
	// VolumeSnapshotConditionError means an error occurred during snapshot creation.
	VolumeSnapshotConditionError VolumeSnapshotConditionType = "Error"
)

// VolumeSnapshotCondition describes the state of a volume snapshot  at a certain point.
type VolumeSnapshotCondition struct {
	// Type of replication controller condition.
	Type VolumeSnapshotConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=VolumeSnapshotConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status core_v1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// The last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime" protobuf:"bytes,3,opt,name=lastTransitionTime"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason" protobuf:"bytes,4,opt,name=reason"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message" protobuf:"bytes,5,opt,name=message"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeSnapshot is the volume snapshot object accessible to the user. Upon succesful creation of the actual
// snapshot by the volume provider it is bound to the corresponding VolumeSnapshotData through
// the VolumeSnapshotSpec
type VolumeSnapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// Spec represents the desired state of the snapshot
	// +optional
	Spec VolumeSnapshotSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`

	// Status represents the latest observer state of the snapshot
	// +optional
	Status VolumeSnapshotStatus `json:"status" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeSnapshotList is a list of VolumeSnapshot objects
type VolumeSnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []VolumeSnapshot `json:"items"`
}

// VolumeSnapshotSpec is the desired state of the volume snapshot
type VolumeSnapshotSpec struct {
	// PersistentVolumeClaimName is the name of the PVC being snapshotted
	// +optional
	PersistentVolumeClaimName string `json:"persistentVolumeClaimName" protobuf:"bytes,1,opt,name=persistentVolumeClaimName"`

	// SnapshotDataName binds the VolumeSnapshot object with the VolumeSnapshotData
	// +optional
	SnapshotDataName string `json:"snapshotDataName" protobuf:"bytes,2,opt,name=snapshotDataName"`
}

// VolumeSnapshotDataStatus is the actual state of the volume snapshot
type VolumeSnapshotDataStatus struct {
	// The time the snapshot was successfully created
	// +optional
	CreationTimestamp metav1.Time `json:"creationTimestamp" protobuf:"bytes,1,opt,name=creationTimestamp"`

	// Representes the lates available observations about the volume snapshot
	Conditions []VolumeSnapshotDataCondition `json:"conditions" protobuf:"bytes,2,rep,name=conditions"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeSnapshotDataList is a list of VolumeSnapshotData objects
type VolumeSnapshotDataList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []VolumeSnapshotData `json:"items"`
}

// VolumeSnapshotDataConditionType is the type of the VolumeSnapshotData condition
type VolumeSnapshotDataConditionType string

// These are valid conditions of a volume snapshot.
const (
	// VolumeSnapshotDataReady is added when the on-disk snapshot has been successfully created.
	VolumeSnapshotDataConditionReady VolumeSnapshotDataConditionType = "Ready"
	// VolumeSnapshotDataPending is added when the on-disk snapshot has been successfully created but is not available to use.
	VolumeSnapshotDataConditionPending VolumeSnapshotDataConditionType = "Pending"
	// VolumeSnapshotDataError is added but the on-disk snapshot is failed to created
	VolumeSnapshotDataConditionError VolumeSnapshotDataConditionType = "Error"
)

// VolumeSnapshotDataCondition describes the state of a volume snapshot  at a certain point.
type VolumeSnapshotDataCondition struct {
	// Type of volume snapshot condition.
	Type VolumeSnapshotDataConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=VolumeSnapshotDataConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status core_v1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// The last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime" protobuf:"bytes,3,opt,name=lastTransitionTime"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason" protobuf:"bytes,4,opt,name=reason"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message" protobuf:"bytes,5,opt,name=message"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeSnapshotData represents the actual "on-disk" snapshot object
type VolumeSnapshotData struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata"`

	// Spec represents the desired state of the snapshot
	// +optional
	Spec VolumeSnapshotDataSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`

	// Status represents the latest observed state of the snapshot
	// +optional
	Status VolumeSnapshotDataStatus `json:"status" protobuf:"bytes,3,opt,name=status"`
}

// VolumeSnapshotDataSpec is the spec of the volume snapshot data
type VolumeSnapshotDataSpec struct {
	// Source represents the location and type of the volume snapshot
	VolumeSnapshotDataSource `json:",inline" protobuf:"bytes,1,opt,name=volumeSnapshotDataSource"`

	// VolumeSnapshotRef is part of bi-directional binding between VolumeSnapshot
	// and VolumeSnapshotData
	// +optional
	VolumeSnapshotRef *core_v1.ObjectReference `json:"volumeSnapshotRef" protobuf:"bytes,2,opt,name=volumeSnapshotRef"`

	// PersistentVolumeRef represents the PersistentVolume that the snapshot has been
	// taken from
	// +optional
	PersistentVolumeRef *core_v1.ObjectReference `json:"persistentVolumeRef" protobuf:"bytes,3,opt,name=persistentVolumeRef"`
}

// HostPathVolumeSnapshotSource is HostPath volume snapshot source
type HostPathVolumeSnapshotSource struct {
	// Path represents a tar file that stores the HostPath volume source
	Path string `json:"snapshot"`
}

// GlusterVolumeSnapshotSource is Gluster volume snapshot source
type GlusterVolumeSnapshotSource struct {
	// UniqueID represents a snapshot resource.
	SnapshotID string `json:"snapshotId"`
}

// AWSElasticBlockStoreVolumeSnapshotSource is AWS EBS volume snapshot source
type AWSElasticBlockStoreVolumeSnapshotSource struct {
	// Unique id of the persistent disk snapshot resource. Used to identify the disk snapshot in AWS
	SnapshotID string `json:"snapshotId"`
}

// CinderVolumeSnapshotSource is Cinder volume snapshot source
type CinderVolumeSnapshotSource struct {
	// Unique id of the cinder volume snapshot resource. Used to identify the snapshot in OpenStack
	SnapshotID string `json:"snapshotId"`
}

// GCEPersistentDiskSnapshotSource is GCE PD volume snapshot source
type GCEPersistentDiskSnapshotSource struct {
	// Unique id of the persistent disk snapshot resource. Used to identify the disk snapshot in GCE
	SnapshotName string `json:"snapshotId"`
}

// VolumeSnapshotDataSource represents the actual location and type of the snapshot. Only one of its members may be specified.
type VolumeSnapshotDataSource struct {
	// HostPath represents a directory on the host.
	// Provisioned by a developer or tester.
	// This is useful for single-node development and testing only!
	// On-host storage is not supported in any way and WILL NOT WORK in a multi-node cluster.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath
	// +optional
	HostPath *HostPathVolumeSnapshotSource `json:"hostPath,omitempty"`
	// AWSElasticBlockStore represents an AWS Disk resource that is attached to a
	// kubelet's host machine and then exposed to the pod.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore
	// +optional
	//GlusterSnapshotSource represents a gluster snapshot resource
	GlusterSnapshotVolume *GlusterVolumeSnapshotSource `json:"glusterSnapshotVolume,omitempty"`
	// +optional
	AWSElasticBlockStore *AWSElasticBlockStoreVolumeSnapshotSource `json:"awsElasticBlockStore,omitempty"`
	// GCEPersistentDiskSnapshotSource represents an GCE PD snapshot resource
	// +optional
	GCEPersistentDiskSnapshot *GCEPersistentDiskSnapshotSource `json:"gcePersistentDisk,omitempty"`
	// CinderVolumeSnapshotSource represents Cinder snapshot resource
	// +optional
	CinderSnapshot *CinderVolumeSnapshotSource `json:"cinderVolume,omitempty"`
}

// GetSupportedVolumeFromPVSpec gets supported volume from PV spec
func GetSupportedVolumeFromPVSpec(spec *core_v1.PersistentVolumeSpec) string {
	if spec.HostPath != nil {
		return "hostPath"
	}
	if spec.AWSElasticBlockStore != nil {
		return "aws_ebs"
	}
	if spec.GCEPersistentDisk != nil {
		return "gce-pd"
	}
	if spec.Cinder != nil {
		return "cinder"
	}
	if spec.Glusterfs != nil {
		return "glusterfs"
	}
	return ""
}

// GetSupportedVolumeFromSnapshotDataSpec gets supported volume from snapshot data spec
func GetSupportedVolumeFromSnapshotDataSpec(spec *VolumeSnapshotDataSpec) string {
	if spec.HostPath != nil {
		return "hostPath"
	}
	if spec.AWSElasticBlockStore != nil {
		return "aws_ebs"
	}
	if spec.GCEPersistentDiskSnapshot != nil {
		return "gce-pd"
	}
	if spec.CinderSnapshot != nil {
		return "cinder"
	}
	if spec.GlusterSnapshotVolume != nil {
		return "glusterfs"
	}
	return ""
}
