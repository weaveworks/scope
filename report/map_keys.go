package report

// node metadata keys
const (
	// probe/endpoint
	ReverseDNSNames = "reverse_dns_names"
	SnoopedDNSNames = "snooped_dns_names"
	CopyOf          = "copy_of"
	// probe/process
	PID     = "pid"
	Name    = "name" // also used by probe/docker
	PPID    = "ppid"
	Cmdline = "cmdline"
	Threads = "threads"
	// probe/docker
	DockerContainerID            = "docker_container_id"
	DockerImageID                = "docker_image_id"
	DockerImageName              = "docker_image_name"
	DockerImageTag               = "docker_image_tag"
	DockerImageSize              = "docker_image_size"
	DockerImageVirtualSize       = "docker_image_virtual_size"
	DockerIsInHostNetwork        = "docker_is_in_host_network"
	DockerServiceName            = "service_name"
	DockerStackNamespace         = "stack_namespace"
	DockerStopContainer          = "docker_stop_container"
	DockerStartContainer         = "docker_start_container"
	DockerRestartContainer       = "docker_restart_container"
	DockerPauseContainer         = "docker_pause_container"
	DockerUnpauseContainer       = "docker_unpause_container"
	DockerRemoveContainer        = "docker_remove_container"
	DockerAttachContainer        = "docker_attach_container"
	DockerExecContainer          = "docker_exec_container"
	DockerContainerName          = "docker_container_name"
	DockerContainerCommand       = "docker_container_command"
	DockerContainerPorts         = "docker_container_ports"
	DockerContainerCreated       = "docker_container_created"
	DockerContainerNetworks      = "docker_container_networks"
	DockerContainerIPs           = "docker_container_ips"
	DockerContainerHostname      = "docker_container_hostname"
	DockerContainerIPsWithScopes = "docker_container_ips_with_scopes"
	DockerContainerState         = "docker_container_state"
	DockerContainerStateHuman    = "docker_container_state_human"
	DockerContainerUptime        = "docker_container_uptime"
	DockerContainerRestartCount  = "docker_container_restart_count"
	DockerContainerNetworkMode   = "docker_container_network_mode"
	DockerEnvPrefix              = "docker_env_"
	// probe/kubernetes
	KubernetesName                 = "kubernetes_name"
	KubernetesNamespace            = "kubernetes_namespace"
	KubernetesCreated              = "kubernetes_created"
	KubernetesIP                   = "kubernetes_ip"
	KubernetesObservedGeneration   = "kubernetes_observed_generation"
	KubernetesReplicas             = "kubernetes_replicas"
	KubernetesDesiredReplicas      = "kubernetes_desired_replicas"
	KubernetesNodeType             = "kubernetes_node_type"
	KubernetesGetLogs              = "kubernetes_get_logs"
	KubernetesDeletePod            = "kubernetes_delete_pod"
	KubernetesScaleUp              = "kubernetes_scale_up"
	KubernetesScaleDown            = "kubernetes_scale_down"
	KubernetesUpdatedReplicas      = "kubernetes_updated_replicas"
	KubernetesAvailableReplicas    = "kubernetes_available_replicas"
	KubernetesUnavailableReplicas  = "kubernetes_unavailable_replicas"
	KubernetesStrategy             = "kubernetes_strategy"
	KubernetesFullyLabeledReplicas = "kubernetes_fully_labeled_replicas"
	KubernetesState                = "kubernetes_state"
	KubernetesIsInHostNetwork      = "kubernetes_is_in_host_network"
	KubernetesRestartCount         = "kubernetes_restart_count"
	KubernetesMisscheduledReplicas = "kubernetes_misscheduled_replicas"
	KubernetesPublicIP             = "kubernetes_public_ip"
	KubernetesSchedule             = "kubernetes_schedule"
	KubernetesSuspended            = "kubernetes_suspended"
	KubernetesLastScheduled        = "kubernetes_last_scheduled"
	KubernetesActiveJobs           = "kubernetes_active_jobs"
	KubernetesType                 = "kubernetes_type"
	KubernetesPorts                = "kubernetes_ports"
	KubernetesVolumeClaim          = "kubernetes_volume_claim"
	KubernetesStorageClassName     = "kubernetes_storage_class_name"
	KubernetesAccessModes          = "kubernetes_access_modes"
	KubernetesReclaimPolicy        = "kubernetes_reclaim_policy"
	KubernetesStatus               = "kubernetes_status"
	KubernetesMessage              = "kubernetes_message"
	KubernetesVolumeName           = "kubernetes_volume_name"
	KubernetesProvisioner          = "kubernetes_provisioner"
	KubernetesStorageDriver        = "kubernetes_storage_driver"
	KubernetesVolumeSnapshotName   = "kubernetes_volume_snapshot_name"
	KubernetesSnapshotData         = "kuberneets_snapshot_data"
	KubernetesCreateVolumeSnapshot = "kubernetes_create_volume_snapshot"
	KubernetesVolumeCapacity       = "kubernetes_volume_capacity"
	KubernetesCloneVolumeSnapshot  = "kubernetes_clone_volume_snapshot"
	KubernetesDeleteVolumeSnapshot = "kubernetes_delete_volume_snapshot"
	// probe/awsecs
	ECSCluster             = "ecs_cluster"
	ECSCreatedAt           = "ecs_created_at"
	ECSTaskFamily          = "ecs_task_family"
	ECSServiceDesiredCount = "ecs_service_desired_count"
	ECSServiceRunningCount = "ecs_service_running_count"
	ECSScaleUp             = "ecs_scale_up"
	ECSScaleDown           = "ecs_scale_down"
)

/* Lookup table to allow msgpack/json decoder to avoid heap allocation
   for common ps.Map keys. The map is static so we don't have to lock
   access from multiple threads and don't have to worry about it
   getting clogged with values that are only used once.
*/
var commonKeys = map[string]string{
	Endpoint:              Endpoint,
	Process:               Process,
	Container:             Container,
	Pod:                   Pod,
	Service:               Service,
	Deployment:            Deployment,
	ReplicaSet:            ReplicaSet,
	DaemonSet:             DaemonSet,
	StatefulSet:           StatefulSet,
	CronJob:               CronJob,
	ContainerImage:        ContainerImage,
	Host:                  Host,
	Overlay:               Overlay,
	ECSService:            ECSService,
	ECSTask:               ECSTask,
	SwarmService:          SwarmService,
	PersistentVolume:      PersistentVolume,
	PersistentVolumeClaim: PersistentVolumeClaim,
	StorageClass:          StorageClass,
	VolumeSnapshot:        VolumeSnapshot,
	VolumeSnapshotData:    VolumeSnapshotData,

	HostNodeID:             HostNodeID,
	ControlProbeID:         ControlProbeID,
	DoesNotMakeConnections: DoesNotMakeConnections,

	ReverseDNSNames: ReverseDNSNames,
	SnoopedDNSNames: SnoopedDNSNames,
	CopyOf:          CopyOf,

	PID:     PID,
	Name:    Name,
	PPID:    PPID,
	Cmdline: Cmdline,
	Threads: Threads,

	DockerContainerID:            DockerContainerID,
	DockerImageID:                DockerImageID,
	DockerImageName:              DockerImageName,
	DockerImageTag:               DockerImageTag,
	DockerImageSize:              DockerImageSize,
	DockerImageVirtualSize:       DockerImageVirtualSize,
	DockerIsInHostNetwork:        DockerIsInHostNetwork,
	DockerServiceName:            DockerServiceName,
	DockerStackNamespace:         DockerStackNamespace,
	DockerStopContainer:          DockerStopContainer,
	DockerStartContainer:         DockerStartContainer,
	DockerRestartContainer:       DockerRestartContainer,
	DockerPauseContainer:         DockerPauseContainer,
	DockerUnpauseContainer:       DockerUnpauseContainer,
	DockerRemoveContainer:        DockerRemoveContainer,
	DockerAttachContainer:        DockerAttachContainer,
	DockerExecContainer:          DockerExecContainer,
	DockerContainerName:          DockerContainerName,
	DockerContainerCommand:       DockerContainerCommand,
	DockerContainerPorts:         DockerContainerPorts,
	DockerContainerCreated:       DockerContainerCreated,
	DockerContainerNetworks:      DockerContainerNetworks,
	DockerContainerIPs:           DockerContainerIPs,
	DockerContainerHostname:      DockerContainerHostname,
	DockerContainerIPsWithScopes: DockerContainerIPsWithScopes,
	DockerContainerState:         DockerContainerState,
	DockerContainerStateHuman:    DockerContainerStateHuman,
	DockerContainerUptime:        DockerContainerUptime,
	DockerContainerRestartCount:  DockerContainerRestartCount,
	DockerContainerNetworkMode:   DockerContainerNetworkMode,

	KubernetesName:                 KubernetesName,
	KubernetesNamespace:            KubernetesNamespace,
	KubernetesCreated:              KubernetesCreated,
	KubernetesIP:                   KubernetesIP,
	KubernetesObservedGeneration:   KubernetesObservedGeneration,
	KubernetesReplicas:             KubernetesReplicas,
	KubernetesDesiredReplicas:      KubernetesDesiredReplicas,
	KubernetesNodeType:             KubernetesNodeType,
	KubernetesGetLogs:              KubernetesGetLogs,
	KubernetesDeletePod:            KubernetesDeletePod,
	KubernetesScaleUp:              KubernetesScaleUp,
	KubernetesScaleDown:            KubernetesScaleDown,
	KubernetesUpdatedReplicas:      KubernetesUpdatedReplicas,
	KubernetesAvailableReplicas:    KubernetesAvailableReplicas,
	KubernetesUnavailableReplicas:  KubernetesUnavailableReplicas,
	KubernetesStrategy:             KubernetesStrategy,
	KubernetesFullyLabeledReplicas: KubernetesFullyLabeledReplicas,
	KubernetesState:                KubernetesState,
	KubernetesIsInHostNetwork:      KubernetesIsInHostNetwork,
	KubernetesRestartCount:         KubernetesRestartCount,
	KubernetesMisscheduledReplicas: KubernetesMisscheduledReplicas,
	KubernetesPublicIP:             KubernetesPublicIP,
	KubernetesSchedule:             KubernetesSchedule,
	KubernetesSuspended:            KubernetesSuspended,
	KubernetesLastScheduled:        KubernetesLastScheduled,
	KubernetesActiveJobs:           KubernetesActiveJobs,
	KubernetesType:                 KubernetesType,
	KubernetesPorts:                KubernetesPorts,

	ECSCluster:             ECSCluster,
	ECSCreatedAt:           ECSCreatedAt,
	ECSTaskFamily:          ECSTaskFamily,
	ECSServiceDesiredCount: ECSServiceDesiredCount,
	ECSServiceRunningCount: ECSServiceRunningCount,
	ECSScaleUp:             ECSScaleUp,
	ECSScaleDown:           ECSScaleDown,
}

func lookupCommonKey(b []byte) string {
	if key, ok := commonKeys[string(b)]; ok {
		return key
	}
	return string(b)
}

func isCommandKey(key string) bool {
	return key == Cmdline || key == DockerContainerCommand
}
