package report_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/report"
)

func TestMulticolumnTables(t *testing.T) {
	want := []report.Row{
		{
			ID: "row1",
			Entries: map[string]string{
				"col1": "r1c1",
				"col2": "r1c2",
				"col3": "r1c3",
			},
		},
		{
			ID: "row2",
			Entries: map[string]string{
				"col1": "r2c1",
				"col3": "r2c3",
			},
		},
	}

	nmd := report.MakeNode("foo1")
	nmd = nmd.AddPrefixMulticolumnTable("foo_", want)

	template := report.TableTemplate{
		Type:   report.MulticolumnTableType,
		Prefix: "foo_",
	}

	have, truncationCount := nmd.ExtractTable(template)

	if truncationCount != 0 {
		t.Error("Table shouldn't had been truncated")
	}

	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestPrefixPropertyLists(t *testing.T) {
	want := []report.Row{
		{
			ID: "label_foo1",
			Entries: map[string]string{
				"label": "foo1",
				"value": "bar1",
			},
		},
		{
			ID: "label_foo3",
			Entries: map[string]string{
				"label": "foo3",
				"value": "bar3",
			},
		},
	}

	nmd := report.MakeNode("foo1")
	nmd = nmd.AddPrefixPropertyList("foo_", map[string]string{
		"foo3": "bar3",
		"foo1": "bar1",
	})
	nmd = nmd.AddPrefixPropertyList("zzz_", map[string]string{
		"foo2": "bar2",
	})

	template := report.TableTemplate{
		Type:   report.PropertyListType,
		Prefix: "foo_",
	}

	have, truncationCount := nmd.ExtractTable(template)

	if truncationCount != 0 {
		t.Error("Table shouldn't had been truncated")
	}

	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestFixedPropertyLists(t *testing.T) {
	want := []report.Row{
		{
			ID: "label_foo1",
			Entries: map[string]string{
				"label": "foo1",
				"value": "bar1",
			},
		},
		{
			ID: "label_foo2",
			Entries: map[string]string{
				"label": "foo2",
				"value": "bar2",
			},
		},
	}

	nmd := report.MakeNodeWith("foo1", map[string]string{
		"foo2key": "bar2",
		"foo1key": "bar1",
	})

	template := report.TableTemplate{
		Type: report.PropertyListType,
		FixedRows: map[string]string{
			"foo2key": "foo2",
			"foo1key": "foo1",
		},
	}

	have, truncationCount := nmd.ExtractTable(template)

	if truncationCount != 0 {
		t.Error("Table shouldn't had been truncated")
	}

	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestTables(t *testing.T) {
	want := []report.Table{
		{
			ID:      "AAA",
			Label:   "Aaa",
			Type:    report.PropertyListType,
			Columns: nil,
			Rows: []report.Row{
				{
					ID: "label_foo1",
					Entries: map[string]string{
						"label": "foo1",
						"value": "bar1",
					},
				},
				{
					ID: "label_foo3",
					Entries: map[string]string{
						"label": "foo3",
						"value": "bar3",
					},
				},
			},
		},
		{
			ID:      "BBB",
			Label:   "Bbb",
			Type:    report.MulticolumnTableType,
			Columns: []report.Column{{ID: "col1", Label: "Column 1"}},
			Rows: []report.Row{
				{
					ID: "row1",
					Entries: map[string]string{
						"col1": "r1c1",
					},
				},
				{
					ID: "row2",
					Entries: map[string]string{
						"col3": "r2c3",
					},
				},
			},
		},
		{
			ID:      "CCC",
			Label:   "Ccc",
			Type:    report.PropertyListType,
			Columns: nil,
			Rows: []report.Row{
				{
					ID: "label_foo3",
					Entries: map[string]string{
						"label": "foo3",
						"value": "bar3",
					},
				},
			},
		},
	}

	nmd := report.MakeNodeWith("foo1", map[string]string{
		"foo3key": "bar3",
		"foo1key": "bar1",
	})
	nmd = nmd.AddPrefixMulticolumnTable("bbb_", []report.Row{
		{ID: "row1", Entries: map[string]string{"col1": "r1c1"}},
		{ID: "row2", Entries: map[string]string{"col3": "r2c3"}},
	})
	nmd = nmd.AddPrefixPropertyList("aaa_", map[string]string{
		"foo3": "bar3",
		"foo1": "bar1",
	})

	aaaTemplate := report.TableTemplate{
		ID:     "AAA",
		Label:  "Aaa",
		Prefix: "aaa_",
		Type:   report.PropertyListType,
	}
	bbbTemplate := report.TableTemplate{
		ID:      "BBB",
		Label:   "Bbb",
		Prefix:  "bbb_",
		Type:    report.MulticolumnTableType,
		Columns: []report.Column{{ID: "col1", Label: "Column 1"}},
	}
	cccTemplate := report.TableTemplate{
		ID:        "CCC",
		Label:     "Ccc",
		Prefix:    "ccc_",
		Type:      report.PropertyListType,
		FixedRows: map[string]string{"foo3key": "foo3"},
	}
	templates := report.TableTemplates{
		aaaTemplate.ID: aaaTemplate,
		bbbTemplate.ID: bbbTemplate,
		cccTemplate.ID: cccTemplate,
	}

	have := templates.Tables(nmd)

	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
