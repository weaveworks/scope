package detailed

import (
	"bytes"
	"net/url"
	"strings"
	"text/template"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"

	"github.com/ugorji/go/codec"
)

// Replacement variable name for the query in the metrics graph url
const urlQueryVarName = ":query"

var (
	// As configured by the user
	metricsGraphURL = ""

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
			ID:    "receive_bytes",
			Label: "Rx/s",
		},
		{
			ID:    "transmit_bytes",
			Label: "Tx/s",
		},
	}

	// Prometheus queries for topologies
	//
	// Metrics
	// - `container_cpu_usage_seconds_total` --> cAdvisor in Kubelets.
	// - `container_memory_usage_bytes` --> cAdvisor in Kubelets.
	topologyQueries = map[string]map[string]*template.Template{
		// Containers

		report.Container: {
			docker.MemoryUsage:   parsedTemplate(`sum(container_memory_usage_bytes{container_name="{{.Label}}"})`),
			docker.CPUTotalUsage: parsedTemplate(`sum(rate(container_cpu_usage_seconds_total{container_name="{{.Label}}"}[1m]))`),
		},
		report.ContainerImage: {
			docker.MemoryUsage:   parsedTemplate(`sum(container_memory_usage_bytes{image="{{.Label}}"})`),
			docker.CPUTotalUsage: parsedTemplate(`sum(rate(container_cpu_usage_seconds_total{image="{{.Label}}"}[1m]))`),
		},
		"group:container:docker_container_hostname": {
			docker.MemoryUsage:   parsedTemplate(`sum(container_memory_usage_bytes{pod_name="{{.Label}}"})`),
			docker.CPUTotalUsage: parsedTemplate(`sum(rate(container_cpu_usage_seconds_total{pod_name="{{.Label}}"}[1m]))`),
		},

		// Kubernetes topologies

		report.Pod: {
			docker.MemoryUsage:   parsedTemplate(`sum(container_memory_usage_bytes{pod_name="{{.Label}}"})`),
			docker.CPUTotalUsage: parsedTemplate(`sum(rate(container_cpu_usage_seconds_total{pod_name="{{.Label}}"}[1m]))`),
			"receive_bytes":      parsedTemplate(`sum(rate(container_network_receive_bytes_total{pod_name="{{.Label}}"}[5m]))`),
			"transmit_bytes":     parsedTemplate(`sum(rate(container_network_transmit_bytes_total{pod_name="{{.Label}}"}[5m]))`),
		},
		// Pod naming:
		// https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#pod-template-hash-label
		"__k8s_controllers": {
			docker.MemoryUsage:   parsedTemplate(`sum(container_memory_usage_bytes{pod_name=~"^{{.Label}}-[^-]+-[^-]+$"})`),
			docker.CPUTotalUsage: parsedTemplate(`sum(rate(container_cpu_usage_seconds_total{pod_name=~"^{{.Label}}-[^-]+-[^-]+$"}[1m]))`),
		},
		report.DaemonSet: {
			docker.MemoryUsage:   parsedTemplate(`sum(container_memory_usage_bytes{pod_name=~"^{{.Label}}-[^-]+$"})`),
			docker.CPUTotalUsage: parsedTemplate(`sum(rate(container_cpu_usage_seconds_total{pod_name=~"^{{.Label}}-[^-]+$"}[1m]))`),
		},
		report.Service: {
			// These recording rules must be defined in the prometheus config.
			// NB: Pods need to be labeled and selected by their respective Service name, meaning:
			// - The Service's `spec.selector` needs to select on `name`
			// - The Service's `metadata.name` needs to be the same value as `spec.selector.name`
			docker.MemoryUsage:   parsedTemplate(`namespace_label_name:container_memory_usage_bytes:sum{label_name="{{.Label}}"}`),
			docker.CPUTotalUsage: parsedTemplate(`namespace_label_name:container_cpu_usage_seconds_total:sum_rate{label_name="{{.Label}}}`),
		},
	}
	k8sControllers = map[string]struct{}{
		report.Deployment: {},
		"stateful_set":    {},
		"cron_job":        {},
	}
)

// SetMetricsGraphURL sets the URL we deduce our eventual metric link from.
// Supports placeholders such as `:orgID` and `:query`. An empty url disables
// this feature. If the `:query` part is missing, a JSON version will be
// appended, see `queryParamsAsJSON()` for more info.
func SetMetricsGraphURL(url string) {
	metricsGraphURL = url
}

// RenderMetricURLs sets respective URLs for metrics in a node summary. Missing metrics
// where we have a query for will be appended as an empty metric (no values or samples).
func RenderMetricURLs(summary NodeSummary, n report.Node) NodeSummary {
	if metricsGraphURL == "" {
		return summary
	}

	queries := getTopologyQueries(n.Topology)
	if len(queries) == 0 {
		return summary
	}

	var maxprio float64
	var bs bytes.Buffer
	var ms []report.MetricRow
	found := make(map[string]struct{})

	// Set URL on existing metrics
	for _, metric := range summary.Metrics {
		if metric.Priority > maxprio {
			maxprio = metric.Priority
		}
		tpl := queries[metric.ID]
		if tpl == nil {
			continue
		}

		bs.Reset()
		if err := tpl.Execute(&bs, summary); err != nil {
			continue
		}

		ms = append(ms, metric)
		ms[len(ms)-1].URL = buildURL(bs.String())
		found[metric.ID] = struct{}{}
	}

	// Append empty metrics for unattached queries
	for _, metadata := range shownQueries {
		if _, ok := found[metadata.ID]; ok {
			continue
		}

		tpl := queries[metadata.ID]
		if tpl == nil {
			continue
		}

		bs.Reset()
		if err := tpl.Execute(&bs, summary); err != nil {
			continue
		}

		maxprio++
		ms = append(ms, report.MetricRow{
			ID:         metadata.ID,
			Label:      metadata.Label,
			URL:        buildURL(bs.String()),
			Metric:     &report.Metric{},
			Priority:   maxprio,
			ValueEmpty: true,
		})
	}

	summary.Metrics = ms

	return summary
}

// buildURL puts together the URL by looking at the configured `metricsGraphURL`.
func buildURL(query string) string {
	if strings.Contains(metricsGraphURL, urlQueryVarName) {
		return strings.Replace(metricsGraphURL, urlQueryVarName, url.QueryEscape(query), -1)
	}

	params, err := queryParamsAsJSON(query)
	if err != nil {
		return ""
	}

	if metricsGraphURL[len(metricsGraphURL)-1] != '/' {
		metricsGraphURL += "/"
	}

	return metricsGraphURL + url.QueryEscape(params)
}

// queryParamsAsJSON packs the query into a JSON of the
// format `{"cells":[{"queries":[$query]}]}`.
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

// parsedTemplate initializes unnamed text templates.
func parsedTemplate(query string) *template.Template {
	tpl, err := template.New("").Parse(query)
	if err != nil {
		panic(err)
	}

	return tpl
}

func getTopologyQueries(t string) map[string]*template.Template {
	if _, ok := k8sControllers[t]; ok {
		t = "__k8s_controllers"
	}
	return topologyQueries[t]
}
