package tag

import "github.com/weaveworks/scope/report"

// Tagger tags nodes with value-add node metadata.
type Tagger interface {
	Tag(r report.Report) report.Report
}

// Apply tags the report with all the taggers.
func Apply(r report.Report, taggers []Tagger) report.Report {
	for _, tagger := range taggers {
		r = tagger.Tag(r)
	}
	return r
}
