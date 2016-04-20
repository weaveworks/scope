/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package route

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/util/wait"
)

type RouteController struct {
	routes      cloudprovider.Routes
	kubeClient  clientset.Interface
	clusterName string
	clusterCIDR *net.IPNet
}

func New(routes cloudprovider.Routes, kubeClient clientset.Interface, clusterName string, clusterCIDR *net.IPNet) *RouteController {
	return &RouteController{
		routes:      routes,
		kubeClient:  kubeClient,
		clusterName: clusterName,
		clusterCIDR: clusterCIDR,
	}
}

func (rc *RouteController) Run(syncPeriod time.Duration) {
	go wait.Until(func() {
		if err := rc.reconcileNodeRoutes(); err != nil {
			glog.Errorf("Couldn't reconcile node routes: %v", err)
		}
	}, syncPeriod, wait.NeverStop)
}

func (rc *RouteController) reconcileNodeRoutes() error {
	routeList, err := rc.routes.ListRoutes(rc.clusterName)
	if err != nil {
		return fmt.Errorf("error listing routes: %v", err)
	}
	// TODO (cjcullen): use pkg/controller/framework.NewInformer to watch this
	// and reduce the number of lists needed.
	nodeList, err := rc.kubeClient.Core().Nodes().List(api.ListOptions{})
	if err != nil {
		return fmt.Errorf("error listing nodes: %v", err)
	}
	return rc.reconcile(nodeList.Items, routeList)
}

func (rc *RouteController) reconcile(nodes []api.Node, routes []*cloudprovider.Route) error {
	// nodeCIDRs maps nodeName->nodeCIDR
	nodeCIDRs := make(map[string]string)
	// routeMap maps routeTargetInstance->route
	routeMap := make(map[string]*cloudprovider.Route)
	for _, route := range routes {
		routeMap[route.TargetInstance] = route
	}
	wg := sync.WaitGroup{}
	for _, node := range nodes {
		// Skip if the node hasn't been assigned a CIDR yet.
		if node.Spec.PodCIDR == "" {
			continue
		}
		// Check if we have a route for this node w/ the correct CIDR.
		r := routeMap[node.Name]
		if r == nil || r.DestinationCIDR != node.Spec.PodCIDR {
			// If not, create the route.
			route := &cloudprovider.Route{
				TargetInstance:  node.Name,
				DestinationCIDR: node.Spec.PodCIDR,
			}
			nameHint := string(node.UID)
			wg.Add(1)
			glog.Infof("Creating route for node %s %s with hint %s", node.Name, route.DestinationCIDR, nameHint)
			go func(nodeName string, nameHint string, route *cloudprovider.Route, startTime time.Time) {
				if err := rc.routes.CreateRoute(rc.clusterName, nameHint, route); err != nil {
					glog.Errorf("Could not create route %s %s for node %s after %v: %v", nameHint, route.DestinationCIDR, nodeName, time.Now().Sub(startTime), err)
				} else {
					glog.Infof("Created route for node %s %s with hint %s after %v", nodeName, route.DestinationCIDR, nameHint, time.Now().Sub(startTime))
				}
				wg.Done()
			}(node.Name, nameHint, route, time.Now())
		}
		nodeCIDRs[node.Name] = node.Spec.PodCIDR
	}
	for _, route := range routes {
		if rc.isResponsibleForRoute(route) {
			// Check if this route applies to a node we know about & has correct CIDR.
			if nodeCIDRs[route.TargetInstance] != route.DestinationCIDR {
				wg.Add(1)
				// Delete the route.
				glog.Infof("Deleting route %s %s", route.Name, route.DestinationCIDR)
				go func(route *cloudprovider.Route, startTime time.Time) {
					if err := rc.routes.DeleteRoute(rc.clusterName, route); err != nil {
						glog.Errorf("Could not delete route %s %s after %v: %v", route.Name, route.DestinationCIDR, time.Now().Sub(startTime), err)
					} else {
						glog.Infof("Deleted route %s %s after %v", route.Name, route.DestinationCIDR, time.Now().Sub(startTime))
					}
					wg.Done()

				}(route, time.Now())
			}
		}
	}
	wg.Wait()
	return nil
}

func (rc *RouteController) isResponsibleForRoute(route *cloudprovider.Route) bool {
	_, cidr, err := net.ParseCIDR(route.DestinationCIDR)
	if err != nil {
		glog.Errorf("Ignoring route %s, unparsable CIDR: %v", route.Name, err)
		return false
	}
	// Not responsible if this route's CIDR is not within our clusterCIDR
	lastIP := make([]byte, len(cidr.IP))
	for i := range lastIP {
		lastIP[i] = cidr.IP[i] | ^cidr.Mask[i]
	}
	if !rc.clusterCIDR.Contains(cidr.IP) || !rc.clusterCIDR.Contains(lastIP) {
		return false
	}
	return true
}
