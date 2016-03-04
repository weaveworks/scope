package plugins

import (
	log "github.com/Sirupsen/logrus"

	"github.com/weaveworks/scope/probe"
	"github.com/weaveworks/scope/report"
)

// Reporter implements the Reporter interface for a plugin registry.
func Reporter(pluginRegistry *Registry) probe.Reporter {
	return probe.ReporterFunc("plugins", func() (report.Report, error) {
		rpt := report.MakeReport()
		pluginRegistry.Implementers("reporter", func(plugin *Plugin) {
			pluginReport, err := plugin.Report()
			if err != nil {
				log.Errorf("plugins: error getting report from %s: %v", plugin.ID, err)
				return
			}
			pluginReport.Plugins = pluginReport.Plugins.Add(plugin.PluginSpec)
			rpt = rpt.Merge(pluginReport)
		})
		return rpt, nil
	})
}
