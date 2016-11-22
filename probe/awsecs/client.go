package awsecs

import (
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

// a wrapper around an AWS client that makes all the needed calls and just exposes the final results
type ecsClient struct {
	client  *ecs.ECS
	cluster string
}

func newClient(cluster string) (*ecsClient, error) {
	sess := session.New()

	region, err := ec2metadata.New(sess).Region()
	if err != nil {
		return nil, err
	}

	return &ecsClient{
		client:  ecs.New(sess, &aws.Config{Region: aws.String(region)}),
		cluster: cluster,
	}, nil
}

// returns a map from deployment ids to service names
// cannot fail as it will attempt to deliver partial results, though that may end up being no results
func (c ecsClient) getDeploymentMap() map[string]string {
	results := make(map[string]string)
	lock := sync.Mutex{} // lock mediates access to results

	group := sync.WaitGroup{}

	err := c.client.ListServicesPages(
		&ecs.ListServicesInput{Cluster: &c.cluster},
		func(page *ecs.ListServicesOutput, lastPage bool) bool {
			// describe each page of 10 (the max for one describe command) concurrently
			group.Add(1)
			go func() {
				defer group.Done()

				resp, err := c.client.DescribeServices(&ecs.DescribeServicesInput{
					Cluster:  &c.cluster,
					Services: page.ServiceArns,
				})
				if err != nil {
					// rather than trying to propogate errors up, just log a warning here
					log.Warnf("Error describing some ECS services, ECS service report may be incomplete: %v", err)
					return
				}

				for _, failure := range resp.Failures {
					// log the failures but still continue with what succeeded
					log.Warnf("Failed to describe ECS service %s, ECS service report may be incomplete: %s", failure.Arn, failure.Reason)
				}

				lock.Lock()
				for _, service := range resp.Services {
					for _, deployment := range service.Deployments {
						results[*deployment.Id] = *service.ServiceName
					}
				}
				lock.Unlock()
			}()
			return true
		},
	)
	group.Wait()

	if err != nil {
		// We want to still return partial results if we have any, so just log a warning
		log.Warnf("Error listing ECS services, ECS service report may be incomplete: %v", err)
	}
	return results
}

// returns a map from task ARNs to deployment ids
func (c ecsClient) getTaskDeployments(taskArns []string) (map[string]string, error) {
	taskPtrs := make([]*string, len(taskArns))
	for i := range taskArns {
		taskPtrs[i] = &taskArns[i]
	}

	// You'd think there's a limit on how many tasks can be described here,
	// but the docs don't mention anything.
	resp, err := c.client.DescribeTasks(&ecs.DescribeTasksInput{
		Cluster: &c.cluster,
		Tasks:   taskPtrs,
	})
	if err != nil {
		return nil, err
	}

	for _, failure := range resp.Failures {
		// log the failures but still continue with what succeeded
		log.Warnf("Failed to describe ECS task %s, ECS service report may be incomplete: %s", failure.Arn, failure.Reason)
	}

	results := make(map[string]string)
	for _, task := range resp.Tasks {
		results[*task.TaskArn] = *task.StartedBy
	}
	return results, nil
}

// returns a map from task ARNs to service names
func (c ecsClient) getTaskServices(taskArns []string) (map[string]string, error) {
	deploymentMapChan := make(chan map[string]string)
	go func() {
		deploymentMapChan <- c.getDeploymentMap()
	}()

	// do these two fetches in parallel
	taskDeployments, err := c.getTaskDeployments(taskArns)
	deploymentMap := <-deploymentMapChan

	if err != nil {
		return nil, err
	}

	results := make(map[string]string)
	for taskArn, depID := range taskDeployments {
		// Note not all tasks map to a deployment, or we could otherwise mismatch due to races.
		// It's safe to just ignore all these cases and consider them "non-service" tasks.
		if service, ok := deploymentMap[depID]; ok {
			results[taskArn] = service
		}
	}

	return results, nil
}
