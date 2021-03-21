package detailed

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/report"

	"github.com/ugorji/go/codec"
)

const (
	// Replacement variable name for the query in the metrics graph url
	urlQueryVarName = ":query"

	idReceiveBytes  = "receive_bytes"
	idTransmitBytes = "transmit_bytes"
)

var (
	// Metadata for shown queries
	shownQueries = []struct {
		ID    string
		Label string
	}{
		{
			ID:    docker.CPUTotalUsage,
			Label: "CPU",
		},
		{
			ID:    docker.MemoryUsage,
			Label: "Memory",
		},
		{
			ID:    idReceiveBytes,
			Label: "Rx/s",
		},
		{
			ID:    idTransmitBytes,
			Label: "Tx/s",
		},
	}

	// Queries on pod names of the format `name-<id>-<hash>`
	// See also:  https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#pod-template-hash-label
	podIDHashQueries = formatMetricQueries(`pod=~"^{{label}}-[^-]+-[^-]+$",namespace="{{namespace}}"`, []string{docker.MemoryUsage, docker.CPUTotalUsage})

	// Prometheus queries for topologies
	topologyQueries = map[string]map[string]string{
		// Containers

		report.Container:      formatMetricQueries(`name="{{containerName}}"`, []string{docker.MemoryUsage, docker.CPUTotalUsage}),
		report.ContainerImage: formatMetricQueries(`image="{{label}}"`, []string{docker.MemoryUsage, docker.CPUTotalUsage}),

		// Kubernetes topologies

		report.Pod: formatMetricQueries(
			`pod="{{label}}",namespace="{{namespace}}"`,
			[]string{docker.MemoryUsage, docker.CPUTotalUsage},
		),
		report.DaemonSet:   formatMetricQueries(`pod=~"^{{label}}-[^-]+$",namespace="{{namespace}}"`, []string{docker.MemoryUsage, docker.CPUTotalUsage}),
		report.Deployment:  podIDHashQueries,
		report.StatefulSet: podIDHashQueries,
		report.CronJob:     podIDHashQueries,
		report.Service: {
			docker.CPUTotalUsage: `sum(rate(container_cpu_usage_seconds_total{image!="",namespace="{{namespace}}",_weave_service="{{label}}"}[5m]))`,
			docker.MemoryUsage:   `sum(rate(container_memory_usage_bytes{image!="",namespace="{{namespace}}",_weave_service="{{label}}"}[5m]))`,
		},
	}
)

func formatMetricQueries(filter string, ids []string) map[string]string {
	queries := make(map[string]string)
	for _, id := range ids {
		// All  `container_*`metrics  are provided by cAdvisor in Kubelets
		switch id {
		case docker.MemoryUsage:
			queries[id] = fmt.Sprintf("sum(container_memory_usage_bytes{%s})", filter)
		case docker.CPUTotalUsage:
			queries[id] = fmt.Sprintf(
				"sum(rate(container_cpu_usage_seconds_total{%s}[1m]))*100",
				filter,
			)
		case idReceiveBytes:
			queries[id] = fmt.Sprintf("sum(rate(container_network_receive_bytes_total{%s}[5m]))", filter)
		case idTransmitBytes:
			queries[id] = fmt.Sprintf("sum(rate(container_network_transmit_bytes_total{%s}[5m]))", filter)
		}
	}

	return queries
}

// RenderMetricURLs sets respective URLs for metrics in a node summary. Missing metrics
// where we have a query for will be appended as an empty metric (no values or samples).
func RenderMetricURLs(summary NodeSummary, n report.Node, r report.Report, metricsGraphURL string) NodeSummary {
	if metricsGraphURL == "" {
		return summary
	}

	var maxprio float64
	var ms []report.MetricRow
	found := make(map[string]struct{})

	// Set URL on existing metrics
	for _, metric := range summary.Metrics {
		if metric.Priority > maxprio {
			maxprio = metric.Priority
		}

		query := metricQuery(summary, n, r, metric.ID)
		ms = append(ms, metric)

		if query != "" {
			ms[len(ms)-1].URL = metricURL(query, metricsGraphURL)
		}

		found[metric.ID] = struct{}{}
	}

	// Append empty metrics for unattached queries
	for _, metadata := range shownQueries {
		if _, ok := found[metadata.ID]; ok {
			continue
		}

		query := metricQuery(summary, n, r, metadata.ID)
		if query == "" {
			continue
		}

		maxprio++
		ms = append(ms, report.MetricRow{
			ID:         metadata.ID,
			Label:      metadata.Label,
			URL:        metricURL(query, metricsGraphURL),
			Metric:     &report.Metric{},
			Priority:   maxprio,
			ValueEmpty: true,
		})
	}

	summary.Metrics = ms

	return summary
}

// metricQuery returns the query for the given node and metric.
func metricQuery(summary NodeSummary, n report.Node, r report.Report, metricID string) string {
	queries := topologyQueries[n.Topology]
	if len(queries) == 0 {
		return ""
	}

	namespace, _ := n.Latest.Lookup(kubernetes.Namespace)
	name, _ := n.Latest.Lookup(docker.ContainerName)
	replacer := strings.NewReplacer(
		"{{label}}", metricLabel(summary, n, r),
		"{{namespace}}", namespace,
		"{{containerName}}", name,
	)
	return replacer.Replace(queries[metricID])
}

func metricLabel(summary NodeSummary, n report.Node, r report.Report) string {
	label := summary.Label
	if n.Topology == report.Service {
		deploymentTopology, ok := r.Topology(report.Deployment)
		if ok {
			deploymentNames := []string{}
			for _, pod := range r.Pod.Nodes {
				serviceParents, serviceOk := pod.Parents.Lookup(report.Service)
				deploymentParents, deploymentOk := pod.Parents.Lookup(report.Deployment)
				if serviceOk && deploymentOk && serviceParents.Contains(n.ID) {
					for _, id := range deploymentParents {
						deploymentNode, ok := deploymentTopology.Nodes[id]
						if !ok {
							continue
						}
						if name, ok := deploymentNode.Latest.Lookup(report.KubernetesName); ok {
							deploymentNames = append(deploymentNames, name)
						}
					}
					break
				}
			}
			if len(deploymentNames) == 1 {
				label = deploymentNames[0]
			}
		}
	}
	return label
}

// metricURL builds the URL by embedding it into the configured `metricsGraphURL`.
func metricURL(query, metricsGraphURL string) string {
	if strings.Contains(metricsGraphURL, urlQueryVarName) {
		return strings.Replace(metricsGraphURL, urlQueryVarName, queryEscape(query), -1)
	}

	params, err := queryParamsAsJSON(query)
	if err != nil {
		return ""
	}

	if metricsGraphURL[len(metricsGraphURL)-1] != '/' {
		metricsGraphURL += "/"
	}

	return metricsGraphURL + queryEscape(params)
}

// queryParamsAsJSON packs the query into a JSON of the format `{"cells":[{"queries":[$query]}]}`.
func queryParamsAsJSON(query string) (string, error) {
	type cell struct {
		Queries []string `json:"queries"`
	}
	type queryParams struct {
		Cells []cell `json:"cells"`
	}
	params := &queryParams{[]cell{{[]string{query}}}}

	buf := &bytes.Buffer{}
	encoder := codec.NewEncoder(buf, &codec.JsonHandle{})
	if err := encoder.Encode(params); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// queryEscape uses `%20` instead of `+` to encode whitespaces. Both are
// valid but react-router does not decode `+` properly.
func queryEscape(query string) string {
	return url.QueryEscape(strings.Replace(query, " ", "%20", -1))
}
