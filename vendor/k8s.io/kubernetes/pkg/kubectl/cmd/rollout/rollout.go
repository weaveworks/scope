/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package rollout

import (
	"io"

	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

const (
	rollout_long    = `Manages a deployment using subcommands like "kubectl rollout undo deployment/abc"`
	rollout_example = `# Rollback to the previous deployment
kubectl rollout undo deployment/abc`
	rollout_valid_resources = `Valid resource types include:
   * deployments
`
)

func NewCmdRollout(f *cmdutil.Factory, out io.Writer) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "rollout SUBCOMMAND",
		Short:   "rollout manages a deployment",
		Long:    rollout_long,
		Example: rollout_example,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	// subcommands
	cmd.AddCommand(NewCmdRolloutHistory(f, out))
	cmd.AddCommand(NewCmdRolloutPause(f, out))
	cmd.AddCommand(NewCmdRolloutResume(f, out))
	cmd.AddCommand(NewCmdRolloutUndo(f, out))

	return cmd
}
