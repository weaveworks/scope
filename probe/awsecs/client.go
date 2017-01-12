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
	"github.com/bluele/gcache"
)

// A wrapper around an AWS client that makes all the needed calls and just exposes the final results.
// We create an interface so we can mock for testing
type ecsClient interface {
	// Returns a ecsInfo struct containing data needed for a report.
	getInfo([]string) ecsInfo
}

// actual implementation
type ecsClientImpl struct {
	client       *ecs.ECS
	cluster      string
	taskCache    gcache.Cache // Keys are task ARNs.
	serviceCache gcache.Cache // Keys are service names.
}

// Since we're caching tasks heavily, we ensure no mistakes by casting into a structure
// that only contains immutable attributes of the resource.
type ecsTask struct {
	taskARN           string
	createdAt         time.Time
	taskDefinitionARN string

	// These started fields are immutable once set, and guaranteed to be set once the task is running,
	// which we know it is because otherwise we wouldn't be looking at it.
	startedAt time.Time
	startedBy string // tag or deployment id
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
}

type ecsInfo struct {
	tasks          map[string]ecsTask
	services       map[string]ecsService
	taskServiceMap map[string]string
}

func newClient(cluster string, cacheSize int, cacheExpiry time.Duration) (*ecsClient, error) {
	sess := session.New()

	region, err := ec2metadata.New(sess).Region()
	if err != nil {
		return nil, err
	}

	return &ecsClientImpl{
		client:       ecs.New(sess, &aws.Config{Region: aws.String(region)}),
		cluster:      cluster,
		taskCache:    gcache.New(cacheSize).LRU().Expiration(cacheExpiry).Build(),
		serviceCache: gcache.New(cacheSize).LRU().Expiration(cacheExpiry).Build(),
	}, nil
}

func newECSTask(task *ecs.Task) ecsTask {
	return ecsTask{
		taskARN:           *task.TaskArn,
		createdAt:         *task.CreatedAt,
		taskDefinitionARN: *task.TaskDefinitionArn,
		startedAt:         *task.StartedAt,
		startedBy:         *task.StartedBy,
	}
}

func newECSService(service *ecs.Service) ecsService {
	deploymentIDs := make([]string, len(service.Deployments))
	for i, deployment := range service.Deployments {
		deploymentIDs[i] = *deployment.Id
	}
	return ecsService{
		serviceName:       *service.ServiceName,
		deploymentIDs:     deploymentIDs,
		desiredCount:      *service.DesiredCount,
		pendingCount:      *service.PendingCount,
		runningCount:      *service.RunningCount,
		taskDefinitionARN: *service.TaskDefinition,
	}
}

// Returns a channel from which service ARNs can be read.
// Cannot fail as it will attempt to deliver partial results, though that may end up being no results.
func (c ecsClientImpl) listServices() <-chan string {
	log.Debugf("Listing ECS services")
	results := make(chan string)
	go func() {
		count := 0
		err := c.client.ListServicesPages(
			&ecs.ListServicesInput{Cluster: &c.cluster},
			func(page *ecs.ListServicesOutput, lastPage bool) bool {
				for _, arn := range page.ServiceArns {
					count++
					results <- *arn
				}
				return true
			},
		)
		if err != nil {
			log.Warnf("Error listing ECS services, ECS service report may be incomplete: %v", err)
		}
		log.Debugf("Listed %d services", count)
		close(results)
	}()
	return results
}

// Returns (input, done) channels. Service ARNs given to input are batched and details are fetched,
// with full ecsService objects being put into the cache. Closes done when finished.
func (c ecsClientImpl) describeServices() (chan<- string, <-chan struct{}) {
	input := make(chan string)
	done := make(chan struct{})

	log.Debugf("Describing ECS services")

	go func() {
		const maxServices = 10 // How many services we can put in one Describe command
		group := sync.WaitGroup{}

		// count and calls is just for logging
		count := 0
		calls := 0

		batch := make([]string, 0, maxServices)
		for arn := range input {
			batch = append(batch, arn)
			if len(batch) == maxServices {
				group.Add(1)
				go func(arns []string) {
					defer group.Done()
					c.describeServicesBatch(arns)
				}(batch)
				count += len(batch)
				calls++
				batch = make([]string, 0, maxServices)
			}
		}
		if len(batch) > 0 {
			c.describeServicesBatch(batch)
			count += len(batch)
			calls++
		}

		log.Debugf("Described %d services in %d calls", count, calls)
		group.Wait()
		close(done)
	}()

	return input, done
}

func (c ecsClientImpl) describeServicesBatch(arns []string) {
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

	for _, service := range resp.Services {
		c.serviceCache.Set(*service.ServiceName, newECSService(service))
	}
}

// get details on given tasks, updating cache with the results
func (c ecsClientImpl) getTasks(taskARNs []string) {
	log.Debugf("Describing %d ECS tasks", len(taskARNs))

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
		c.taskCache.Set(*task.TaskArn, newECSTask(task))
	}
}

// Try to match a list of task ARNs to service names using cached info.
// Returns (task to service map, unmatched tasks). Ignores tasks whose startedby values
// don't appear to point to a service.
func (c ecsClientImpl) matchTasksServices(taskARNs []string) (map[string]string, []string) {
	const servicePrefix = "ecs-svc"

	deploymentMap := map[string]string{}
	for _, serviceNameRaw := range c.serviceCache.Keys() {
		serviceRaw, err := c.serviceCache.Get(serviceNameRaw)
		if err != nil {
			// This is rare, but possible if service was evicted after the loop began
			continue
		}
		serviceName := serviceNameRaw.(string)
		service := serviceRaw.(ecsService)
		for _, deployment := range service.deploymentIDs {
			deploymentMap[deployment] = serviceName
		}
	}
	log.Debugf("Mapped %d deployments from %d services", len(deploymentMap), c.serviceCache.Len())

	results := map[string]string{}
	unmatched := []string{}
	for _, taskARN := range taskARNs {
		taskRaw, err := c.taskCache.Get(taskARN)
		if err != nil {
			// this can happen if we have a failure while describing tasks, just pretend the task doesn't exist
			continue
		}
		task := taskRaw.(ecsTask)
		if !strings.HasPrefix(task.startedBy, servicePrefix) {
			// task was not started by a service
			continue
		}
		if serviceName, ok := deploymentMap[task.startedBy]; ok {
			results[taskARN] = serviceName
		} else {
			unmatched = append(unmatched, taskARN)
		}
	}

	log.Debugf("Matched %d from %d tasks, %d unmatched", len(results), len(taskARNs), len(unmatched))
	return results, unmatched
}

func (c ecsClientImpl) ensureTasks(taskARNs []string) {
	tasksToFetch := []string{}
	for _, taskARN := range taskARNs {
		if _, err := c.taskCache.Get(taskARN); err != nil {
			tasksToFetch = append(tasksToFetch, taskARN)
		}
	}
	if len(tasksToFetch) > 0 {
		// This might not fully succeed, but we only try once and ignore any further missing tasks.
		c.getTasks(tasksToFetch)
	}
}

func (c ecsClientImpl) refreshServices(taskServiceMap map[string]string) map[string]bool {
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
	return servicesRefreshed
}

func (c ecsClientImpl) describeAllServices(servicesRefreshed map[string]bool) {
	serviceNamesChan := c.listServices()
	toDescribe, done := c.describeServices()
	go func() {
		for serviceName := range serviceNamesChan {
			if !servicesRefreshed[serviceName] {
				toDescribe <- serviceName
				servicesRefreshed[serviceName] = true
			}
		}
		close(toDescribe)
	}()
	<-done
}

func (c ecsClientImpl) makeECSInfo(taskARNs []string, taskServiceMap map[string]string) ecsInfo {
	// The maps to return are the referenced subsets of the full caches
	tasks := map[string]ecsTask{}
	for _, taskARN := range taskARNs {
		// It's possible that tasks could still be missing from the cache if describe tasks failed.
		// We'll just pretend they don't exist.
		if taskRaw, err := c.taskCache.Get(taskARN); err == nil {
			task := taskRaw.(ecsTask)
			tasks[taskARN] = task
		}
	}

	services := map[string]ecsService{}
	for taskARN, serviceName := range taskServiceMap {
		if _, ok := taskServiceMap[serviceName]; ok {
			// Already present. This is expected since multiple tasks can map to the same service.
			continue
		}
		if serviceRaw, err := c.serviceCache.Get(serviceName); err == nil {
			service := serviceRaw.(ecsService)
			services[serviceName] = service
		} else {
			log.Errorf("Service %s referenced by task %s in service map but not found in cache, this shouldn't be able to happen. Removing task and continuing.", serviceName, taskARN)
			delete(taskServiceMap, taskARN)
		}
	}

	return ecsInfo{services: services, tasks: tasks, taskServiceMap: taskServiceMap}
}

// Implements ecsClient.getInfo
func (c ecsClientImpl) getInfo(taskARNs []string) ecsInfo {
	log.Debugf("Getting ECS info on %d tasks", len(taskARNs))

	// We do a weird order of operations here to minimize unneeded cache refreshes.
	// First, we ensure we have all the tasks we need, and fetch the ones we don't.
	// We also mark the tasks as being used here to prevent eviction.
	c.ensureTasks(taskARNs)

	// We're going to do this matching process potentially several times, but that's ok - it's quite cheap.
	// First, we want to see how far we get with existing data, and identify the set of services
	// we'll need to refresh regardless.
	taskServiceMap, unmatched := c.matchTasksServices(taskARNs)

	// In order to ensure service details are fresh, we need to refresh any referenced services
	log.Debugf("Refreshing ECS services")
	servicesRefreshed := c.refreshServices(taskServiceMap)

	// In refreshing, we may have picked up any new deployment ids.
	// If we still have tasks unmatched, we try again.
	if len(unmatched) > 0 {
		taskServiceMap, unmatched = c.matchTasksServices(taskARNs)
	}

	// If we still have tasks unmatched, we'll have to try harder. Get a list of all services and,
	// if not already refreshed, fetch them.
	log.Debugf("After refreshing services, %d tasks unmatched", len(unmatched))
	if len(unmatched) > 0 {
		c.describeAllServices(servicesRefreshed)

		taskServiceMap, unmatched = c.matchTasksServices(taskARNs)
		// If we still have unmatched at this point, we don't care - this may be due to partial failures,
		// race conditions, and other weirdness.
	}

	info := c.makeECSInfo(taskARNs, taskServiceMap)

	return info
}
