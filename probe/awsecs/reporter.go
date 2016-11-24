package awsecs

import (
	log "github.com/Sirupsen/logrus"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
)

const (
	TaskFamily = "ecs_task_family"
)

type taskInfo struct {
	containerIDs []string
	family       string
}

// return map from cluster to map of task arns to task infos
func getLabelInfo(rpt report.Report) map[string]map[string]*taskInfo {
	results := make(map[string]map[string]*taskInfo)
	log.Debug("scanning for ECS containers")
	for nodeID, node := range rpt.Container.Nodes {

		taskArn, taskArnOk := node.Latest.Lookup(docker.LabelPrefix + "com.amazonaws.ecs.task-arn")
		cluster, clusterOk := node.Latest.Lookup(docker.LabelPrefix + "com.amazonaws.ecs.cluster")
		family, familyOk := node.Latest.Lookup(docker.LabelPrefix + "com.amazonaws.ecs.task-definition-family")

		if taskArnOk && clusterOk && familyOk {
			taskMap, ok := results[cluster]
			if !ok {
				taskMap = make(map[string]*taskInfo)
				results[cluster] = taskMap
			}

			task, ok := taskMap[taskArn]
			if !ok {
				task = &taskInfo{containerIDs: make([]string, 0), family: family}
				taskMap[taskArn] = task
			}

			task.containerIDs = append(task.containerIDs, nodeID)
		}
	}
	log.Debug("Got ECS container info: %v", results)
	return results
}

// implements Tagger
type Reporter struct {
}

// Tag needed for Tagger
func (r Reporter) Tag(rpt report.Report) (report.Report, error) {
	rpt = rpt.Copy()

	clusterMap := getLabelInfo(rpt)

	for cluster, taskMap := range clusterMap {
		log.Debugf("Fetching ECS info for cluster %v with %v tasks", cluster, len(taskMap))

		client, err := newClient(cluster)
		if err != nil {
			return rpt, err
		}

		taskArns := make([]string, 0, len(taskMap))
		for taskArn := range taskMap {
			taskArns = append(taskArns, taskArn)
		}

		taskServices, err := client.getTaskServices(taskArns)
		if err != nil {
			return rpt, err
		}

		// Create all the services first
		unique := make(map[string]bool)
		for _, serviceName := range taskServices {
			if !unique[serviceName] {
				serviceID := report.MakeECSServiceNodeID(serviceName)
				rpt.ECSService = rpt.ECSService.AddNode(report.MakeNode(serviceID))
				unique[serviceName] = true
			}
		}
		log.Debugf("Created %v ECS service nodes", len(taskServices))

		for taskArn, info := range taskMap {

			// new task node
			taskID := report.MakeECSTaskNodeID(taskArn)
			node := report.MakeNodeWith(taskID, map[string]string{TaskFamily: info.family})
			rpt.ECSTask = rpt.ECSTask.AddNode(node)

			// parents sets to merge into all matching container nodes
			parentsSets := report.MakeSets()
			parentsSets = parentsSets.Add(report.ECSTask, report.MakeStringSet(taskID))
			if serviceName, ok := taskServices[taskArn]; ok {
				serviceID := report.MakeECSServiceNodeID(serviceName)
				parentsSets = parentsSets.Add(report.ECSService, report.MakeStringSet(serviceID))
			}

			for _, containerID := range info.containerIDs {
				if containerNode, ok := rpt.Container.Nodes[containerID]; ok {
					rpt.Container.Nodes[containerID] = containerNode.WithParents(parentsSets)
				} else {
					log.Warnf("Got task info for non-existent container %v, this shouldn't be able to happen", containerID)
				}
			}

		}

	}

	return rpt, nil
}

// Name needed for Tagger
func (r Reporter) Name() string {
	return "awsecs"
}
