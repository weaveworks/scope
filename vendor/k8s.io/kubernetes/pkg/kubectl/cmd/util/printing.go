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

package util

import (
	"fmt"
	"io"
	"strings"

	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/kubectl"

	"github.com/spf13/cobra"
)

// AddPrinterFlags adds printing related flags to a command (e.g. output format, no headers, template path)
func AddPrinterFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "", "Output format. One of: json|yaml|wide|name|go-template=...|go-template-file=...|jsonpath=...|jsonpath-file=... See golang template [http://golang.org/pkg/text/template/#pkg-overview] and jsonpath template [http://releases.k8s.io/HEAD/docs/user-guide/jsonpath.md].")
	cmd.Flags().String("output-version", "", "Output the formatted object with the given group version (for ex: 'extensions/v1beta1').")
	cmd.Flags().Bool("no-headers", false, "When using the default output, don't print headers.")
	cmd.Flags().Bool("show-labels", false, "When printing, show all labels as the last column (default hide labels column)")
	// template shorthand -t is deprecated to support -t for --tty
	// TODO: remove template flag shorthand -t
	cmd.Flags().StringP("template", "t", "", "Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview].")
	cmd.MarkFlagFilename("template")
	cmd.Flags().MarkShorthandDeprecated("template", "please use --template instead")
	cmd.Flags().String("sort-by", "", "If non-empty, sort list types using this field specification.  The field specification is expressed as a JSONPath expression (e.g. '{.metadata.name}'). The field in the API resource specified by this JSONPath expression must be an integer or a string.")
	cmd.Flags().BoolP("show-all", "a", false, "When printing, show all resources (default hide terminated pods.)")
}

// AddOutputFlagsForMutation adds output related flags to a command. Used by mutations only.
func AddOutputFlagsForMutation(cmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "", "Output mode. Use \"-o name\" for shorter output (resource/name).")
}

// PrintSuccess prints message after finishing mutating operations
func PrintSuccess(mapper meta.RESTMapper, shortOutput bool, out io.Writer, resource string, name string, operation string) {
	resource, _ = mapper.ResourceSingularizer(resource)
	if shortOutput {
		// -o name: prints resource/name
		if len(resource) > 0 {
			fmt.Fprintf(out, "%s/%s\n", resource, name)
		} else {
			fmt.Fprintf(out, "%s\n", name)
		}
	} else {
		// understandable output by default
		if len(resource) > 0 {
			fmt.Fprintf(out, "%s \"%s\" %s\n", resource, name, operation)
		} else {
			fmt.Fprintf(out, "\"%s\" %s\n", name, operation)
		}
	}
}

// ValidateOutputArgs validates -o flag args for mutations
func ValidateOutputArgs(cmd *cobra.Command) error {
	outputMode := GetFlagString(cmd, "output")
	if outputMode != "" && outputMode != "name" {
		return UsageError(cmd, "Unexpected -o output mode: %v. We only support '-o name'.", outputMode)
	}
	return nil
}

// OutputVersion returns the preferred output version for generic content (JSON, YAML, or templates)
// defaultVersion is never mutated.  Nil simply allows clean passing in common usage from client.Config
func OutputVersion(cmd *cobra.Command, defaultVersion *unversioned.GroupVersion) (unversioned.GroupVersion, error) {
	outputVersionString := GetFlagString(cmd, "output-version")
	if len(outputVersionString) == 0 {
		if defaultVersion == nil {
			return unversioned.GroupVersion{}, nil
		}

		return *defaultVersion, nil
	}

	return unversioned.ParseGroupVersion(outputVersionString)
}

// PrinterForCommand returns the default printer for this command.
// Requires that printer flags have been added to cmd (see AddPrinterFlags).
func PrinterForCommand(cmd *cobra.Command) (kubectl.ResourcePrinter, bool, error) {
	outputFormat := GetFlagString(cmd, "output")

	// templates are logically optional for specifying a format.
	// TODO once https://github.com/kubernetes/kubernetes/issues/12668 is fixed, this should fall back to GetFlagString
	templateFile, _ := cmd.Flags().GetString("template")
	if len(outputFormat) == 0 && len(templateFile) != 0 {
		outputFormat = "template"
	}

	templateFormat := []string{
		"go-template=", "go-template-file=", "jsonpath=", "jsonpath-file=", "custom-columns=", "custom-columns-file=",
	}
	for _, format := range templateFormat {
		if strings.HasPrefix(outputFormat, format) {
			templateFile = outputFormat[len(format):]
			outputFormat = format[:len(format)-1]
		}
	}

	printer, generic, err := kubectl.GetPrinter(outputFormat, templateFile)
	if err != nil {
		return nil, generic, err
	}

	return maybeWrapSortingPrinter(cmd, printer), generic, nil
}

func maybeWrapSortingPrinter(cmd *cobra.Command, printer kubectl.ResourcePrinter) kubectl.ResourcePrinter {
	sorting, err := cmd.Flags().GetString("sort-by")
	if err != nil {
		// error can happen on missing flag or bad flag type.  In either case, this command didn't intent to sort
		return printer
	}

	if len(sorting) != 0 {
		return &kubectl.SortingPrinter{
			Delegate:  printer,
			SortField: fmt.Sprintf("{%s}", sorting),
		}
	}
	return printer
}
