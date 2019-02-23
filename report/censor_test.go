package report_test

import (
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
)

func TestCensorRawReport(t *testing.T) {
	r := report.Report{
		Container: report.Topology{
			Nodes: report.Nodes{
				"a": report.MakeNodeWith("a", map[string]string{
					"docker_container_command": "prog -a --b=c",
					"blublu":                   "blu blu",
					"docker_env_":              "env_var",
				}),
			},
		},
		Process: report.Topology{
			Nodes: report.Nodes{
				"b": report.MakeNodeWith("b", map[string]string{
					"cmdline": "scope --token=blibli",
					"blibli":  "bli bli",
				}),
				"c": report.MakeNodeWith("c", map[string]string{
					"docker_env_": "var",
				}),
			},
		},
	}

	for _, c := range []struct {
		label      string
		have, want report.Report
	}{
		{
			label: "no censoring",
			have: report.CensorRawReport(r, report.CensorConfig{
				HideCommandLineArguments: false,
				HideEnvironmentVariables: false,
			}),
			want: report.Report{
				Container: report.Topology{
					Nodes: report.Nodes{
						"a": report.MakeNodeWith("a", map[string]string{
							"docker_container_command": "prog -a --b=c",
							"blublu":                   "blu blu",
							"docker_env_":              "env_var",
						}),
					},
				},
				Process: report.Topology{
					Nodes: report.Nodes{
						"b": report.MakeNodeWith("b", map[string]string{
							"cmdline": "scope --token=blibli",
							"blibli":  "bli bli",
						}),
						"c": report.MakeNodeWith("c", map[string]string{
							"docker_env_": "var",
						}),
					},
				},
			},
		},
		// {
		// 	label: "censor only command line args",
		// 	have: report.CensorRawReport(r, report.CensorConfig{
		// 		HideCommandLineArguments: true,
		// 		HideEnvironmentVariables: false,
		// 	}),
		// 	want: report.Report{
		// 		Container: report.Topology{
		// 			Nodes: report.Nodes{
		// 				"a": report.MakeNodeWith("a", map[string]string{
		// 					"docker_container_command": "prog",
		// 					"blublu":                   "blu blu",
		// 					"docker_env_":              "env_var",
		// 				}),
		// 			},
		// 		},
		// 		Process: report.Topology{
		// 			Nodes: report.Nodes{
		// 				"b": report.MakeNodeWith("b", map[string]string{
		// 					"cmdline": "scope",
		// 					"blibli":  "bli bli",
		// 				}),
		// 				"c": report.MakeNodeWith("c", map[string]string{
		// 					"docker_env_": "var",
		// 				}),
		// 			},
		// 		},
		// 	},
		// },
	} {
		if !reflect.DeepEqual(c.want, c.have) {
			t.Errorf("%s - %s", c.label, test.Diff(c.want, c.have))
		}
	}
}
