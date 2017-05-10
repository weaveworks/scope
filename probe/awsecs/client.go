package awsecs

import (
	"fmt"
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

const servicePrefix = "ecs-svc" // Task StartedBy field begins with this if it was started by a service

// EcsClient is a wrapper around an AWS client that makes all the needed calls and just exposes the final results.
// We create an interface so we can mock for testing.
type EcsClient interface {
	// Returns a EcsInfo struct containing data needed for a report.
	GetInfo([]string) EcsInfo
	// Scales a service up or down by amount
	ScaleService(string, int) error
}

// actual implementation
type ecsClientImpl struct {
	client       *ecs.ECS
	cluster      string
	taskCache    gcache.Cache // Keys are task ARNs.
	serviceCache gcache.Cache // Keys are service names.
}

// EcsTask describes the parts of ECS tasks we care about.
// Since we're caching tasks heavily, we ensure no mistakes by casting into a structure
// that only contains immutable attributes of the resource.
// Exported for test.
type EcsTask struct {
	TaskARN           string
	CreatedAt         time.Time
	TaskDefinitionARN string

	// These started fields are immutable once set, and guaranteed to be set once the task is running,
	// which we know it is because otherwise we wouldn't be looking at it.
	StartedAt time.Time
	StartedBy string // tag or deployment id
}

// EcsService describes the parts of ECS services we care about.
// Services are highly mutable and so we can only cache them on a best-effort basis.
// We have to refresh referenced (ie. has an associated task) services each report
// but we avoid re-listing services unless we can't find a service for a task.
// Exported for test.
type EcsService struct {
	ServiceName string
	// The following values may be stale in a cached copy
	DeploymentIDs     []string
	DesiredCount      int64
	PendingCount      int64
	RunningCount      int64
	TaskDefinitionARN string
}

// EcsInfo is exported for test
type EcsInfo struct {
	Tasks          map[string]EcsTask
	Services       map[string]EcsService
	TaskServiceMap map[string]string
}

func newClient(cluster string, cacheSize int, cacheExpiry time.Duration) (EcsClient, error) {
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

func newECSTask(task *ecs.Task) EcsTask {
	return EcsTask{
		TaskARN:           *task.TaskArn,
		CreatedAt:         *task.CreatedAt,
		TaskDefinitionARN: *task.TaskDefinitionArn,
		StartedAt:         *task.StartedAt,
		StartedBy:         *task.StartedBy,
	}
}

func newECSService(service *ecs.Service) EcsService {
	deploymentIDs := make([]string, len(service.Deployments))
	for i, deployment := range service.Deployments {
		deploymentIDs[i] = *deployment.Id
	}
	return EcsService{
		ServiceName:       *service.ServiceName,
		DeploymentIDs:     deploymentIDs,
		DesiredCount:      *service.DesiredCount,
		PendingCount:      *service.PendingCount,
		RunningCount:      *service.RunningCount,
		TaskDefinitionARN: *service.TaskDefinition,
	}
}

// IsServiceManaged returns true if the task was started by a service.
func (t EcsTask) IsServiceManaged() bool {
	return strings.HasPrefix(t.StartedBy, servicePrefix)
}

// Fetches a task from the cache, returning (task, ok) as per map[]
func (c ecsClientImpl) getCachedTask(taskARN string) (EcsTask, bool) {
	if taskRaw, err := c.taskCache.Get(taskARN); err == nil {
		return taskRaw.(EcsTask), true
	}
	return EcsTask{}, false
}

// Fetches a service from the cache, returning (service, ok) as per map[]
func (c ecsClientImpl) getCachedService(serviceName string) (EcsService, bool) {
	if serviceRaw, err := c.serviceCache.Get(serviceName); err == nil {
		return serviceRaw.(EcsService), true
	}
	return EcsService{}, false
}

// Returns a list of service names.
// Cannot fail as it will attempt to deliver partial results, though that may end up being no results.
func (c ecsClientImpl) listServices() []string {
	log.Debugf("Listing ECS services")
	results := []string{}
	err := c.client.ListServicesPages(
		&ecs.ListServicesInput{Cluster: &c.cluster},
		func(page *ecs.ListServicesOutput, lastPage bool) bool {
			if page == nil {
				return true
			}
			for _, name := range page.ServiceArns {
				if name != nil {
					results = append(results, *name)
				}
			}
			return true
		},
	)
	if err != nil {
		log.Warnf("Error listing ECS services, ECS service report may be incomplete: %v", err)
	}
	log.Debugf("Listed %d services", len(results))
	return results
}

// Service names given are batched and details are fetched,
// with full EcsService objects being put into the cache.
// Cannot fail as it will attempt to deliver partial results.
func (c ecsClientImpl) describeServices(services []string) {
	const maxServices = 10 // How many services we can put in one Describe command
	group := sync.WaitGroup{}

	log.Debugf("Describing ECS services")

	// split into batches
	batches := make([][]string, 0, len(services)/maxServices+1)
	for len(services) > maxServices {
		batch := services[:maxServices]
		services = services[maxServices:]
		batches = append(batches, batch)
	}
	if len(services) > 0 {
		batches = append(batches, services)
	}

	for _, batch := range batches {
		group.Add(1)
		go func(names []string) {
			defer group.Done()
			c.describeServicesBatch(names)
		}(batch)
	}

	group.Wait()
}

func (c ecsClientImpl) describeServicesBatch(names []string) {
	namePtrs := make([]*string, 0, len(names))
	for i := range names {
		namePtrs = append(namePtrs, &names[i])
	}

	resp, err := c.client.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  &c.cluster,
		Services: namePtrs,
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
		if task.TaskArn != nil {
			c.taskCache.Set(*task.TaskArn, newECSTask(task))
		}
	}
}

// Try to match a list of task ARNs to service names using cached info.
// Returns (task to service map, unmatched tasks). Ignores tasks whose startedby values
// don't appear to point to a service.
func (c ecsClientImpl) matchTasksServices(taskARNs []string) (map[string]string, []string) {
	deploymentMap := map[string]string{}
	for _, serviceNameRaw := range c.serviceCache.Keys() {
		serviceName := serviceNameRaw.(string)
		service, ok := c.getCachedService(serviceName)
		if !ok {
			// This is rare, but possible if service was evicted after the loop began
			continue
		}
		for _, deployment := range service.DeploymentIDs {
			deploymentMap[deployment] = serviceName
		}
	}
	log.Debugf("Mapped %d deployments from %d services", len(deploymentMap), c.serviceCache.Len())

	results := map[string]string{}
	unmatched := []string{}
	for _, taskARN := range taskARNs {
		task, ok := c.getCachedTask(taskARN)
		if !ok {
			// this can happen if we have a failure while describing tasks, just pretend the task doesn't exist
			continue
		}
		if !task.IsServiceManaged() {
			continue
		}
		if serviceName, ok := deploymentMap[task.StartedBy]; ok {
			results[taskARN] = serviceName
		} else {
			unmatched = append(unmatched, taskARN)
		}
	}

	log.Debugf("Matched %d from %d tasks, %d unmatched", len(results), len(taskARNs), len(unmatched))
	return results, unmatched
}

func (c ecsClientImpl) ensureTasksAreCached(taskARNs []string) {
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
	servicesRefreshed := map[string]bool{}
	toDescribe := []string{}
	for _, serviceName := range taskServiceMap {
		if servicesRefreshed[serviceName] {
			continue
		}
		toDescribe = append(toDescribe, serviceName)
		servicesRefreshed[serviceName] = true
	}
	c.describeServices(toDescribe)
	return servicesRefreshed
}

func (c ecsClientImpl) describeAllServices(servicesRefreshed map[string]bool) {
	toDescribe := []string{}
	for _, serviceName := range c.listServices() {
		if !servicesRefreshed[serviceName] {
			toDescribe = append(toDescribe, serviceName)
			servicesRefreshed[serviceName] = true
		}
	}
	c.describeServices(toDescribe)
}

func (c ecsClientImpl) makeECSInfo(taskARNs []string, taskServiceMap map[string]string) EcsInfo {
	// The maps to return are the referenced subsets of the full caches
	tasks := map[string]EcsTask{}
	for _, taskARN := range taskARNs {
		// It's possible that tasks could still be missing from the cache if describe tasks failed.
		// We'll just pretend they don't exist.
		if task, ok := c.getCachedTask(taskARN); ok {
			tasks[taskARN] = task
		}
	}

	services := map[string]EcsService{}
	for taskARN, serviceName := range taskServiceMap {
		if _, ok := taskServiceMap[serviceName]; ok {
			// Already present. This is expected since multiple tasks can map to the same service.
			continue
		}
		if service, ok := c.getCachedService(serviceName); ok {
			services[serviceName] = service
		} else {
			log.Errorf("Service %s referenced by task %s in service map but not found in cache, this shouldn't be able to happen. Removing task and continuing.", serviceName, taskARN)
			delete(taskServiceMap, taskARN)
		}
	}

	return EcsInfo{Services: services, Tasks: tasks, TaskServiceMap: taskServiceMap}
}

// Implements EcsClient.GetInfo
func (c ecsClientImpl) GetInfo(taskARNs []string) EcsInfo {
	log.Debugf("Getting ECS info on %d tasks", len(taskARNs))

	// We do a weird order of operations here to minimize unneeded cache refreshes.
	// First, we ensure we have all the tasks we need, and fetch the ones we don't.
	// We also mark the tasks as being used here to prevent eviction.
	c.ensureTasksAreCached(taskARNs)

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

// Implements EcsClient.ScaleService
func (c ecsClientImpl) ScaleService(serviceName string, amount int) error {
	// Note this is inherently racey, due to needing to get, modify, then update the DesiredCount.

	// refresh service in cache
	c.describeServices([]string{serviceName})
	// now check the cache to see if it worked
	service, ok := c.getCachedService(serviceName)
	if !ok {
		return fmt.Errorf("Service %s not found", serviceName)
	}

	newCount := service.DesiredCount + int64(amount)
	if newCount < 1 {
		return fmt.Errorf("Cannot reduce count below one")
	}
	_, err := c.client.UpdateService(&ecs.UpdateServiceInput{
		Cluster:      &c.cluster,
		Service:      &serviceName,
		DesiredCount: &newCount,
	})
	return err
}
