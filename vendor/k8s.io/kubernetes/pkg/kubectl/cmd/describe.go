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

package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	apierrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/kubectl"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/runtime"
	utilerrors "k8s.io/kubernetes/pkg/util/errors"
)

// DescribeOptions is the start of the data required to perform the operation.  As new fields are added, add them here instead of
// referencing the cmd.Flags()
type DescribeOptions struct {
	Filenames []string
	Recursive bool
}

const (
	describe_long = `Show details of a specific resource or group of resources.

This command joins many API calls together to form a detailed description of a
given resource or group of resources.

$ kubectl describe TYPE NAME_PREFIX

will first check for an exact match on TYPE and NAME_PREFIX. If no such resource
exists, it will output details for every resource that has a name prefixed with NAME_PREFIX

` + kubectl.PossibleResourceTypes
	describe_example = `# Describe a node
kubectl describe nodes kubernetes-minion-emt8.c.myproject.internal

# Describe a pod
kubectl describe pods/nginx

# Describe a pod identified by type and name in "pod.json"
kubectl describe -f pod.json

# Describe all pods
kubectl describe pods

# Describe pods by label name=myLabel
kubectl describe po -l name=myLabel

# Describe all pods managed by the 'frontend' replication controller (rc-created pods
# get the name of the rc as a prefix in the pod the name).
kubectl describe pods frontend`
)

func NewCmdDescribe(f *cmdutil.Factory, out io.Writer) *cobra.Command {
	options := &DescribeOptions{}

	cmd := &cobra.Command{
		Use:     "describe (-f FILENAME | TYPE [NAME_PREFIX | -l label] | TYPE/NAME)",
		Short:   "Show details of a specific resource or group of resources",
		Long:    describe_long,
		Example: describe_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunDescribe(f, out, cmd, args, options)
			cmdutil.CheckErr(err)
		},
		ValidArgs: kubectl.DescribableResources(),
	}
	usage := "Filename, directory, or URL to a file containing the resource to describe"
	kubectl.AddJsonFilenameFlag(cmd, &options.Filenames, usage)
	cmdutil.AddRecursiveFlag(cmd, &options.Recursive)
	cmd.Flags().StringP("selector", "l", "", "Selector (label query) to filter on")
	cmdutil.AddInclude3rdPartyFlags(cmd)
	return cmd
}

func RunDescribe(f *cmdutil.Factory, out io.Writer, cmd *cobra.Command, args []string, options *DescribeOptions) error {
	selector := cmdutil.GetFlagString(cmd, "selector")
	cmdNamespace, enforceNamespace, err := f.DefaultNamespace()
	if err != nil {
		return err
	}
	if len(args) == 0 && len(options.Filenames) == 0 {
		fmt.Fprint(out, "You must specify the type of resource to describe. ", valid_resources)
		return cmdutil.UsageError(cmd, "Required resource not specified.")
	}

	mapper, typer := f.Object(cmdutil.GetIncludeThirdPartyAPIs(cmd))
	r := resource.NewBuilder(mapper, typer, resource.ClientMapperFunc(f.ClientForMapping), f.Decoder(true)).
		ContinueOnError().
		NamespaceParam(cmdNamespace).DefaultNamespace().
		FilenameParam(enforceNamespace, options.Recursive, options.Filenames...).
		SelectorParam(selector).
		ResourceTypeOrNameArgs(true, args...).
		Flatten().
		Do()
	err = r.Err()
	if err != nil {
		return err
	}

	allErrs := []error{}
	infos, err := r.Infos()
	if err != nil {
		if apierrors.IsNotFound(err) && len(args) == 2 {
			return DescribeMatchingResources(mapper, typer, f, cmdNamespace, args[0], args[1], out, err)
		}
		allErrs = append(allErrs, err)
	}

	for _, info := range infos {
		mapping := info.ResourceMapping()
		describer, err := f.Describer(mapping)
		if err != nil {
			allErrs = append(allErrs, err)
			continue
		}
		s, err := describer.Describe(info.Namespace, info.Name)
		if err != nil {
			allErrs = append(allErrs, err)
			continue
		}
		fmt.Fprintf(out, "%s\n\n", s)
	}

	return utilerrors.NewAggregate(allErrs)
}

func DescribeMatchingResources(mapper meta.RESTMapper, typer runtime.ObjectTyper, f *cmdutil.Factory, namespace, rsrc, prefix string, out io.Writer, originalError error) error {
	r := resource.NewBuilder(mapper, typer, resource.ClientMapperFunc(f.ClientForMapping), f.Decoder(true)).
		NamespaceParam(namespace).DefaultNamespace().
		ResourceTypeOrNameArgs(true, rsrc).
		SingleResourceType().
		Flatten().
		Do()
	mapping, err := r.ResourceMapping()
	if err != nil {
		return err
	}
	describer, err := f.Describer(mapping)
	if err != nil {
		return err
	}
	infos, err := r.Infos()
	if err != nil {
		return err
	}
	isFound := false
	for ix := range infos {
		info := infos[ix]
		if strings.HasPrefix(info.Name, prefix) {
			isFound = true
			s, err := describer.Describe(info.Namespace, info.Name)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "%s\n", s)
		}
	}
	if !isFound {
		return originalError
	}
	return nil
}
