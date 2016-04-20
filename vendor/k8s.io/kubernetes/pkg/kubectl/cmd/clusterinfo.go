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

package cmd

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"k8s.io/kubernetes/pkg/api"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"

	"github.com/daviddengcn/go-colortext"
	"github.com/spf13/cobra"
)

func NewCmdClusterInfo(f *cmdutil.Factory, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use: "cluster-info",
		// clusterinfo is deprecated.
		Aliases: []string{"clusterinfo"},
		Short:   "Display cluster info",
		Long:    "Display addresses of the master and services with label kubernetes.io/cluster-service=true",
		Run: func(cmd *cobra.Command, args []string) {
			err := RunClusterInfo(f, out, cmd)
			cmdutil.CheckErr(err)
		},
	}
	cmdutil.AddInclude3rdPartyFlags(cmd)
	return cmd
}

func RunClusterInfo(f *cmdutil.Factory, out io.Writer, cmd *cobra.Command) error {
	if len(os.Args) > 1 && os.Args[1] == "clusterinfo" {
		printDeprecationWarning("cluster-info", "clusterinfo")
	}

	client, err := f.ClientConfig()
	if err != nil {
		return err
	}
	printService(out, "Kubernetes master", client.Host)

	mapper, typer := f.Object(cmdutil.GetIncludeThirdPartyAPIs(cmd))
	cmdNamespace := cmdutil.GetFlagString(cmd, "namespace")
	if cmdNamespace == "" {
		cmdNamespace = api.NamespaceSystem
	}

	// TODO use generalized labels once they are implemented (#341)
	b := resource.NewBuilder(mapper, typer, resource.ClientMapperFunc(f.ClientForMapping), f.Decoder(true)).
		NamespaceParam(cmdNamespace).DefaultNamespace().
		SelectorParam("kubernetes.io/cluster-service=true").
		ResourceTypeOrNameArgs(false, []string{"services"}...).
		Latest()
	b.Do().Visit(func(r *resource.Info, err error) error {
		if err != nil {
			return err
		}
		services := r.Object.(*api.ServiceList).Items
		for _, service := range services {
			var link string
			if len(service.Status.LoadBalancer.Ingress) > 0 {
				ingress := service.Status.LoadBalancer.Ingress[0]
				ip := ingress.IP
				if ip == "" {
					ip = ingress.Hostname
				}
				for _, port := range service.Spec.Ports {
					link += "http://" + ip + ":" + strconv.Itoa(port.Port) + " "
				}
			} else {
				if len(client.GroupVersion.Group) == 0 {
					link = client.Host + "/api/" + client.GroupVersion.Version + "/proxy/namespaces/" + service.ObjectMeta.Namespace + "/services/" + service.ObjectMeta.Name
				} else {
					link = client.Host + "/api/" + client.GroupVersion.Group + "/" + client.GroupVersion.Version + "/proxy/namespaces/" + service.ObjectMeta.Namespace + "/services/" + service.ObjectMeta.Name

				}
			}
			name := service.ObjectMeta.Labels["kubernetes.io/name"]
			if len(name) == 0 {
				name = service.ObjectMeta.Name
			}
			printService(out, name, link)
		}
		return nil
	})
	return nil

	// TODO consider printing more information about cluster
}

func printService(out io.Writer, name, link string) {
	ct.ChangeColor(ct.Green, false, ct.None, false)
	fmt.Fprint(out, name)
	ct.ResetColor()
	fmt.Fprintf(out, " is running at ")
	ct.ChangeColor(ct.Yellow, false, ct.None, false)
	fmt.Fprint(out, link)
	ct.ResetColor()
	fmt.Fprintln(out, "")
}
