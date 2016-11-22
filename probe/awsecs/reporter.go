package awsecs

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
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
				rpt.ECSService = rpt.ECSService.AddNode(report.MakeNode(serviceNodeID(serviceName)))
				unique[serviceName] = true
			}
		}
		log.Debugf("Created %v ECS service nodes", len(taskServices))

		for taskArn, info := range taskMap {

			// new task node
			node := report.MakeNodeWith(taskNodeID(taskArn), map[string]string{"family": info.family})

			rpt.ECSTask = rpt.ECSTask.AddNode(node)

			for _, containerID := range info.containerIDs {
				// TODO set task node as parent of container
				log.Debugf("task %v has container %v", taskArn, containerID)
			}

			if serviceName, ok := taskServices[taskArn]; ok {
				// TODO set service node as parent of task node
				log.Debugf("service %v has task %v", serviceName, taskArn)
			}
		}

	}

	return rpt, nil
}

// Name needed for Tagger
func (r Reporter) Name() string {
	return "awsecs"
}

func serviceNodeID(id string) string {
	return fmt.Sprintf("%s;ECSService", id)
}

func taskNodeID(id string) string {
	return fmt.Sprintf("%s;ECSTask", id)
}
