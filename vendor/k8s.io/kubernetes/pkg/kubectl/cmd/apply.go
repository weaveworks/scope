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

	"github.com/spf13/cobra"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/kubectl"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/strategicpatch"
)

// ApplyOptions stores cmd.Flag values for apply.  As new fields are added,
// add them here instead of referencing the cmd.Flags()
type ApplyOptions struct {
	Filenames []string
	Recursive bool
}

const (
	apply_long = `Apply a configuration to a resource by filename or stdin.
The resource will be created if it doesn't exist yet. 

JSON and YAML formats are accepted.`
	apply_example = `# Apply the configuration in pod.json to a pod.
kubectl apply -f ./pod.json

# Apply the JSON passed into stdin to a pod.
cat pod.json | kubectl apply -f -`
)

func NewCmdApply(f *cmdutil.Factory, out io.Writer) *cobra.Command {
	options := &ApplyOptions{}

	cmd := &cobra.Command{
		Use:     "apply -f FILENAME",
		Short:   "Apply a configuration to a resource by filename or stdin",
		Long:    apply_long,
		Example: apply_example,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(validateArgs(cmd, args))
			cmdutil.CheckErr(cmdutil.ValidateOutputArgs(cmd))
			cmdutil.CheckErr(RunApply(f, cmd, out, options))
		},
	}

	usage := "Filename, directory, or URL to file that contains the configuration to apply"
	kubectl.AddJsonFilenameFlag(cmd, &options.Filenames, usage)
	cmd.MarkFlagRequired("filename")
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddRecursiveFlag(cmd, &options.Recursive)
	cmdutil.AddOutputFlagsForMutation(cmd)
	cmdutil.AddRecordFlag(cmd)
	cmdutil.AddInclude3rdPartyFlags(cmd)
	return cmd
}

func validateArgs(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmdutil.UsageError(cmd, "Unexpected args: %v", args)
	}

	return nil
}

func RunApply(f *cmdutil.Factory, cmd *cobra.Command, out io.Writer, options *ApplyOptions) error {
	shortOutput := cmdutil.GetFlagString(cmd, "output") == "name"
	schema, err := f.Validator(cmdutil.GetFlagBool(cmd, "validate"), cmdutil.GetFlagString(cmd, "schema-cache-dir"))
	if err != nil {
		return err
	}

	cmdNamespace, enforceNamespace, err := f.DefaultNamespace()
	if err != nil {
		return err
	}

	mapper, typer := f.Object(cmdutil.GetIncludeThirdPartyAPIs(cmd))
	r := resource.NewBuilder(mapper, typer, resource.ClientMapperFunc(f.ClientForMapping), f.Decoder(true)).
		Schema(schema).
		ContinueOnError().
		NamespaceParam(cmdNamespace).DefaultNamespace().
		FilenameParam(enforceNamespace, options.Recursive, options.Filenames...).
		Flatten().
		Do()
	err = r.Err()
	if err != nil {
		return err
	}

	encoder := f.JSONEncoder()
	decoder := f.Decoder(false)

	count := 0
	err = r.Visit(func(info *resource.Info, err error) error {
		// In this method, info.Object contains the object retrieved from the server
		// and info.VersionedObject contains the object decoded from the input source.
		if err != nil {
			return err
		}

		// Get the modified configuration of the object. Embed the result
		// as an annotation in the modified configuration, so that it will appear
		// in the patch sent to the server.
		modified, err := kubectl.GetModifiedConfiguration(info, true, encoder)
		if err != nil {
			return cmdutil.AddSourceToErr(fmt.Sprintf("retrieving modified configuration from:\n%v\nfor:", info), info.Source, err)
		}

		if err := info.Get(); err != nil {
			if !errors.IsNotFound(err) {
				return cmdutil.AddSourceToErr(fmt.Sprintf("retrieving current configuration of:\n%v\nfrom server for:", info), info.Source, err)
			}
			// Create the resource if it doesn't exist
			// First, update the annotation used by kubectl apply
			if err := kubectl.CreateApplyAnnotation(info, encoder); err != nil {
				return cmdutil.AddSourceToErr("creating", info.Source, err)
			}

			if cmdutil.ShouldRecord(cmd, info) {
				if err := cmdutil.RecordChangeCause(info.Object, f.Command()); err != nil {
					return cmdutil.AddSourceToErr("creating", info.Source, err)
				}
			}

			// Then create the resource and skip the three-way merge
			if err := createAndRefresh(info); err != nil {
				return cmdutil.AddSourceToErr("creating", info.Source, err)
			}
			count++
			cmdutil.PrintSuccess(mapper, shortOutput, out, info.Mapping.Resource, info.Name, "created")
			return nil
		}

		// Serialize the current configuration of the object from the server.
		current, err := runtime.Encode(encoder, info.Object)
		if err != nil {
			return cmdutil.AddSourceToErr(fmt.Sprintf("serializing current configuration from:\n%v\nfor:", info), info.Source, err)
		}

		// Retrieve the original configuration of the object from the annotation.
		original, err := kubectl.GetOriginalConfiguration(info)
		if err != nil {
			return cmdutil.AddSourceToErr(fmt.Sprintf("retrieving original configuration from:\n%v\nfor:", info), info.Source, err)
		}

		// Create the versioned struct from the original from the server for
		// strategic patch.
		// TODO: Move all structs in apply to use raw data. Can be done once
		// builder has a RawResult method which delivers raw data instead of
		// internal objects.
		versionedObject, _, err := decoder.Decode(current, nil, nil)
		if err != nil {
			return cmdutil.AddSourceToErr(fmt.Sprintf("converting encoded server-side object back to versioned struct:\n%v\nfor:", info), info.Source, err)
		}

		// Compute a three way strategic merge patch to send to server.
		patch, err := strategicpatch.CreateThreeWayMergePatch(original, modified, current, versionedObject, true)
		if err != nil {
			format := "creating patch with:\noriginal:\n%s\nmodified:\n%s\ncurrent:\n%s\nfrom:\n%v\nfor:"
			return cmdutil.AddSourceToErr(fmt.Sprintf(format, original, modified, current, info), info.Source, err)
		}

		helper := resource.NewHelper(info.Client, info.Mapping)
		_, err = helper.Patch(info.Namespace, info.Name, api.StrategicMergePatchType, patch)
		if err != nil {
			return cmdutil.AddSourceToErr(fmt.Sprintf("applying patch:\n%s\nto:\n%v\nfor:", patch, info), info.Source, err)
		}

		if cmdutil.ShouldRecord(cmd, info) {
			patch, err = cmdutil.ChangeResourcePatch(info, f.Command())
			if err != nil {
				return err
			}
			_, err = helper.Patch(info.Namespace, info.Name, api.StrategicMergePatchType, patch)
			if err != nil {
				return cmdutil.AddSourceToErr(fmt.Sprintf("applying patch:\n%s\nto:\n%v\nfor:", patch, info), info.Source, err)
			}
		}

		count++
		cmdutil.PrintSuccess(mapper, shortOutput, out, info.Mapping.Resource, info.Name, "configured")
		return nil
	})

	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("no objects passed to apply")
	}

	return nil
}
