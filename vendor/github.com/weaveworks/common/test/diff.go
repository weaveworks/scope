package test

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/pmezard/go-difflib/difflib"
)

// Diff diffs two arbitrary data structures, giving human-readable output.
func Diff(want, have interface{}) string {
	config := spew.NewDefaultConfig()
	config.ContinueOnMethod = true
	config.SortKeys = true
	config.SpewKeys = true
	text, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(config.Sdump(want)),
		B:        difflib.SplitLines(config.Sdump(have)),
		FromFile: "want",
		ToFile:   "have",
		Context:  3,
	})
	return "\n" + text
}
