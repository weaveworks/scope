package docker

import (
	"strings"

	"github.com/weaveworks/scope/report"
)

// LabelPrefix is the key prefix used for Docker labels in Node (e.g. a
// Docker label "labelKey"="labelValue" will get encoded as
// "docker_label_labelKey"="dockerValue" in the metadata)
const LabelPrefix = "docker_label_"

// AddLabels appends Docker labels to the Node from a topology.
func AddLabels(node report.Node, labels map[string]string) report.Node {
	node = node.Copy()
	for key, value := range labels {
		node = node.WithLatests(map[string]string{
			LabelPrefix + key: value,
		})
	}
	return node
}

// ExtractLabels returns the list of Docker labels given a Node from a topology.
func ExtractLabels(node report.Node) map[string]string {
	result := map[string]string{}
	node.Latest.ForEach(func(key, value string) {
		if strings.HasPrefix(key, LabelPrefix) {
			label := key[len(LabelPrefix):]
			result[label] = value
		}
	})
	return result
}
