package docker

import (
	"strings"

	"github.com/weaveworks/scope/report"
)

// LabelPrefix is the key prefix used for Docker labels in NodeMetadata (e.g. a
// Docker label "labelKey"="labelValue" will get encoded as
// "docker_label_labelKey"="dockerValue" in the metadata)
const LabelPrefix = "docker_label_"

// AddLabels appends Docker labels to the NodeMetadata from a topology.
func AddLabels(nmd report.NodeMetadata, labels map[string]string) {
	for key, value := range labels {
		nmd.Metadata[LabelPrefix+key] = value
	}
}

// ExtractLabels returns the list of Docker labels given a NodeMetadata from a topology.
func ExtractLabels(nmd report.NodeMetadata) map[string]string {
	result := map[string]string{}
	for key, value := range nmd.Metadata {
		if strings.HasPrefix(key, LabelPrefix) {
			label := key[len(LabelPrefix):]
			result[label] = value
		}
	}
	return result
}
