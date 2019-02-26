package detailed_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/render/detailed"
	"github.com/weaveworks/scope/report"
)

func TestCensorNode(t *testing.T) {
	node := detailed.Node{
		NodeSummary: detailed.NodeSummary{
			Metadata: []report.MetadataRow{
				{ID: "cmdline", Label: "Command", Value: "prog -a --b=c"},
			},
			Tables: []report.Table{
				{ID: "blibli", Rows: []report.Row{{ID: "bli"}}},
				{ID: "docker_env_", Rows: []report.Row{{ID: "env_var"}}},
			},
		},
	}

	for _, c := range []struct {
		label      string
		have, want detailed.Node
	}{
		{
			label: "no censoring",
			have: detailed.CensorNode(node, report.CensorConfig{
				HideCommandLineArguments: false,
				HideEnvironmentVariables: false,
			}),
			want: detailed.Node{
				NodeSummary: detailed.NodeSummary{
					Metadata: []report.MetadataRow{
						{ID: "cmdline", Label: "Command", Value: "prog -a --b=c"},
					},
					Tables: []report.Table{
						{ID: "blibli", Rows: []report.Row{{ID: "bli"}}},
						{ID: "docker_env_", Rows: []report.Row{{ID: "env_var"}}},
					},
				},
			},
		},
		{
			label: "censor only command line args",
			have: detailed.CensorNode(node, report.CensorConfig{
				HideCommandLineArguments: true,
				HideEnvironmentVariables: false,
			}),
			want: detailed.Node{
				NodeSummary: detailed.NodeSummary{
					Metadata: []report.MetadataRow{
						{ID: "cmdline", Label: "Command", Value: "prog"},
					},
					Tables: []report.Table{
						{ID: "blibli", Rows: []report.Row{{ID: "bli"}}},
						{ID: "docker_env_", Rows: []report.Row{{ID: "env_var"}}},
					},
				},
			},
		},
		{
			label: "censor only env variables",
			have: detailed.CensorNode(node, report.CensorConfig{
				HideCommandLineArguments: false,
				HideEnvironmentVariables: true,
			}),
			want: detailed.Node{
				NodeSummary: detailed.NodeSummary{
					Metadata: []report.MetadataRow{
						{ID: "cmdline", Label: "Command", Value: "prog -a --b=c"},
					},
					Tables: []report.Table{
						{ID: "blibli", Rows: []report.Row{{ID: "bli"}}},
					},
				},
			},
		},
		{
			label: "censor both command line args and env vars",
			have: detailed.CensorNode(node, report.CensorConfig{
				HideCommandLineArguments: true,
				HideEnvironmentVariables: true,
			}),
			want: detailed.Node{
				NodeSummary: detailed.NodeSummary{
					Metadata: []report.MetadataRow{
						{ID: "cmdline", Label: "Command", Value: "prog"},
					},
					Tables: []report.Table{
						{ID: "blibli", Rows: []report.Row{{ID: "bli"}}},
					},
				},
			},
		},
	} {
		if !reflect.DeepEqual(c.want, c.have) {
			t.Errorf("%s - %s", c.label, test.Diff(c.want, c.have))
		}
	}
}

func TestCensorNodeSummaries(t *testing.T) {
	summaries := detailed.NodeSummaries{
		"a": detailed.NodeSummary{
			Metadata: []report.MetadataRow{
				{ID: "blublu", Label: "blabla", Value: "blu blu"},
				{ID: "docker_container_command", Label: "Command", Value: "scope --token=blibli"},
			},
		},
		"b": detailed.NodeSummary{
			Metadata: []report.MetadataRow{
				{ID: "cmdline", Label: "Command", Value: "prog -a --b=c"},
			},
			Tables: []report.Table{
				{ID: "blibli", Rows: []report.Row{{ID: "bli"}}},
				{ID: "docker_env_", Rows: []report.Row{{ID: "env_var"}}},
			},
		},
	}

	for _, c := range []struct {
		label      string
		have, want detailed.NodeSummaries
	}{
		{
			label: "no censoring",
			have: detailed.CensorNodeSummaries(summaries, report.CensorConfig{
				HideCommandLineArguments: false,
				HideEnvironmentVariables: false,
			}),
			want: detailed.NodeSummaries{
				"a": detailed.NodeSummary{
					Metadata: []report.MetadataRow{
						{ID: "blublu", Label: "blabla", Value: "blu blu"},
						{ID: "docker_container_command", Label: "Command", Value: "scope --token=blibli"},
					},
				},
				"b": detailed.NodeSummary{
					Metadata: []report.MetadataRow{
						{ID: "cmdline", Label: "Command", Value: "prog -a --b=c"},
					},
					Tables: []report.Table{
						{ID: "blibli", Rows: []report.Row{{ID: "bli"}}},
						{ID: "docker_env_", Rows: []report.Row{{ID: "env_var"}}},
					},
				},
			},
		},
		{
			label: "censor only command line args",
			have: detailed.CensorNodeSummaries(summaries, report.CensorConfig{
				HideCommandLineArguments: true,
				HideEnvironmentVariables: false,
			}),
			want: detailed.NodeSummaries{
				"a": detailed.NodeSummary{
					Metadata: []report.MetadataRow{
						{ID: "blublu", Label: "blabla", Value: "blu blu"},
						{ID: "docker_container_command", Label: "Command", Value: "scope"},
					},
				},
				"b": detailed.NodeSummary{
					Metadata: []report.MetadataRow{
						{ID: "cmdline", Label: "Command", Value: "prog"},
					},
					Tables: []report.Table{
						{ID: "blibli", Rows: []report.Row{{ID: "bli"}}},
						{ID: "docker_env_", Rows: []report.Row{{ID: "env_var"}}},
					},
				},
			},
		},
		{
			label: "censor only env variables",
			have: detailed.CensorNodeSummaries(summaries, report.CensorConfig{
				HideCommandLineArguments: false,
				HideEnvironmentVariables: true,
			}),
			want: detailed.NodeSummaries{
				"a": detailed.NodeSummary{
					Metadata: []report.MetadataRow{
						{ID: "blublu", Label: "blabla", Value: "blu blu"},
						{ID: "docker_container_command", Label: "Command", Value: "scope --token=blibli"},
					},
				},
				"b": detailed.NodeSummary{
					Metadata: []report.MetadataRow{
						{ID: "cmdline", Label: "Command", Value: "prog -a --b=c"},
					},
					Tables: []report.Table{
						{ID: "blibli", Rows: []report.Row{{ID: "bli"}}},
					},
				},
			},
		},
		{
			label: "censor both command line args and env vars",
			have: detailed.CensorNodeSummaries(summaries, report.CensorConfig{
				HideCommandLineArguments: true,
				HideEnvironmentVariables: true,
			}),
			want: detailed.NodeSummaries{
				"a": detailed.NodeSummary{
					Metadata: []report.MetadataRow{
						{ID: "blublu", Label: "blabla", Value: "blu blu"},
						{ID: "docker_container_command", Label: "Command", Value: "scope"},
					},
				},
				"b": detailed.NodeSummary{
					Metadata: []report.MetadataRow{
						{ID: "cmdline", Label: "Command", Value: "prog"},
					},
					Tables: []report.Table{
						{ID: "blibli", Rows: []report.Row{{ID: "bli"}}},
					},
				},
			},
		},
	} {
		if !reflect.DeepEqual(c.want, c.have) {
			t.Errorf("%s - %s", c.label, test.Diff(c.want, c.have))
		}
	}
}
