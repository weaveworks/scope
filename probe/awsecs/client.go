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

type ecsInfo struct {
	tasks          map[string]*ecs.Task
	services       map[string]*ecs.Service
	taskServiceMap map[string]string
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
func (c ecsClient) getDeploymentMap(services map[string]*ecs.Service) map[string]string {
	results := map[string]string{}
	for serviceName, service := range services {
		for _, deployment := range service.Deployments {
			results[*deployment.Id] = serviceName
		}
	}
	return results
}

// cannot fail as it will attempt to deliver partial results, though that may end up being no results
func (c ecsClient) getServices() map[string]*ecs.Service {
	results := map[string]*ecs.Service{}
	lock := sync.Mutex{} // lock mediates access to results

	group := sync.WaitGroup{}

	err := c.client.ListServicesPages(
		&ecs.ListServicesInput{Cluster: &c.cluster},
		func(page *ecs.ListServicesOutput, lastPage bool) bool {
			// describe each page of 10 (the max for one describe command) concurrently
			group.Add(1)
			serviceArns := page.ServiceArns
			go func() {
				defer group.Done()

				resp, err := c.client.DescribeServices(&ecs.DescribeServicesInput{
					Cluster:  &c.cluster,
					Services: serviceArns,
				})
				if err != nil {
					log.Warnf("Error describing some ECS services, ECS service report may be incomplete: %v", err)
					return
				}

				for _, failure := range resp.Failures {
					log.Warnf("Failed to describe ECS service %s, ECS service report may be incomplete: %s", failure.Arn, failure.Reason)
				}

				lock.Lock()
				for _, service := range resp.Services {
					results[*service.ServiceName] = service
				}
				lock.Unlock()
			}()
			return true
		},
	)
	group.Wait()

	if err != nil {
		log.Warnf("Error listing ECS services, ECS service report may be incomplete: %v", err)
	}
	return results
}

func (c ecsClient) getTasks(taskArns []string) (map[string]*ecs.Task, error) {
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
		log.Warnf("Failed to describe ECS task %s, ECS service report may be incomplete: %s", failure.Arn, failure.Reason)
	}

	results := make(map[string]*ecs.Task, len(resp.Tasks))
	for _, task := range resp.Tasks {
		results[*task.TaskArn] = task
	}
	return results, nil
}

// returns a map from task ARNs to service names
func (c ecsClient) getInfo(taskArns []string) (ecsInfo, error) {
	servicesChan := make(chan map[string]*ecs.Service)
	go func() {
		servicesChan <- c.getServices()
	}()

	// do these two fetches in parallel
	tasks, err := c.getTasks(taskArns)
	services := <-servicesChan

	if err != nil {
		return ecsInfo{}, err
	}

	deploymentMap := c.getDeploymentMap(services)

	taskServiceMap := map[string]string{}
	for taskArn, task := range tasks {
		// Note not all tasks map to a deployment, or we could otherwise mismatch due to races.
		// It's safe to just ignore all these cases and consider them "non-service" tasks.
		if serviceName, ok := deploymentMap[*task.StartedBy]; ok {
			taskServiceMap[taskArn] = serviceName
		}
	}

	return ecsInfo{services: services, tasks: tasks, taskServiceMap: taskServiceMap}, nil
}
