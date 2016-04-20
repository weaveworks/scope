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
	"os"

	"github.com/spf13/cobra"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/kubectl"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	utilerrors "k8s.io/kubernetes/pkg/util/errors"
)

// ScaleOptions is the start of the data required to perform the operation.  As new fields are added, add them here instead of
// referencing the cmd.Flags()
type ScaleOptions struct {
	Filenames []string
	Recursive bool
}

const (
	scale_long = `Set a new size for a Deployment, ReplicaSet, Replication Controller, or Job.

Scale also allows users to specify one or more preconditions for the scale action.
If --current-replicas or --resource-version is specified, it is validated before the
scale is attempted, and it is guaranteed that the precondition holds true when the
scale is sent to the server.`
	scale_example = `# Scale a replicaset named 'foo' to 3.
kubectl scale --replicas=3 rs/foo

# Scale a resource identified by type and name specified in "foo.yaml" to 3.
kubectl scale --replicas=3 -f foo.yaml

# If the deployment named mysql's current size is 2, scale mysql to 3.
kubectl scale --current-replicas=2 --replicas=3 deployment/mysql

# Scale multiple replication controllers.
kubectl scale --replicas=5 rc/foo rc/bar rc/baz

# Scale job named 'cron' to 3.
kubectl scale --replicas=3 job/cron`
)

// NewCmdScale returns a cobra command with the appropriate configuration and flags to run scale
func NewCmdScale(f *cmdutil.Factory, out io.Writer) *cobra.Command {
	options := &ScaleOptions{}

	cmd := &cobra.Command{
		Use: "scale [--resource-version=version] [--current-replicas=count] --replicas=COUNT (-f FILENAME | TYPE NAME)",
		// resize is deprecated
		Aliases: []string{"resize"},
		Short:   "Set a new size for a Deployment, ReplicaSet, Replication Controller, or Job.",
		Long:    scale_long,
		Example: scale_example,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(cmdutil.ValidateOutputArgs(cmd))
			shortOutput := cmdutil.GetFlagString(cmd, "output") == "name"
			err := RunScale(f, out, cmd, args, shortOutput, options)
			cmdutil.CheckErr(err)
		},
	}
	cmd.Flags().String("resource-version", "", "Precondition for resource version. Requires that the current resource version match this value in order to scale.")
	cmd.Flags().Int("current-replicas", -1, "Precondition for current size. Requires that the current size of the resource match this value in order to scale.")
	cmd.Flags().Int("replicas", -1, "The new desired number of replicas. Required.")
	cmd.MarkFlagRequired("replicas")
	cmd.Flags().Duration("timeout", 0, "The length of time to wait before giving up on a scale operation, zero means don't wait.")
	cmdutil.AddOutputFlagsForMutation(cmd)
	cmdutil.AddRecordFlag(cmd)
	cmdutil.AddInclude3rdPartyFlags(cmd)

	usage := "Filename, directory, or URL to a file identifying the resource to set a new size"
	kubectl.AddJsonFilenameFlag(cmd, &options.Filenames, usage)
	cmdutil.AddRecursiveFlag(cmd, &options.Recursive)
	return cmd
}

// RunScale executes the scaling
func RunScale(f *cmdutil.Factory, out io.Writer, cmd *cobra.Command, args []string, shortOutput bool, options *ScaleOptions) error {
	if len(os.Args) > 1 && os.Args[1] == "resize" {
		printDeprecationWarning("scale", "resize")
	}

	count := cmdutil.GetFlagInt(cmd, "replicas")
	if count < 0 {
		return cmdutil.UsageError(cmd, "--replicas=COUNT is required, and COUNT must be greater than or equal to 0")
	}

	cmdNamespace, enforceNamespace, err := f.DefaultNamespace()
	if err != nil {
		return err
	}

	mapper, typer := f.Object(cmdutil.GetIncludeThirdPartyAPIs(cmd))
	r := resource.NewBuilder(mapper, typer, resource.ClientMapperFunc(f.ClientForMapping), f.Decoder(true)).
		ContinueOnError().
		NamespaceParam(cmdNamespace).DefaultNamespace().
		FilenameParam(enforceNamespace, options.Recursive, options.Filenames...).
		ResourceTypeOrNameArgs(false, args...).
		Flatten().
		Do()
	err = r.Err()
	if err != nil {
		return err
	}

	infos, err := r.Infos()
	if err != nil {
		return err
	}
	info := infos[0]
	mapping := info.ResourceMapping()
	scaler, err := f.Scaler(mapping)
	if err != nil {
		return err
	}

	resourceVersion := cmdutil.GetFlagString(cmd, "resource-version")
	if len(resourceVersion) != 0 && len(infos) > 1 {
		return fmt.Errorf("cannot use --resource-version with multiple resources")
	}
	currentSize := cmdutil.GetFlagInt(cmd, "current-replicas")
	if currentSize != -1 && len(infos) > 1 {
		return fmt.Errorf("cannot use --current-replicas with multiple resources")
	}
	precondition := &kubectl.ScalePrecondition{Size: currentSize, ResourceVersion: resourceVersion}
	retry := kubectl.NewRetryParams(kubectl.Interval, kubectl.Timeout)
	var waitForReplicas *kubectl.RetryParams
	if timeout := cmdutil.GetFlagDuration(cmd, "timeout"); timeout != 0 {
		waitForReplicas = kubectl.NewRetryParams(kubectl.Interval, timeout)
	}

	errs := []error{}
	for _, info := range infos {
		if err := scaler.Scale(info.Namespace, info.Name, uint(count), precondition, retry, waitForReplicas); err != nil {
			errs = append(errs, err)
			continue
		}
		if cmdutil.ShouldRecord(cmd, info) {
			patchBytes, err := cmdutil.ChangeResourcePatch(info, f.Command())
			if err != nil {
				errs = append(errs, err)
				continue
			}
			mapping := info.ResourceMapping()
			client, err := f.ClientForMapping(mapping)
			if err != nil {
				return err
			}
			helper := resource.NewHelper(client, mapping)
			_, err = helper.Patch(info.Namespace, info.Name, api.StrategicMergePatchType, patchBytes)
			if err != nil {
				errs = append(errs, err)
				continue
			}
		}
		cmdutil.PrintSuccess(mapper, shortOutput, out, info.Mapping.Resource, info.Name, "scaled")
	}

	return utilerrors.NewAggregate(errs)
}
