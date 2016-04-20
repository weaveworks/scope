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

// A set of common functions needed by cmd/kubectl and pkg/kubectl packages.
package kubectl

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

const (
	kubectlAnnotationPrefix = "kubectl.kubernetes.io/"
	// TODO: auto-generate this
	PossibleResourceTypes = `Possible resource types include (case insensitive): pods (po), services (svc), deployments,
replicasets (rs), replicationcontrollers (rc), nodes (no), events (ev), limitranges (limits),
persistentvolumes (pv), persistentvolumeclaims (pvc), resourcequotas (quota), namespaces (ns),
serviceaccounts (sa), ingresses (ing), horizontalpodautoscalers (hpa), daemonsets (ds), configmaps,
componentstatuses (cs), endpoints (ep), and secrets.`
)

type NamespaceInfo struct {
	Namespace string
}

func listOfImages(spec *api.PodSpec) []string {
	var images []string
	for _, container := range spec.Containers {
		images = append(images, container.Image)
	}
	return images
}

func makeImageList(spec *api.PodSpec) string {
	return strings.Join(listOfImages(spec), ",")
}

func NewThirdPartyResourceMapper(gvs []unversioned.GroupVersion, gvks []unversioned.GroupVersionKind) (meta.RESTMapper, error) {
	mapper := meta.NewDefaultRESTMapper(gvs, func(gv unversioned.GroupVersion) (*meta.VersionInterfaces, error) {
		for ix := range gvs {
			if gvs[ix].Group == gv.Group && gvs[ix].Version == gv.Version {
				return &meta.VersionInterfaces{
					ObjectConvertor:  api.Scheme,
					MetadataAccessor: meta.NewAccessor(),
				}, nil
			}
		}
		groupVersions := []string{}
		for ix := range gvs {
			groupVersions = append(groupVersions, gvs[ix].String())
		}
		return nil, fmt.Errorf("unsupported storage version: %s (valid: %s)", gv.String(), strings.Join(groupVersions, ", "))
	})
	for ix := range gvks {
		mapper.Add(gvks[ix], meta.RESTScopeNamespace)
	}
	return mapper, nil
}

// OutputVersionMapper is a RESTMapper that will prefer mappings that
// correspond to a preferred output version (if feasible)
type OutputVersionMapper struct {
	meta.RESTMapper

	// output versions takes a list of preferred GroupVersions. Only the first
	// hit for a given group will have effect.  This allows different output versions
	// depending upon the group of the kind being requested
	OutputVersions []unversioned.GroupVersion
}

// RESTMapping implements meta.RESTMapper by prepending the output version to the preferred version list.
func (m OutputVersionMapper) RESTMapping(gk unversioned.GroupKind, versions ...string) (*meta.RESTMapping, error) {
	for _, preferredVersion := range m.OutputVersions {
		if gk.Group == preferredVersion.Group {
			mapping, err := m.RESTMapper.RESTMapping(gk, preferredVersion.Version)
			if err == nil {
				return mapping, nil
			}

			break
		}
	}

	return m.RESTMapper.RESTMapping(gk, versions...)
}

// ShortcutExpander is a RESTMapper that can be used for Kubernetes
// resources.  It expands the resource first, then invokes the wrapped RESTMapper
type ShortcutExpander struct {
	RESTMapper meta.RESTMapper
}

var _ meta.RESTMapper = &ShortcutExpander{}

func (e ShortcutExpander) KindFor(resource unversioned.GroupVersionResource) (unversioned.GroupVersionKind, error) {
	return e.RESTMapper.KindFor(expandResourceShortcut(resource))
}

func (e ShortcutExpander) KindsFor(resource unversioned.GroupVersionResource) ([]unversioned.GroupVersionKind, error) {
	return e.RESTMapper.KindsFor(expandResourceShortcut(resource))
}

func (e ShortcutExpander) ResourcesFor(resource unversioned.GroupVersionResource) ([]unversioned.GroupVersionResource, error) {
	return e.RESTMapper.ResourcesFor(expandResourceShortcut(resource))
}

func (e ShortcutExpander) ResourceFor(resource unversioned.GroupVersionResource) (unversioned.GroupVersionResource, error) {
	return e.RESTMapper.ResourceFor(expandResourceShortcut(resource))
}

func (e ShortcutExpander) ResourceSingularizer(resource string) (string, error) {
	return e.RESTMapper.ResourceSingularizer(expandResourceShortcut(unversioned.GroupVersionResource{Resource: resource}).Resource)
}

func (e ShortcutExpander) RESTMapping(gk unversioned.GroupKind, versions ...string) (*meta.RESTMapping, error) {
	return e.RESTMapper.RESTMapping(gk, versions...)
}

func (e ShortcutExpander) AliasesForResource(resource string) ([]string, bool) {
	return e.RESTMapper.AliasesForResource(expandResourceShortcut(unversioned.GroupVersionResource{Resource: resource}).Resource)
}

// shortForms is the list of short names to their expanded names
var shortForms = map[string]string{
	// Please keep this alphabetized
	// If you add an entry here, please also take a look at pkg/kubectl/cmd/cmd.go
	// and add an entry to valid_resources when appropriate.
	"cs":     "componentstatuses",
	"ds":     "daemonsets",
	"ep":     "endpoints",
	"ev":     "events",
	"hpa":    "horizontalpodautoscalers",
	"ing":    "ingresses",
	"limits": "limitranges",
	"no":     "nodes",
	"ns":     "namespaces",
	"po":     "pods",
	"psp":    "podSecurityPolicies",
	"pvc":    "persistentvolumeclaims",
	"pv":     "persistentvolumes",
	"quota":  "resourcequotas",
	"rc":     "replicationcontrollers",
	"rs":     "replicasets",
	"sa":     "serviceaccounts",
	"svc":    "services",
}

// expandResourceShortcut will return the expanded version of resource
// (something that a pkg/api/meta.RESTMapper can understand), if it is
// indeed a shortcut. Otherwise, will return resource unmodified.
func expandResourceShortcut(resource unversioned.GroupVersionResource) unversioned.GroupVersionResource {
	if expanded, ok := shortForms[resource.Resource]; ok {
		// don't change the group or version that's already been specified
		resource.Resource = expanded
	}
	return resource
}

// parseFileSource parses the source given. Acceptable formats include:
//
// 1.  source-path: the basename will become the key name
// 2.  source-name=source-path: the source-name will become the key name and source-path is the path to the key file
//
// Key names cannot include '='.
func parseFileSource(source string) (keyName, filePath string, err error) {
	numSeparators := strings.Count(source, "=")
	switch {
	case numSeparators == 0:
		return path.Base(source), source, nil
	case numSeparators == 1 && strings.HasPrefix(source, "="):
		return "", "", fmt.Errorf("key name for file path %v missing.", strings.TrimPrefix(source, "="))
	case numSeparators == 1 && strings.HasSuffix(source, "="):
		return "", "", fmt.Errorf("file path for key name %v missing.", strings.TrimSuffix(source, "="))
	case numSeparators > 1:
		return "", "", errors.New("Key names or file paths cannot contain '='.")
	default:
		components := strings.Split(source, "=")
		return components[0], components[1], nil
	}
}

// parseLiteralSource parses the source key=val pair
func parseLiteralSource(source string) (keyName, value string, err error) {
	items := strings.Split(source, "=")
	if len(items) != 2 {
		return "", "", fmt.Errorf("invalid literal source %v, expected key=value", source)
	}

	return items[0], items[1], nil
}
