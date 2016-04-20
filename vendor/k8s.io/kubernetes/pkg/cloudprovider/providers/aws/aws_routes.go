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

package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

func (s *AWSCloud) findRouteTable(clusterName string) (*ec2.RouteTable, error) {
	// This should be unnecessary (we already filter on TagNameKubernetesCluster,
	// and something is broken if cluster name doesn't match, but anyway...
	// TODO: All clouds should be cluster-aware by default
	filters := []*ec2.Filter{newEc2Filter("tag:"+TagNameKubernetesCluster, clusterName)}
	request := &ec2.DescribeRouteTablesInput{Filters: s.addFilters(filters)}

	tables, err := s.ec2.DescribeRouteTables(request)
	if err != nil {
		return nil, err
	}

	if len(tables) == 0 {
		return nil, fmt.Errorf("unable to find route table for AWS cluster: %s", clusterName)
	}

	if len(tables) != 1 {
		return nil, fmt.Errorf("found multiple matching AWS route tables for AWS cluster: %s", clusterName)
	}
	return tables[0], nil
}

// ListRoutes implements Routes.ListRoutes
// List all routes that match the filter
func (s *AWSCloud) ListRoutes(clusterName string) ([]*cloudprovider.Route, error) {
	table, err := s.findRouteTable(clusterName)
	if err != nil {
		return nil, err
	}

	var routes []*cloudprovider.Route
	var instanceIDs []*string

	for _, r := range table.Routes {
		instanceID := orEmpty(r.InstanceId)

		if instanceID == "" {
			continue
		}

		instanceIDs = append(instanceIDs, &instanceID)
	}

	instances, err := s.getInstancesByIDs(instanceIDs)
	if err != nil {
		return nil, err
	}

	for _, r := range table.Routes {
		instanceID := orEmpty(r.InstanceId)
		destinationCIDR := orEmpty(r.DestinationCidrBlock)

		if instanceID == "" || destinationCIDR == "" {
			continue
		}

		instance, found := instances[instanceID]
		if !found {
			glog.Warningf("unable to find instance ID %s in the list of instances being routed to", instanceID)
			continue
		}
		instanceName := orEmpty(instance.PrivateDnsName)
		routeName := clusterName + "-" + destinationCIDR
		routes = append(routes, &cloudprovider.Route{Name: routeName, TargetInstance: instanceName, DestinationCIDR: destinationCIDR})
	}

	return routes, nil
}

// Sets the instance attribute "source-dest-check" to the specified value
func (s *AWSCloud) configureInstanceSourceDestCheck(instanceID string, sourceDestCheck bool) error {
	request := &ec2.ModifyInstanceAttributeInput{}
	request.InstanceId = aws.String(instanceID)
	request.SourceDestCheck = &ec2.AttributeBooleanValue{Value: aws.Bool(sourceDestCheck)}

	_, err := s.ec2.ModifyInstanceAttribute(request)
	if err != nil {
		return fmt.Errorf("error configuring source-dest-check on instance %s: %v", instanceID, err)
	}
	return nil
}

// CreateRoute implements Routes.CreateRoute
// Create the described route
func (s *AWSCloud) CreateRoute(clusterName string, nameHint string, route *cloudprovider.Route) error {
	instance, err := s.getInstanceByNodeName(route.TargetInstance)
	if err != nil {
		return err
	}

	// In addition to configuring the route itself, we also need to configure the instance to accept that traffic
	// On AWS, this requires turning source-dest checks off
	err = s.configureInstanceSourceDestCheck(orEmpty(instance.InstanceId), false)
	if err != nil {
		return err
	}

	table, err := s.findRouteTable(clusterName)
	if err != nil {
		return err
	}

	var deleteRoute *ec2.Route
	for _, r := range table.Routes {
		destinationCIDR := aws.StringValue(r.DestinationCidrBlock)

		if destinationCIDR != route.DestinationCIDR {
			continue
		}

		if aws.StringValue(r.State) == ec2.RouteStateBlackhole {
			deleteRoute = r
		}
	}

	if deleteRoute != nil {
		glog.Infof("deleting blackholed route: %s", aws.StringValue(deleteRoute.DestinationCidrBlock))

		request := &ec2.DeleteRouteInput{}
		request.DestinationCidrBlock = deleteRoute.DestinationCidrBlock
		request.RouteTableId = table.RouteTableId

		_, err = s.ec2.DeleteRoute(request)
		if err != nil {
			return fmt.Errorf("error deleting blackholed AWS route (%s): %v", aws.StringValue(deleteRoute.DestinationCidrBlock), err)
		}
	}

	request := &ec2.CreateRouteInput{}
	// TODO: use ClientToken for idempotency?
	request.DestinationCidrBlock = aws.String(route.DestinationCIDR)
	request.InstanceId = instance.InstanceId
	request.RouteTableId = table.RouteTableId

	_, err = s.ec2.CreateRoute(request)
	if err != nil {
		return fmt.Errorf("error creating AWS route (%s): %v", route.DestinationCIDR, err)
	}

	return nil
}

// DeleteRoute implements Routes.DeleteRoute
// Delete the specified route
func (s *AWSCloud) DeleteRoute(clusterName string, route *cloudprovider.Route) error {
	table, err := s.findRouteTable(clusterName)
	if err != nil {
		return err
	}

	request := &ec2.DeleteRouteInput{}
	request.DestinationCidrBlock = aws.String(route.DestinationCIDR)
	request.RouteTableId = table.RouteTableId

	_, err = s.ec2.DeleteRoute(request)
	if err != nil {
		return fmt.Errorf("error deleting AWS route (%s): %v", route.DestinationCIDR, err)
	}

	return nil
}
