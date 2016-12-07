package awsecs

import (
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

// a wrapper around an AWS client that makes all the needed calls and just exposes the final results
type ecsClient struct {
	client       *ecs.ECS
	cluster      string
	taskCache    map[string]ecsTask
	serviceCache map[string]ecsService
}

// Since we're caching tasks heavily, we ensure no mistakes by casting into a structure
// that only contains immutable attributes of the resource.
type ecsTask struct {
	taskARN           string
	createdAt         time.Time
	taskDefinitionARN string

	// These started fields are immutable once set, and guarenteed to be set once the task is running,
	// which we know it is because otherwise we wouldn't be looking at it.
	startedAt time.Time
	startedBy string // tag or deployment id

	// Metadata about this cache copy
	fetchedAt  time.Time
	lastUsedAt time.Time
}

// Services are highly mutable and so we can only cache them on a best-effort basis.
// We have to refresh referenced (ie. has an associated task) services each report
// but we avoid re-listing services unless we can't find a service for a task.
type ecsService struct {
	serviceName string

	// The following values may be stale in a cached copy
	deploymentIDs     []string
	desiredCount      int64
	pendingCount      int64
	runningCount      int64
	taskDefinitionARN string

	// Metadata about this cache copy
	fetchedAt  time.Time
	lastUsedAt time.Time
}

type ecsInfo struct {
	tasks          map[string]ecsTask
	services       map[string]ecsService
	taskServiceMap map[string]string
}

func newClient(cluster string) (*ecsClient, error) {
	sess := session.New()

	region, err := ec2metadata.New(sess).Region()
	if err != nil {
		return nil, err
	}

	return &ecsClient{
		client:       ecs.New(sess, &aws.Config{Region: aws.String(region)}),
		cluster:      cluster,
		taskCache:    map[string]ecsTask{},
		serviceCache: map[string]ecsService{},
	}, nil
}

func newECSTask(task *ecs.Task) ecsTask {
	now := time.Now()
	return ecsTask{
		taskARN:           *task.TaskArn,
		createdAt:         *task.CreatedAt,
		taskDefinitionARN: *task.TaskDefinitionArn,
		startedAt:         *task.StartedAt,
		startedBy:         *task.StartedBy,
		fetchedAt:         now,
		lastUsedAt:        now,
	}
}

func newECSService(service *ecs.Service) ecsService {
	now := time.Now()
	deploymentIDs := make([]string, 0, len(service.Deployments))
	for _, deployment := range service.Deployments {
		deploymentIDs = append(deploymentIDs, *deployment.Id)
	}
	return ecsService{
		serviceName:       *service.ServiceName,
		deploymentIDs:     deploymentIDs,
		desiredCount:      *service.DesiredCount,
		pendingCount:      *service.PendingCount,
		runningCount:      *service.RunningCount,
		taskDefinitionARN: *service.TaskDefinition,
		fetchedAt:         now,
		lastUsedAt:        now,
	}
}

// Returns a channel from which service ARNs can be read.
// Cannot fail as it will attempt to deliver partial results, though that may end up being no results.
func (c ecsClient) listServices() <-chan string {
	results := make(chan string)
	go func() {
		err := c.client.ListServicesPages(
			&ecs.ListServicesInput{Cluster: &c.cluster},
			func(page *ecs.ListServicesOutput, lastPage bool) bool {
				for _, arn := range page.ServiceArns {
					results <- *arn
				}
				return true
			},
		)
		if err != nil {
			log.Warnf("Error listing ECS services, ECS service report may be incomplete: %v", err)
		}
		close(results)
	}()
	return results
}

// Returns (input, done) channels. Service ARNs given to input are batched and details are fetched,
// with full ecsService objects being put into the cache. Closes done when finished.
func (c ecsClient) describeServices() (chan<- string, <-chan bool) {
	input := make(chan string)
	done := make(chan bool)

	go func() {
		const MAX_SERVICES = 10 // How many services we can put in one Describe command
		group := sync.WaitGroup{}
		lock := sync.Mutex{} // mediates access to the service cache when writing results

		page := make([]string, 0, MAX_SERVICES)
		for arn := range input {
			page = append(page, arn)
			if len(page) == MAX_SERVICES {
				group.Add(1)

				go func(arns []string) {
					defer group.Done()

					arnPtrs := make([]*string, 0, len(arns))
					for i := range arns {
						arnPtrs = append(arnPtrs, &arns[i])
					}

					resp, err := c.client.DescribeServices(&ecs.DescribeServicesInput{
						Cluster:  &c.cluster,
						Services: arnPtrs,
					})
					if err != nil {
						log.Warnf("Error describing some ECS services, ECS service report may be incomplete: %v", err)
						return
					}

					for _, failure := range resp.Failures {
						log.Warnf("Failed to describe ECS service %s, ECS service report may be incomplete: %s", *failure.Arn, failure.Reason)
					}

					lock.Lock()
					for _, service := range resp.Services {
						c.serviceCache[*service.ServiceName] = newECSService(service)
					}
					lock.Unlock()
				}(page)

				page = make([]string, 0, MAX_SERVICES)
			}
		}

		group.Wait()
		close(done)
	}()

	return input, done
}

// get details on given tasks, updating cache with the results
func (c ecsClient) getTasks(taskARNs []string) {
	taskPtrs := make([]*string, len(taskARNs))
	for i := range taskARNs {
		taskPtrs[i] = &taskARNs[i]
	}

	// You'd think there's a limit on how many tasks can be described here,
	// but the docs don't mention anything.
	resp, err := c.client.DescribeTasks(&ecs.DescribeTasksInput{
		Cluster: &c.cluster,
		Tasks:   taskPtrs,
	})
	if err != nil {
		log.Warnf("Failed to describe ECS tasks, ECS service report may be incomplete: %v", err)
		return
	}

	for _, failure := range resp.Failures {
		log.Warnf("Failed to describe ECS task %s, ECS service report may be incomplete: %s", *failure.Arn, *failure.Reason)
	}

	for _, task := range resp.Tasks {
		c.taskCache[*task.TaskArn] = newECSTask(task)
	}
}

// Evict entries from the caches which have not been used within the eviction interval.
func (c ecsClient) evictOldCacheItems() {
	const EVICT_TIME = time.Minute
	now := time.Now()

	for arn, task := range c.taskCache {
		if now.Sub(task.lastUsedAt) > EVICT_TIME {
			delete(c.taskCache, arn)
		}
	}

	for name, service := range c.serviceCache {
		if now.Sub(service.lastUsedAt) > EVICT_TIME {
			delete(c.serviceCache, name)
		}
	}
}

// Try to match a list of task ARNs to service names using cached info.
// Returns (task to service map, unmatched tasks). Ignores tasks whose startedby values
// don't appear to point to a service.
func (c ecsClient) matchTasksServices(taskARNs []string) (map[string]string, []string) {
	const SERVICE_PREFIX = "aws-svc" // TODO confirm this

	deploymentMap := map[string]string{}
	for serviceName, service := range c.serviceCache {
		for _, deployment := range service.deploymentIDs {
			deploymentMap[deployment] = serviceName
		}
	}

	results := map[string]string{}
	unmatched := []string{}
	for _, taskARN := range taskARNs {
		task, ok := c.taskCache[taskARN]
		if !ok {
			// this can happen if we have a failure while describing tasks, just pretend the task doesn't exist
			continue
		}
		if !strings.HasPrefix(task.startedBy, SERVICE_PREFIX) {
			// task was not started by a service
			continue
		}
		if serviceName, ok := deploymentMap[task.startedBy]; ok {
			results[taskARN] = serviceName
		} else {
			unmatched = append(unmatched, taskARN)
		}
	}

	return results, unmatched
}

// Returns a ecsInfo struct containing data needed for a report.
func (c ecsClient) getInfo(taskARNs []string) ecsInfo {

	// We do a weird order of operations here to minimize unneeded cache refreshes.
	// First, we ensure we have all the tasks we need, and fetch the ones we don't.
	// We also mark the tasks as being used here to prevent eviction.
	tasksToFetch := []string{}
	now := time.Now()
	for _, taskARN := range taskARNs {
		if task, ok := c.taskCache[taskARN]; ok {
			task.lastUsedAt = now
		} else {
			tasksToFetch = append(tasksToFetch, taskARN)
		}
	}
	// This might not fully succeed, but we only try once and ignore any further missing tasks.
	c.getTasks(tasksToFetch)

	// We're going to do this matching process potentially several times, but that's ok - it's quite cheap.
	// First, we want to see how far we get with existing data, and identify the set of services
	// we'll need to refresh regardless.
	taskServiceMap, unmatched := c.matchTasksServices(taskARNs)

	// In order to ensure service details are fresh, we need to refresh any referenced services
	toDescribe, done := c.describeServices()
	servicesRefreshed := map[string]bool{}
	for _, serviceName := range taskServiceMap {
		if servicesRefreshed[serviceName] {
			continue
		}
		toDescribe <- serviceName
		servicesRefreshed[serviceName] = true
	}
	close(toDescribe)
	<-done

	// In refreshing, we may have picked up any new deployment ids.
	// If we still have tasks unmatched, we try again.
	if len(unmatched) > 0 {
		taskServiceMap, unmatched = c.matchTasksServices(taskARNs)
	}

	// If we still have tasks unmatched, we'll have to try harder. Get a list of all services and,
	// if not already refreshed, fetch them.
	if len(unmatched) > 0 {
		serviceNamesChan := c.listServices()
		toDescribe, done := c.describeServices()
		go func() {
			for serviceName := range serviceNamesChan {
				if !servicesRefreshed[serviceName] {
					toDescribe <- serviceName
					servicesRefreshed[serviceName] = true
				}
				close(toDescribe)
			}
		}()
		<-done

		taskServiceMap, unmatched = c.matchTasksServices(taskARNs)
		// If we still have unmatched at this point, we don't care - this may be due to partial failures,
		// race conditions, and other weirdness.
	}

	// The maps to return are the referenced subsets of the full caches
	tasks := map[string]ecsTask{}
	for _, taskARN := range taskARNs {
		// It's possible that tasks could still be missing from the cache if describe tasks failed.
		// We'll just pretend they don't exist.
		if task, ok := c.taskCache[taskARN]; ok {
			tasks[taskARN] = task
		}
	}

	services := map[string]ecsService{}
	for taskARN, serviceName := range taskServiceMap {
		if _, ok := taskServiceMap[serviceName]; ok {
			// Already present. This is expected since multiple tasks can map to the same service.
			continue
		}
		if service, ok := c.serviceCache[serviceName]; ok {
			services[serviceName] = service
		} else {
			log.Errorf("Service %s referenced by task %s in service map but not found in cache, this shouldn't be able to happen. Removing task and continuing.", serviceName, taskARN)
			delete(taskServiceMap, taskARN)
		}
	}

	c.evictOldCacheItems()

	return ecsInfo{services: services, tasks: tasks, taskServiceMap: taskServiceMap}
}
