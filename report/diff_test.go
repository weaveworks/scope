package report_test

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/pmezard/go-difflib/difflib"
)

func init() {
	spew.Config.SortKeys = true // :\
}

func diff(want, have interface{}) string {
	text, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(spew.Sdump(want)),
		B:        difflib.SplitLines(spew.Sdump(have)),
		FromFile: "want",
		ToFile:   "have",
		Context:  5,
	})
	return "\n" + text
}
