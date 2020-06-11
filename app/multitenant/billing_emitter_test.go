package multitenant

import "testing"

func Test_intervalFromCommand(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want string
	}{
		{cmd: "/home/weave/scope --mode=probe --probe-only --probe.kubernetes=true --probe.spy.interval=3s --probe.publish.interval=5s --probe.processes=false --probe.conntrack=false --probe.ebpf.connections=false --probe.docker.bridge=docker0 --probe.docker=true https://redacted@cloud.weave.works.", want: "5s", name: "seconds"},
		{cmd: "/home/weave/scope --mode=probe --probe-only --probe.kubernetes=true --probe.spy.interval=3s --probe.publish.interval 5s --probe.processes=false --probe.conntrack=false --probe.ebpf.connections=false --probe.docker.bridge=docker0 --probe.docker=true https://redacted@cloud.weave.works.", want: "5s", name: "space"},
		{cmd: "/home/weave/scope --mode=probe --no-app --probe.docker=true --probe.kubernetes.role=host --weave=false --probe.publish.interval=4500ms --probe.spy.interval=2s --probe.http.listen=:4041 --probe.conntrack.buffersize=4194304 https://redacted@cloud.weave.works scope.weave.svc.cluster.local:80", want: "4500ms", name: "miliseconds"},
		{cmd: "/home/weave/scope --mode=probe --no-app --probe.docker=true --probe.kubernetes.role=host --weave=false --probe.spy.interval=2s --probe.http.listen=:4041 --probe.conntrack.buffersize=4194304 https://redacted@cloud.weave.works scope.weave.svc.cluster.local:80", want: "", name: "notset"},
		{cmd: "/home/weave/scope --mode=probe --probe-only --probe.kubernetes.role=host --probe.publish.interval=4500ms --probe.spy.interval=10s --probe.docker.bridge=docker0 --probe.docker=true --probe.ebpf.connections=false --probe.conntrack=false https://redacted@cloud.weave.works.", want: "10s", name: "higher-spy-interval"},
		{cmd: "/bin/prometheus --config.file=/etc/prometheus/prometheus.yml --web.listen-address=:8080 --storage.tsdb.retention.time=2h --web.enable-lifecycle", want: "", name: "notscope"},
		{cmd: "", want: "", name: "blank"},
		{cmd: "/home/weave/scope --probe.publish.interval=3s", want: "3s", name: "at-end"},
		{cmd: "/home/weave/scope --probe.publish.interval=", want: "", name: "equals-blank"},
		{cmd: "/home/weave/scope --probe.publish.interval", want: "", name: "no-value"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := intervalFromCommand(tt.cmd); got != tt.want {
				t.Errorf("intervalFromCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
