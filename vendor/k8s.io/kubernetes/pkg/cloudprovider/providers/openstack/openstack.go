/*
Copyright 2014 The Kubernetes Authors All rights reserved.

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

package openstack

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/blockstorage/v1/volumes"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/volumeattach"
	"github.com/rackspace/gophercloud/openstack/compute/v2/flavors"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/lbaas/members"
	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/lbaas/monitors"
	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/lbaas/pools"
	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/lbaas/vips"
	"github.com/rackspace/gophercloud/pagination"
	"gopkg.in/gcfg.v1"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/service"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

const ProviderName = "openstack"

// metadataUrl is URL to OpenStack metadata server. It's hadrcoded IPv4
// link-local address as documented in "OpenStack Cloud Administrator Guide",
// chapter Compute - Networking with nova-network.
// http://docs.openstack.org/admin-guide-cloud/compute-networking-nova.html#metadata-service
const metadataUrl = "http://169.254.169.254/openstack/2012-08-10/meta_data.json"

var ErrNotFound = errors.New("Failed to find object")
var ErrMultipleResults = errors.New("Multiple results where only one expected")
var ErrNoAddressFound = errors.New("No address found for host")
var ErrAttrNotFound = errors.New("Expected attribute not found")

const (
	MiB = 1024 * 1024
	GB  = 1000 * 1000 * 1000
)

// encoding.TextUnmarshaler interface for time.Duration
type MyDuration struct {
	time.Duration
}

func (d *MyDuration) UnmarshalText(text []byte) error {
	res, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	d.Duration = res
	return nil
}

type LoadBalancerOpts struct {
	SubnetId          string     `gcfg:"subnet-id"` // required
	FloatingNetworkId string     `gcfg:"floating-network-id"`
	LBMethod          string     `gcfg:"lb-method"`
	CreateMonitor     bool       `gcfg:"create-monitor"`
	MonitorDelay      MyDuration `gcfg:"monitor-delay"`
	MonitorTimeout    MyDuration `gcfg:"monitor-timeout"`
	MonitorMaxRetries uint       `gcfg:"monitor-max-retries"`
}

// OpenStack is an implementation of cloud provider Interface for OpenStack.
type OpenStack struct {
	provider *gophercloud.ProviderClient
	region   string
	lbOpts   LoadBalancerOpts
	// InstanceID of the server where this OpenStack object is instantiated.
	localInstanceID string
}

type Config struct {
	Global struct {
		AuthUrl    string `gcfg:"auth-url"`
		Username   string
		UserId     string `gcfg:"user-id"`
		Password   string
		ApiKey     string `gcfg:"api-key"`
		TenantId   string `gcfg:"tenant-id"`
		TenantName string `gcfg:"tenant-name"`
		DomainId   string `gcfg:"domain-id"`
		DomainName string `gcfg:"domain-name"`
		Region     string
	}
	LoadBalancer LoadBalancerOpts
}

func init() {
	cloudprovider.RegisterCloudProvider(ProviderName, func(config io.Reader) (cloudprovider.Interface, error) {
		cfg, err := readConfig(config)
		if err != nil {
			return nil, err
		}
		return newOpenStack(cfg)
	})
}

func (cfg Config) toAuthOptions() gophercloud.AuthOptions {
	return gophercloud.AuthOptions{
		IdentityEndpoint: cfg.Global.AuthUrl,
		Username:         cfg.Global.Username,
		UserID:           cfg.Global.UserId,
		Password:         cfg.Global.Password,
		APIKey:           cfg.Global.ApiKey,
		TenantID:         cfg.Global.TenantId,
		TenantName:       cfg.Global.TenantName,
		DomainID:         cfg.Global.DomainId,
		DomainName:       cfg.Global.DomainName,

		// Persistent service, so we need to be able to renew tokens.
		AllowReauth: true,
	}
}

func readConfig(config io.Reader) (Config, error) {
	if config == nil {
		err := fmt.Errorf("no OpenStack cloud provider config file given")
		return Config{}, err
	}

	var cfg Config
	err := gcfg.ReadInto(&cfg, config)
	return cfg, err
}

// parseMetadataUUID reads JSON from OpenStack metadata server and parses
// instance ID out of it.
func parseMetadataUUID(jsonData []byte) (string, error) {
	// We should receive an object with { 'uuid': '<uuid>' } and couple of other
	// properties (which we ignore).

	obj := struct{ UUID string }{}
	err := json.Unmarshal(jsonData, &obj)
	if err != nil {
		return "", err
	}

	uuid := obj.UUID
	if uuid == "" {
		err = fmt.Errorf("cannot parse OpenStack metadata, got empty uuid")
		return "", err
	}

	return uuid, nil
}

func readInstanceID() (string, error) {
	// Try to find instance ID on the local filesystem (created by cloud-init)
	const instanceIDFile = "/var/lib/cloud/data/instance-id"
	idBytes, err := ioutil.ReadFile(instanceIDFile)
	if err == nil {
		instanceID := string(idBytes)
		instanceID = strings.TrimSpace(instanceID)
		glog.V(3).Infof("Got instance id from %s: %s", instanceIDFile, instanceID)
		if instanceID != "" {
			return instanceID, nil
		}
		// Fall through with empty instanceID and try metadata server.
	}
	glog.V(5).Infof("Cannot read %s: '%v', trying metadata server", instanceIDFile, err)

	// Try to get JSON from metdata server.
	resp, err := http.Get(metadataUrl)
	if err != nil {
		glog.V(3).Infof("Cannot read %s: %v", metadataUrl, err)
		return "", err
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("got unexpected status code when reading metadata from %s: %s", metadataUrl, resp.Status)
		glog.V(3).Infof("%v", err)
		return "", err
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.V(3).Infof("Cannot get HTTP response body from %s: %v", metadataUrl, err)
		return "", err
	}
	instanceID, err := parseMetadataUUID(bodyBytes)
	if err != nil {
		glog.V(3).Infof("Cannot parse instance ID from metadata from %s: %v", metadataUrl, err)
		return "", err
	}

	glog.V(3).Infof("Got instance id from %s: %s", metadataUrl, instanceID)
	return instanceID, nil
}

func newOpenStack(cfg Config) (*OpenStack, error) {
	provider, err := openstack.AuthenticatedClient(cfg.toAuthOptions())
	if err != nil {
		return nil, err
	}

	id, err := readInstanceID()
	if err != nil {
		return nil, err
	}

	os := OpenStack{
		provider:        provider,
		region:          cfg.Global.Region,
		lbOpts:          cfg.LoadBalancer,
		localInstanceID: id,
	}

	return &os, nil
}

type Instances struct {
	compute            *gophercloud.ServiceClient
	flavor_to_resource map[string]*api.NodeResources // keyed by flavor id
}

// Instances returns an implementation of Instances for OpenStack.
func (os *OpenStack) Instances() (cloudprovider.Instances, bool) {
	glog.V(4).Info("openstack.Instances() called")

	compute, err := openstack.NewComputeV2(os.provider, gophercloud.EndpointOpts{
		Region: os.region,
	})
	if err != nil {
		glog.Warningf("Failed to find compute endpoint: %v", err)
		return nil, false
	}

	pager := flavors.ListDetail(compute, nil)

	flavor_to_resource := make(map[string]*api.NodeResources)
	err = pager.EachPage(func(page pagination.Page) (bool, error) {
		flavorList, err := flavors.ExtractFlavors(page)
		if err != nil {
			return false, err
		}
		for _, flavor := range flavorList {
			rsrc := api.NodeResources{
				Capacity: api.ResourceList{
					api.ResourceCPU:            *resource.NewQuantity(int64(flavor.VCPUs), resource.DecimalSI),
					api.ResourceMemory:         *resource.NewQuantity(int64(flavor.RAM)*MiB, resource.BinarySI),
					"openstack.org/disk":       *resource.NewQuantity(int64(flavor.Disk)*GB, resource.DecimalSI),
					"openstack.org/rxTxFactor": *resource.NewMilliQuantity(int64(flavor.RxTxFactor)*1000, resource.DecimalSI),
					"openstack.org/swap":       *resource.NewQuantity(int64(flavor.Swap)*MiB, resource.BinarySI),
				},
			}
			flavor_to_resource[flavor.ID] = &rsrc
		}
		return true, nil
	})
	if err != nil {
		glog.Warningf("Failed to find compute flavors: %v", err)
		return nil, false
	}

	glog.V(3).Infof("Found %v compute flavors", len(flavor_to_resource))
	glog.V(1).Info("Claiming to support Instances")

	return &Instances{compute, flavor_to_resource}, true
}

func (i *Instances) List(name_filter string) ([]string, error) {
	glog.V(4).Infof("openstack List(%v) called", name_filter)

	opts := servers.ListOpts{
		Name:   name_filter,
		Status: "ACTIVE",
	}
	pager := servers.List(i.compute, opts)

	ret := make([]string, 0)
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		sList, err := servers.ExtractServers(page)
		if err != nil {
			return false, err
		}
		for _, server := range sList {
			ret = append(ret, server.Name)
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	glog.V(3).Infof("Found %v instances matching %v: %v",
		len(ret), name_filter, ret)

	return ret, nil
}

func getServerByName(client *gophercloud.ServiceClient, name string) (*servers.Server, error) {
	opts := servers.ListOpts{
		Name:   fmt.Sprintf("^%s$", regexp.QuoteMeta(name)),
		Status: "ACTIVE",
	}
	pager := servers.List(client, opts)

	serverList := make([]servers.Server, 0, 1)

	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		s, err := servers.ExtractServers(page)
		if err != nil {
			return false, err
		}
		serverList = append(serverList, s...)
		if len(serverList) > 1 {
			return false, ErrMultipleResults
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	if len(serverList) == 0 {
		return nil, ErrNotFound
	} else if len(serverList) > 1 {
		return nil, ErrMultipleResults
	}

	return &serverList[0], nil
}

func getAddressesByName(client *gophercloud.ServiceClient, name string) ([]api.NodeAddress, error) {
	srv, err := getServerByName(client, name)
	if err != nil {
		return nil, err
	}

	addrs := []api.NodeAddress{}

	for network, netblob := range srv.Addresses {
		list, ok := netblob.([]interface{})
		if !ok {
			continue
		}

		for _, item := range list {
			var addressType api.NodeAddressType

			props, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			extIPType, ok := props["OS-EXT-IPS:type"]
			if (ok && extIPType == "floating") || (!ok && network == "public") {
				addressType = api.NodeExternalIP
			} else {
				addressType = api.NodeInternalIP
			}

			tmp, ok := props["addr"]
			if !ok {
				continue
			}
			addr, ok := tmp.(string)
			if !ok {
				continue
			}

			api.AddToNodeAddresses(&addrs,
				api.NodeAddress{
					Type:    addressType,
					Address: addr,
				},
			)
		}
	}

	// AccessIPs are usually duplicates of "public" addresses.
	if srv.AccessIPv4 != "" {
		api.AddToNodeAddresses(&addrs,
			api.NodeAddress{
				Type:    api.NodeExternalIP,
				Address: srv.AccessIPv4,
			},
		)
	}

	if srv.AccessIPv6 != "" {
		api.AddToNodeAddresses(&addrs,
			api.NodeAddress{
				Type:    api.NodeExternalIP,
				Address: srv.AccessIPv6,
			},
		)
	}

	return addrs, nil
}

func getAddressByName(client *gophercloud.ServiceClient, name string) (string, error) {
	addrs, err := getAddressesByName(client, name)
	if err != nil {
		return "", err
	} else if len(addrs) == 0 {
		return "", ErrNoAddressFound
	}

	for _, addr := range addrs {
		if addr.Type == api.NodeInternalIP {
			return addr.Address, nil
		}
	}

	return addrs[0].Address, nil
}

// Implementation of Instances.CurrentNodeName
func (i *Instances) CurrentNodeName(hostname string) (string, error) {
	return hostname, nil
}

func (i *Instances) AddSSHKeyToAllInstances(user string, keyData []byte) error {
	return errors.New("unimplemented")
}

func (i *Instances) NodeAddresses(name string) ([]api.NodeAddress, error) {
	glog.V(4).Infof("NodeAddresses(%v) called", name)

	addrs, err := getAddressesByName(i.compute, name)
	if err != nil {
		return nil, err
	}

	glog.V(4).Infof("NodeAddresses(%v) => %v", name, addrs)
	return addrs, nil
}

// ExternalID returns the cloud provider ID of the specified instance (deprecated).
func (i *Instances) ExternalID(name string) (string, error) {
	srv, err := getServerByName(i.compute, name)
	if err != nil {
		return "", err
	}
	return srv.ID, nil
}

// InstanceID returns the cloud provider ID of the specified instance.
func (i *Instances) InstanceID(name string) (string, error) {
	srv, err := getServerByName(i.compute, name)
	if err != nil {
		return "", err
	}
	// In the future it is possible to also return an endpoint as:
	// <endpoint>/<instanceid>
	return "/" + srv.ID, nil
}

// InstanceType returns the type of the specified instance.
func (i *Instances) InstanceType(name string) (string, error) {
	return "", nil
}

func (os *OpenStack) Clusters() (cloudprovider.Clusters, bool) {
	return nil, false
}

// ProviderName returns the cloud provider ID.
func (os *OpenStack) ProviderName() string {
	return ProviderName
}

// ScrubDNS filters DNS settings for pods.
func (os *OpenStack) ScrubDNS(nameservers, searches []string) (nsOut, srchOut []string) {
	return nameservers, searches
}

type LoadBalancer struct {
	network *gophercloud.ServiceClient
	compute *gophercloud.ServiceClient
	opts    LoadBalancerOpts
}

func (os *OpenStack) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	glog.V(4).Info("openstack.LoadBalancer() called")

	// TODO: Search for and support Rackspace loadbalancer API, and others.
	network, err := openstack.NewNetworkV2(os.provider, gophercloud.EndpointOpts{
		Region: os.region,
	})
	if err != nil {
		glog.Warningf("Failed to find neutron endpoint: %v", err)
		return nil, false
	}

	compute, err := openstack.NewComputeV2(os.provider, gophercloud.EndpointOpts{
		Region: os.region,
	})
	if err != nil {
		glog.Warningf("Failed to find compute endpoint: %v", err)
		return nil, false
	}

	glog.V(1).Info("Claiming to support LoadBalancer")

	return &LoadBalancer{network, compute, os.lbOpts}, true
}

func isNotFound(err error) bool {
	e, ok := err.(*gophercloud.UnexpectedResponseCodeError)
	return ok && e.Actual == http.StatusNotFound
}

func getPoolByName(client *gophercloud.ServiceClient, name string) (*pools.Pool, error) {
	opts := pools.ListOpts{
		Name: name,
	}
	pager := pools.List(client, opts)

	poolList := make([]pools.Pool, 0, 1)

	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		p, err := pools.ExtractPools(page)
		if err != nil {
			return false, err
		}
		poolList = append(poolList, p...)
		if len(poolList) > 1 {
			return false, ErrMultipleResults
		}
		return true, nil
	})
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if len(poolList) == 0 {
		return nil, ErrNotFound
	} else if len(poolList) > 1 {
		return nil, ErrMultipleResults
	}

	return &poolList[0], nil
}

func getVipByName(client *gophercloud.ServiceClient, name string) (*vips.VirtualIP, error) {
	opts := vips.ListOpts{
		Name: name,
	}
	pager := vips.List(client, opts)

	vipList := make([]vips.VirtualIP, 0, 1)

	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		v, err := vips.ExtractVIPs(page)
		if err != nil {
			return false, err
		}
		vipList = append(vipList, v...)
		if len(vipList) > 1 {
			return false, ErrMultipleResults
		}
		return true, nil
	})
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if len(vipList) == 0 {
		return nil, ErrNotFound
	} else if len(vipList) > 1 {
		return nil, ErrMultipleResults
	}

	return &vipList[0], nil
}

func getFloatingIPByPortID(client *gophercloud.ServiceClient, portID string) (*floatingips.FloatingIP, error) {
	opts := floatingips.ListOpts{
		PortID: portID,
	}
	pager := floatingips.List(client, opts)

	floatingIPList := make([]floatingips.FloatingIP, 0, 1)

	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		f, err := floatingips.ExtractFloatingIPs(page)
		if err != nil {
			return false, err
		}
		floatingIPList = append(floatingIPList, f...)
		if len(floatingIPList) > 1 {
			return false, ErrMultipleResults
		}
		return true, nil
	})
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if len(floatingIPList) == 0 {
		return nil, ErrNotFound
	} else if len(floatingIPList) > 1 {
		return nil, ErrMultipleResults
	}

	return &floatingIPList[0], nil
}

func (lb *LoadBalancer) GetLoadBalancer(service *api.Service) (*api.LoadBalancerStatus, bool, error) {
	loadBalancerName := cloudprovider.GetLoadBalancerName(service)
	vip, err := getVipByName(lb.network, loadBalancerName)
	if err == ErrNotFound {
		return nil, false, nil
	}
	if vip == nil {
		return nil, false, err
	}

	status := &api.LoadBalancerStatus{}
	status.Ingress = []api.LoadBalancerIngress{{IP: vip.Address}}

	return status, true, err
}

// TODO: This code currently ignores 'region' and always creates a
// loadbalancer in only the current OpenStack region.  We should take
// a list of regions (from config) and query/create loadbalancers in
// each region.

func (lb *LoadBalancer) EnsureLoadBalancer(apiService *api.Service, hosts []string, annotations map[string]string) (*api.LoadBalancerStatus, error) {
	glog.V(4).Infof("EnsureLoadBalancer(%v, %v, %v, %v, %v, %v)", apiService.Namespace, apiService.Name, apiService.Spec.LoadBalancerIP, apiService.Spec.Ports, hosts, annotations)

	ports := apiService.Spec.Ports
	if len(ports) > 1 {
		return nil, fmt.Errorf("multiple ports are not yet supported in openstack load balancers")
	} else if len(ports) == 0 {
		return nil, fmt.Errorf("no ports provided to openstack load balancer")
	}

	// The service controller verified all the protocols match on the ports, just check and use the first one
	// TODO: Convert all error messages to use an event recorder
	if ports[0].Protocol != api.ProtocolTCP {
		return nil, fmt.Errorf("Only TCP LoadBalancer is supported for openstack load balancers")
	}

	affinity := apiService.Spec.SessionAffinity
	var persistence *vips.SessionPersistence
	switch affinity {
	case api.ServiceAffinityNone:
		persistence = nil
	case api.ServiceAffinityClientIP:
		persistence = &vips.SessionPersistence{Type: "SOURCE_IP"}
	default:
		return nil, fmt.Errorf("unsupported load balancer affinity: %v", affinity)
	}

	sourceRanges, err := service.GetLoadBalancerSourceRanges(annotations)
	if err != nil {
		return nil, err
	}

	if !service.IsAllowAll(sourceRanges) {
		return nil, fmt.Errorf("Source range restrictions are not supported for openstack load balancers")
	}

	glog.V(2).Infof("Checking if openstack load balancer already exists: %s", cloudprovider.GetLoadBalancerName(apiService))
	_, exists, err := lb.GetLoadBalancer(apiService)
	if err != nil {
		return nil, fmt.Errorf("error checking if openstack load balancer already exists: %v", err)
	}

	// TODO: Implement a more efficient update strategy for common changes than delete & create
	// In particular, if we implement hosts update, we can get rid of UpdateHosts
	if exists {
		err := lb.EnsureLoadBalancerDeleted(apiService)
		if err != nil {
			return nil, fmt.Errorf("error deleting existing openstack load balancer: %v", err)
		}
	}

	lbmethod := lb.opts.LBMethod
	if lbmethod == "" {
		lbmethod = pools.LBMethodRoundRobin
	}
	name := cloudprovider.GetLoadBalancerName(apiService)
	pool, err := pools.Create(lb.network, pools.CreateOpts{
		Name:     name,
		Protocol: pools.ProtocolTCP,
		SubnetID: lb.opts.SubnetId,
		LBMethod: lbmethod,
	}).Extract()
	if err != nil {
		return nil, err
	}

	for _, host := range hosts {
		addr, err := getAddressByName(lb.compute, host)
		if err != nil {
			return nil, err
		}

		_, err = members.Create(lb.network, members.CreateOpts{
			PoolID:       pool.ID,
			ProtocolPort: ports[0].NodePort, //TODO: need to handle multi-port
			Address:      addr,
		}).Extract()
		if err != nil {
			pools.Delete(lb.network, pool.ID)
			return nil, err
		}
	}

	var mon *monitors.Monitor
	if lb.opts.CreateMonitor {
		mon, err = monitors.Create(lb.network, monitors.CreateOpts{
			Type:       monitors.TypeTCP,
			Delay:      int(lb.opts.MonitorDelay.Duration.Seconds()),
			Timeout:    int(lb.opts.MonitorTimeout.Duration.Seconds()),
			MaxRetries: int(lb.opts.MonitorMaxRetries),
		}).Extract()
		if err != nil {
			pools.Delete(lb.network, pool.ID)
			return nil, err
		}

		_, err = pools.AssociateMonitor(lb.network, pool.ID, mon.ID).Extract()
		if err != nil {
			monitors.Delete(lb.network, mon.ID)
			pools.Delete(lb.network, pool.ID)
			return nil, err
		}
	}

	createOpts := vips.CreateOpts{
		Name:         name,
		Description:  fmt.Sprintf("Kubernetes external service %s", name),
		Protocol:     "TCP",
		ProtocolPort: ports[0].Port, //TODO: need to handle multi-port
		PoolID:       pool.ID,
		SubnetID:     lb.opts.SubnetId,
		Persistence:  persistence,
	}

	loadBalancerIP := apiService.Spec.LoadBalancerIP
	if loadBalancerIP != "" {
		createOpts.Address = loadBalancerIP
	}

	vip, err := vips.Create(lb.network, createOpts).Extract()
	if err != nil {
		if mon != nil {
			monitors.Delete(lb.network, mon.ID)
		}
		pools.Delete(lb.network, pool.ID)
		return nil, err
	}

	status := &api.LoadBalancerStatus{}

	status.Ingress = []api.LoadBalancerIngress{{IP: vip.Address}}

	if lb.opts.FloatingNetworkId != "" {
		floatIPOpts := floatingips.CreateOpts{
			FloatingNetworkID: lb.opts.FloatingNetworkId,
			PortID:            vip.PortID,
		}
		floatIP, err := floatingips.Create(lb.network, floatIPOpts).Extract()
		if err != nil {
			return nil, err
		}

		status.Ingress = append(status.Ingress, api.LoadBalancerIngress{IP: floatIP.FloatingIP})
	}

	return status, nil

}

func (lb *LoadBalancer) UpdateLoadBalancer(service *api.Service, hosts []string) error {
	loadBalancerName := cloudprovider.GetLoadBalancerName(service)
	glog.V(4).Infof("UpdateLoadBalancer(%v, %v)", loadBalancerName, hosts)

	vip, err := getVipByName(lb.network, loadBalancerName)
	if err != nil {
		return err
	}

	// Set of member (addresses) that _should_ exist
	addrs := map[string]bool{}
	for _, host := range hosts {
		addr, err := getAddressByName(lb.compute, host)
		if err != nil {
			return err
		}

		addrs[addr] = true
	}

	// Iterate over members that _do_ exist
	pager := members.List(lb.network, members.ListOpts{PoolID: vip.PoolID})
	err = pager.EachPage(func(page pagination.Page) (bool, error) {
		memList, err := members.ExtractMembers(page)
		if err != nil {
			return false, err
		}

		for _, member := range memList {
			if _, found := addrs[member.Address]; found {
				// Member already exists
				delete(addrs, member.Address)
			} else {
				// Member needs to be deleted
				err = members.Delete(lb.network, member.ID).ExtractErr()
				if err != nil {
					return false, err
				}
			}
		}

		return true, nil
	})
	if err != nil {
		return err
	}

	// Anything left in addrs is a new member that needs to be added
	for addr := range addrs {
		_, err := members.Create(lb.network, members.CreateOpts{
			PoolID:       vip.PoolID,
			Address:      addr,
			ProtocolPort: vip.ProtocolPort,
		}).Extract()
		if err != nil {
			return err
		}
	}

	return nil
}

func (lb *LoadBalancer) EnsureLoadBalancerDeleted(service *api.Service) error {
	loadBalancerName := cloudprovider.GetLoadBalancerName(service)
	glog.V(4).Infof("EnsureLoadBalancerDeleted(%v)", loadBalancerName)

	vip, err := getVipByName(lb.network, loadBalancerName)
	if err != nil && err != ErrNotFound {
		return err
	}

	if lb.opts.FloatingNetworkId != "" && vip != nil {
		floatingIP, err := getFloatingIPByPortID(lb.network, vip.PortID)
		if err != nil && !isNotFound(err) {
			return err
		}
		if floatingIP != nil {
			err = floatingips.Delete(lb.network, floatingIP.ID).ExtractErr()
			if err != nil && !isNotFound(err) {
				return err
			}
		}
	}

	// We have to delete the VIP before the pool can be deleted,
	// so no point continuing if this fails.
	if vip != nil {
		err := vips.Delete(lb.network, vip.ID).ExtractErr()
		if err != nil && !isNotFound(err) {
			return err
		}
	}

	var pool *pools.Pool
	if vip != nil {
		pool, err = pools.Get(lb.network, vip.PoolID).Extract()
		if err != nil && !isNotFound(err) {
			return err
		}
	} else {
		// The VIP is gone, but it is conceivable that a Pool
		// still exists that we failed to delete on some
		// previous occasion.  Make a best effort attempt to
		// cleanup any pools with the same name as the VIP.
		pool, err = getPoolByName(lb.network, service.Name)
		if err != nil && err != ErrNotFound {
			return err
		}
	}

	if pool != nil {
		for _, monId := range pool.MonitorIDs {
			_, err = pools.DisassociateMonitor(lb.network, pool.ID, monId).Extract()
			if err != nil {
				return err
			}

			err = monitors.Delete(lb.network, monId).ExtractErr()
			if err != nil && !isNotFound(err) {
				return err
			}
		}
		err = pools.Delete(lb.network, pool.ID).ExtractErr()
		if err != nil && !isNotFound(err) {
			return err
		}
	}

	return nil
}

func (os *OpenStack) Zones() (cloudprovider.Zones, bool) {
	glog.V(1).Info("Claiming to support Zones")

	return os, true
}
func (os *OpenStack) GetZone() (cloudprovider.Zone, error) {
	glog.V(1).Infof("Current zone is %v", os.region)

	return cloudprovider.Zone{Region: os.region}, nil
}

func (os *OpenStack) Routes() (cloudprovider.Routes, bool) {
	return nil, false
}

// Attaches given cinder volume to the compute running kubelet
func (os *OpenStack) AttachDisk(diskName string) (string, error) {
	disk, err := os.getVolume(diskName)
	if err != nil {
		return "", err
	}
	cClient, err := openstack.NewComputeV2(os.provider, gophercloud.EndpointOpts{
		Region: os.region,
	})
	if err != nil || cClient == nil {
		glog.Errorf("Unable to initialize nova client for region: %s", os.region)
		return "", err
	}

	if len(disk.Attachments) > 0 && disk.Attachments[0]["server_id"] != nil {
		if os.localInstanceID == disk.Attachments[0]["server_id"] {
			glog.V(4).Infof("Disk: %q is already attached to compute: %q", diskName, os.localInstanceID)
			return disk.ID, nil
		} else {
			errMsg := fmt.Sprintf("Disk %q is attached to a different compute: %q, should be detached before proceeding", diskName, disk.Attachments[0]["server_id"])
			glog.Errorf(errMsg)
			return "", errors.New(errMsg)
		}
	}
	// add read only flag here if possible spothanis
	_, err = volumeattach.Create(cClient, os.localInstanceID, &volumeattach.CreateOpts{
		VolumeID: disk.ID,
	}).Extract()
	if err != nil {
		glog.Errorf("Failed to attach %s volume to %s compute", diskName, os.localInstanceID)
		return "", err
	}
	glog.V(2).Infof("Successfully attached %s volume to %s compute", diskName, os.localInstanceID)
	return disk.ID, nil
}

// Detaches given cinder volume from the compute running kubelet
func (os *OpenStack) DetachDisk(partialDiskId string) error {
	disk, err := os.getVolume(partialDiskId)
	if err != nil {
		return err
	}
	cClient, err := openstack.NewComputeV2(os.provider, gophercloud.EndpointOpts{
		Region: os.region,
	})
	if err != nil || cClient == nil {
		glog.Errorf("Unable to initialize nova client for region: %s", os.region)
		return err
	}
	if len(disk.Attachments) > 0 && disk.Attachments[0]["server_id"] != nil && os.localInstanceID == disk.Attachments[0]["server_id"] {
		// This is a blocking call and effects kubelet's performance directly.
		// We should consider kicking it out into a separate routine, if it is bad.
		err = volumeattach.Delete(cClient, os.localInstanceID, disk.ID).ExtractErr()
		if err != nil {
			glog.Errorf("Failed to delete volume %s from compute %s attached %v", disk.ID, os.localInstanceID, err)
			return err
		}
		glog.V(2).Infof("Successfully detached volume: %s from compute: %s", disk.ID, os.localInstanceID)
	} else {
		errMsg := fmt.Sprintf("Disk: %s has no attachments or is not attached to compute: %s", disk.Name, os.localInstanceID)
		glog.Errorf(errMsg)
		return errors.New(errMsg)
	}
	return nil
}

// Takes a partial/full disk id or diskname
func (os *OpenStack) getVolume(diskName string) (volumes.Volume, error) {
	sClient, err := openstack.NewBlockStorageV1(os.provider, gophercloud.EndpointOpts{
		Region: os.region,
	})

	var volume volumes.Volume
	if err != nil || sClient == nil {
		glog.Errorf("Unable to initialize cinder client for region: %s", os.region)
		return volume, err
	}

	err = volumes.List(sClient, nil).EachPage(func(page pagination.Page) (bool, error) {
		vols, err := volumes.ExtractVolumes(page)
		if err != nil {
			glog.Errorf("Failed to extract volumes: %v", err)
			return false, err
		} else {
			for _, v := range vols {
				glog.V(4).Infof("%s %s %v", v.ID, v.Name, v.Attachments)
				if v.Name == diskName || strings.Contains(v.ID, diskName) {
					volume = v
					return true, nil
				}
			}
		}
		// if it reached here then no disk with the given name was found.
		errmsg := fmt.Sprintf("Unable to find disk: %s in region %s", diskName, os.region)
		return false, errors.New(errmsg)
	})
	if err != nil {
		glog.Errorf("Error occured getting volume: %s", diskName)
		return volume, err
	}
	return volume, err
}

// Create a volume of given size (in GiB)
func (os *OpenStack) CreateVolume(name string, size int, tags *map[string]string) (volumeName string, err error) {

	sClient, err := openstack.NewBlockStorageV1(os.provider, gophercloud.EndpointOpts{
		Region: os.region,
	})

	if err != nil || sClient == nil {
		glog.Errorf("Unable to initialize cinder client for region: %s", os.region)
		return "", err
	}

	opts := volumes.CreateOpts{
		Name: name,
		Size: size,
	}
	if tags != nil {
		opts.Metadata = *tags
	}
	vol, err := volumes.Create(sClient, opts).Extract()
	if err != nil {
		glog.Errorf("Failed to create a %d GB volume: %v", size, err)
		return "", err
	}
	glog.Infof("Created volume %v", vol.ID)
	return vol.ID, err
}

func (os *OpenStack) DeleteVolume(volumeName string) error {
	sClient, err := openstack.NewBlockStorageV1(os.provider, gophercloud.EndpointOpts{
		Region: os.region,
	})

	if err != nil || sClient == nil {
		glog.Errorf("Unable to initialize cinder client for region: %s", os.region)
		return err
	}
	err = volumes.Delete(sClient, volumeName).ExtractErr()
	if err != nil {
		glog.Errorf("Cannot delete volume %s: %v", volumeName, err)
	}
	return err
}
