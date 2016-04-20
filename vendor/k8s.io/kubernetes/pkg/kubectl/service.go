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

package kubectl

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/intstr"
)

// The only difference between ServiceGeneratorV1 and V2 is that the service port is named "default" in V1, while it is left unnamed in V2.
type ServiceGeneratorV1 struct{}

func (ServiceGeneratorV1) ParamNames() []GeneratorParam {
	return paramNames()
}

func (ServiceGeneratorV1) Generate(params map[string]interface{}) (runtime.Object, error) {
	params["port-name"] = "default"
	return generate(params)
}

type ServiceGeneratorV2 struct{}

func (ServiceGeneratorV2) ParamNames() []GeneratorParam {
	return paramNames()
}

func (ServiceGeneratorV2) Generate(params map[string]interface{}) (runtime.Object, error) {
	return generate(params)
}

func paramNames() []GeneratorParam {
	return []GeneratorParam{
		{"default-name", true},
		{"name", false},
		{"selector", true},
		// port will be used if a user specifies --port OR the exposed object
		// has one port
		{"port", false},
		// ports will be used iff a user doesn't specify --port AND the
		// exposed object has multiple ports
		{"ports", false},
		{"labels", false},
		{"external-ip", false},
		{"create-external-load-balancer", false},
		{"load-balancer-ip", false},
		{"type", false},
		{"protocol", false},
		{"container-port", false}, // alias of target-port
		{"target-port", false},
		{"port-name", false},
		{"session-affinity", false},
	}
}

func generate(genericParams map[string]interface{}) (runtime.Object, error) {
	params := map[string]string{}
	for key, value := range genericParams {
		strVal, isString := value.(string)
		if !isString {
			return nil, fmt.Errorf("expected string, saw %v for '%s'", value, key)
		}
		params[key] = strVal
	}
	selectorString, found := params["selector"]
	if !found || len(selectorString) == 0 {
		return nil, fmt.Errorf("'selector' is a required parameter.")
	}
	selector, err := ParseLabels(selectorString)
	if err != nil {
		return nil, err
	}

	labelsString, found := params["labels"]
	var labels map[string]string
	if found && len(labelsString) > 0 {
		labels, err = ParseLabels(labelsString)
		if err != nil {
			return nil, err
		}
	}

	name, found := params["name"]
	if !found || len(name) == 0 {
		name, found = params["default-name"]
		if !found || len(name) == 0 {
			return nil, fmt.Errorf("'name' is a required parameter.")
		}
	}
	ports := []api.ServicePort{}
	servicePortName, found := params["port-name"]
	if !found {
		// Leave the port unnamed.
		servicePortName = ""
	}
	// ports takes precedence over port since it will be
	// specified only when the user hasn't specified a port
	// via --port and the exposed object has multiple ports.
	var portString string
	if portString, found = params["ports"]; !found {
		portString, found = params["port"]
		if !found {
			return nil, fmt.Errorf("'port' is a required parameter.")
		}
	}
	portStringSlice := strings.Split(portString, ",")
	for i, stillPortString := range portStringSlice {
		port, err := strconv.Atoi(stillPortString)
		if err != nil {
			return nil, err
		}
		name := servicePortName
		// If we are going to assign multiple ports to a service, we need to
		// generate a different name for each one.
		if len(portStringSlice) > 1 {
			name = fmt.Sprintf("port-%d", i+1)
		}
		ports = append(ports, api.ServicePort{
			Name:     name,
			Port:     port,
			Protocol: api.Protocol(params["protocol"]),
		})
	}

	service := api.Service{
		ObjectMeta: api.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: api.ServiceSpec{
			Selector: selector,
			Ports:    ports,
		},
	}
	targetPortString, found := params["target-port"]
	if !found {
		targetPortString, found = params["container-port"]
	}
	if found && len(targetPortString) > 0 {
		var targetPort intstr.IntOrString
		if portNum, err := strconv.Atoi(targetPortString); err != nil {
			targetPort = intstr.FromString(targetPortString)
		} else {
			targetPort = intstr.FromInt(portNum)
		}
		// Use the same target-port for every port
		for i := range service.Spec.Ports {
			service.Spec.Ports[i].TargetPort = targetPort
		}
	} else {
		// If --target-port or --container-port haven't been specified, this
		// should be the same as Port
		for i := range service.Spec.Ports {
			port := service.Spec.Ports[i].Port
			service.Spec.Ports[i].TargetPort = intstr.FromInt(port)
		}
	}
	if params["create-external-load-balancer"] == "true" {
		service.Spec.Type = api.ServiceTypeLoadBalancer
	}
	if len(params["external-ip"]) > 0 {
		service.Spec.ExternalIPs = []string{params["external-ip"]}
	}
	if len(params["type"]) != 0 {
		service.Spec.Type = api.ServiceType(params["type"])
	}
	if service.Spec.Type == api.ServiceTypeLoadBalancer {
		service.Spec.LoadBalancerIP = params["load-balancer-ip"]
	}
	if len(params["session-affinity"]) != 0 {
		switch api.ServiceAffinity(params["session-affinity"]) {
		case api.ServiceAffinityNone:
			service.Spec.SessionAffinity = api.ServiceAffinityNone
		case api.ServiceAffinityClientIP:
			service.Spec.SessionAffinity = api.ServiceAffinityClientIP
		default:
			return nil, fmt.Errorf("unknown session affinity: %s", params["session-affinity"])
		}
	}
	return &service, nil
}
