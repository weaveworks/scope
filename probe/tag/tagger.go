package tag

import (
	"log"

	"github.com/weaveworks/scope/report"
)

// Tagger tags nodes with value-add node metadata.
type Tagger interface {
	Tag(r report.Report) (report.Report, error)
}

// Reporter generates Reports.
type Reporter interface {
	Report() (report.Report, error)
}

// Apply tags the report with all the taggers.
func Apply(r report.Report, taggers []Tagger) report.Report {
	var err error
	for _, tagger := range taggers {
		r, err = tagger.Tag(r)
		if err != nil {
			log.Printf("error applying tagger: %v", err)
		}
	}
	return r
}
