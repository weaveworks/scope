/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package stats

import (
	"testing"
	"time"

	"github.com/google/cadvisor/info/v1"
	"github.com/google/cadvisor/info/v2"
	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kubestats "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/stats"
	"k8s.io/kubernetes/pkg/kubelet/cm"
	"k8s.io/kubernetes/pkg/kubelet/leaky"
)

const (
	// Offsets from seed value in generated container stats.
	offsetCPUUsageCores = iota
	offsetCPUUsageCoreSeconds
	offsetMemPageFaults
	offsetMemMajorPageFaults
	offsetMemUsageBytes
	offsetMemRSSBytes
	offsetMemWorkingSetBytes
	offsetNetRxBytes
	offsetNetRxErrors
	offsetNetTxBytes
	offsetNetTxErrors
)

var (
	timestamp    = time.Now()
	creationTime = timestamp.Add(-5 * time.Minute)
)

func TestBuildSummary(t *testing.T) {
	node := api.Node{}
	node.Name = "FooNode"
	nodeConfig := cm.NodeConfig{
		RuntimeCgroupsName: "/docker-daemon",
		SystemCgroupsName:  "/system",
		KubeletCgroupsName: "/kubelet",
	}
	const (
		namespace0 = "test0"
		namespace2 = "test2"
	)
	const (
		seedRoot           = 0
		seedRuntime        = 100
		seedKubelet        = 200
		seedMisc           = 300
		seedPod0Infra      = 1000
		seedPod0Container0 = 2000
		seedPod0Container1 = 2001
		seedPod1Infra      = 3000
		seedPod1Container  = 4000
		seedPod2Infra      = 5000
		seedPod2Container  = 6000
	)
	const (
		pName0 = "pod0"
		pName1 = "pod1"
		pName2 = "pod0" // ensure pName2 conflicts with pName0, but is in a different namespace
	)
	const (
		cName00 = "c0"
		cName01 = "c1"
		cName10 = "c0" // ensure cName10 conflicts with cName02, but is in a different pod
		cName20 = "c1" // ensure cName20 conflicts with cName01, but is in a different pod + namespace
	)

	prf0 := kubestats.PodReference{Name: pName0, Namespace: namespace0, UID: "UID" + pName0}
	prf1 := kubestats.PodReference{Name: pName1, Namespace: namespace0, UID: "UID" + pName1}
	prf2 := kubestats.PodReference{Name: pName2, Namespace: namespace2, UID: "UID" + pName2}
	infos := map[string]v2.ContainerInfo{
		"/":              summaryTestContainerInfo(seedRoot, "", "", ""),
		"/docker-daemon": summaryTestContainerInfo(seedRuntime, "", "", ""),
		"/kubelet":       summaryTestContainerInfo(seedKubelet, "", "", ""),
		"/system":        summaryTestContainerInfo(seedMisc, "", "", ""),
		// Pod0 - Namespace0
		"/pod0-i":  summaryTestContainerInfo(seedPod0Infra, pName0, namespace0, leaky.PodInfraContainerName),
		"/pod0-c0": summaryTestContainerInfo(seedPod0Container0, pName0, namespace0, cName00),
		"/pod0-c2": summaryTestContainerInfo(seedPod0Container1, pName0, namespace0, cName01),
		// Pod1 - Namespace0
		"/pod1-i":  summaryTestContainerInfo(seedPod1Infra, pName1, namespace0, leaky.PodInfraContainerName),
		"/pod1-c0": summaryTestContainerInfo(seedPod1Container, pName1, namespace0, cName10),
		// Pod2 - Namespace2
		"/pod2-i":  summaryTestContainerInfo(seedPod2Infra, pName2, namespace2, leaky.PodInfraContainerName),
		"/pod2-c0": summaryTestContainerInfo(seedPod2Container, pName2, namespace2, cName20),
	}

	rootfs := v2.FsInfo{}
	imagefs := v2.FsInfo{}

	sb := &summaryBuilder{
		newFsResourceAnalyzer(&MockStatsProvider{}, time.Minute*5), &node, nodeConfig, rootfs, imagefs, infos}
	summary, err := sb.build()

	assert.NoError(t, err)
	nodeStats := summary.Node
	assert.Equal(t, "FooNode", nodeStats.NodeName)
	assert.EqualValues(t, testTime(creationTime, seedRoot).Unix(), nodeStats.StartTime.Time.Unix())
	checkCPUStats(t, "Node", seedRoot, nodeStats.CPU)
	checkMemoryStats(t, "Node", seedRoot, nodeStats.Memory)
	checkNetworkStats(t, "Node", seedRoot, nodeStats.Network)

	systemSeeds := map[string]int{
		kubestats.SystemContainerRuntime: seedRuntime,
		kubestats.SystemContainerKubelet: seedKubelet,
		kubestats.SystemContainerMisc:    seedMisc,
	}
	for _, sys := range nodeStats.SystemContainers {
		name := sys.Name
		seed, found := systemSeeds[name]
		if !found {
			t.Errorf("Unknown SystemContainer: %q", name)
		}
		assert.EqualValues(t, testTime(creationTime, seed).Unix(), sys.StartTime.Time.Unix(), name+".StartTime")
		checkCPUStats(t, name, seed, sys.CPU)
		checkMemoryStats(t, name, seed, sys.Memory)
	}

	assert.Equal(t, 3, len(summary.Pods))
	indexPods := make(map[kubestats.PodReference]kubestats.PodStats, len(summary.Pods))
	for _, pod := range summary.Pods {
		indexPods[pod.PodRef] = pod
	}

	// Validate Pod0 Results
	ps, found := indexPods[prf0]
	assert.True(t, found)
	assert.Len(t, ps.Containers, 2)
	indexCon := make(map[string]kubestats.ContainerStats, len(ps.Containers))
	for _, con := range ps.Containers {
		indexCon[con.Name] = con
	}
	con := indexCon[cName00]
	assert.EqualValues(t, testTime(creationTime, seedPod0Container0).Unix(), con.StartTime.Time.Unix())
	checkCPUStats(t, "container", seedPod0Container0, con.CPU)
	checkMemoryStats(t, "container", seedPod0Container0, con.Memory)

	con = indexCon[cName01]
	assert.EqualValues(t, testTime(creationTime, seedPod0Container1).Unix(), con.StartTime.Time.Unix())
	checkCPUStats(t, "container", seedPod0Container1, con.CPU)
	checkMemoryStats(t, "container", seedPod0Container1, con.Memory)

	assert.EqualValues(t, testTime(creationTime, seedPod0Infra).Unix(), ps.StartTime.Time.Unix())
	checkNetworkStats(t, "Pod", seedPod0Infra, ps.Network)

	// Validate Pod1 Results
	ps, found = indexPods[prf1]
	assert.True(t, found)
	assert.Len(t, ps.Containers, 1)
	con = ps.Containers[0]
	assert.Equal(t, cName10, con.Name)
	checkCPUStats(t, "container", seedPod1Container, con.CPU)
	checkMemoryStats(t, "container", seedPod1Container, con.Memory)
	checkNetworkStats(t, "Pod", seedPod1Infra, ps.Network)

	// Validate Pod2 Results
	ps, found = indexPods[prf2]
	assert.True(t, found)
	assert.Len(t, ps.Containers, 1)
	con = ps.Containers[0]
	assert.Equal(t, cName20, con.Name)
	checkCPUStats(t, "container", seedPod2Container, con.CPU)
	checkMemoryStats(t, "container", seedPod2Container, con.Memory)
	checkNetworkStats(t, "Pod", seedPod2Infra, ps.Network)
}

func generateCustomMetricSpec() []v1.MetricSpec {
	f := fuzz.New().NilChance(0).Funcs(
		func(e *v1.MetricSpec, c fuzz.Continue) {
			c.Fuzz(&e.Name)
			switch c.Intn(3) {
			case 0:
				e.Type = v1.MetricGauge
			case 1:
				e.Type = v1.MetricCumulative
			case 2:
				e.Type = v1.MetricDelta
			}
			switch c.Intn(2) {
			case 0:
				e.Format = v1.IntType
			case 1:
				e.Format = v1.FloatType
			}
			c.Fuzz(&e.Units)
		})
	var ret []v1.MetricSpec
	f.Fuzz(&ret)
	return ret
}

func generateCustomMetrics(spec []v1.MetricSpec) map[string][]v1.MetricVal {
	ret := map[string][]v1.MetricVal{}
	for _, metricSpec := range spec {
		f := fuzz.New().NilChance(0).Funcs(
			func(e *v1.MetricVal, c fuzz.Continue) {
				switch metricSpec.Format {
				case v1.IntType:
					c.Fuzz(&e.IntValue)
				case v1.FloatType:
					c.Fuzz(&e.FloatValue)
				}
			})

		var metrics []v1.MetricVal
		f.Fuzz(&metrics)
		ret[metricSpec.Name] = metrics
	}
	return ret
}

func summaryTestContainerInfo(seed int, podName string, podNamespace string, containerName string) v2.ContainerInfo {
	labels := map[string]string{}
	if podName != "" {
		labels = map[string]string{
			"io.kubernetes.pod.name":       podName,
			"io.kubernetes.pod.uid":        "UID" + podName,
			"io.kubernetes.pod.namespace":  podNamespace,
			"io.kubernetes.container.name": containerName,
		}
	}
	spec := v2.ContainerSpec{
		CreationTime:  testTime(creationTime, seed),
		HasCpu:        true,
		HasMemory:     true,
		HasNetwork:    true,
		Labels:        labels,
		CustomMetrics: generateCustomMetricSpec(),
	}

	stats := v2.ContainerStats{
		Timestamp: testTime(timestamp, seed),
		Cpu:       &v1.CpuStats{},
		CpuInst:   &v2.CpuInstStats{},
		Memory: &v1.MemoryStats{
			Usage:      uint64(seed + offsetMemUsageBytes),
			WorkingSet: uint64(seed + offsetMemWorkingSetBytes),
			RSS:        uint64(seed + offsetMemRSSBytes),
			ContainerData: v1.MemoryStatsMemoryData{
				Pgfault:    uint64(seed + offsetMemPageFaults),
				Pgmajfault: uint64(seed + offsetMemMajorPageFaults),
			},
		},
		Network: &v2.NetworkStats{
			Interfaces: []v1.InterfaceStats{{
				Name:     "eth0",
				RxBytes:  uint64(seed + offsetNetRxBytes),
				RxErrors: uint64(seed + offsetNetRxErrors),
				TxBytes:  uint64(seed + offsetNetTxBytes),
				TxErrors: uint64(seed + offsetNetTxErrors),
			}, {
				Name:     "cbr0",
				RxBytes:  100,
				RxErrors: 100,
				TxBytes:  100,
				TxErrors: 100,
			}},
		},
		CustomMetrics: generateCustomMetrics(spec.CustomMetrics),
	}
	stats.Cpu.Usage.Total = uint64(seed + offsetCPUUsageCoreSeconds)
	stats.CpuInst.Usage.Total = uint64(seed + offsetCPUUsageCores)
	return v2.ContainerInfo{
		Spec:  spec,
		Stats: []*v2.ContainerStats{&stats},
	}
}

func testTime(base time.Time, seed int) time.Time {
	return base.Add(time.Duration(seed) * time.Second)
}

func checkNetworkStats(t *testing.T, label string, seed int, stats *kubestats.NetworkStats) {
	assert.EqualValues(t, testTime(timestamp, seed).Unix(), stats.Time.Time.Unix(), label+".Net.Time")
	assert.EqualValues(t, seed+offsetNetRxBytes, *stats.RxBytes, label+".Net.RxBytes")
	assert.EqualValues(t, seed+offsetNetRxErrors, *stats.RxErrors, label+".Net.RxErrors")
	assert.EqualValues(t, seed+offsetNetTxBytes, *stats.TxBytes, label+".Net.TxBytes")
	assert.EqualValues(t, seed+offsetNetTxErrors, *stats.TxErrors, label+".Net.TxErrors")
}

func checkCPUStats(t *testing.T, label string, seed int, stats *kubestats.CPUStats) {
	assert.EqualValues(t, testTime(timestamp, seed).Unix(), stats.Time.Time.Unix(), label+".CPU.Time")
	assert.EqualValues(t, seed+offsetCPUUsageCores, *stats.UsageNanoCores, label+".CPU.UsageCores")
	assert.EqualValues(t, seed+offsetCPUUsageCoreSeconds, *stats.UsageCoreNanoSeconds, label+".CPU.UsageCoreSeconds")
}

func checkMemoryStats(t *testing.T, label string, seed int, stats *kubestats.MemoryStats) {
	assert.EqualValues(t, testTime(timestamp, seed).Unix(), stats.Time.Time.Unix(), label+".Mem.Time")
	assert.EqualValues(t, seed+offsetMemUsageBytes, *stats.UsageBytes, label+".Mem.UsageBytes")
	assert.EqualValues(t, seed+offsetMemWorkingSetBytes, *stats.WorkingSetBytes, label+".Mem.WorkingSetBytes")
	assert.EqualValues(t, seed+offsetMemRSSBytes, *stats.RSSBytes, label+".Mem.RSSBytes")
	assert.EqualValues(t, seed+offsetMemPageFaults, *stats.PageFaults, label+".Mem.PageFaults")
	assert.EqualValues(t, seed+offsetMemMajorPageFaults, *stats.MajorPageFaults, label+".Mem.MajorPageFaults")
}

func TestCustomMetrics(t *testing.T) {
	spec := []v1.MetricSpec{
		{
			Name:   "qos",
			Type:   v1.MetricGauge,
			Format: v1.IntType,
			Units:  "per second",
		},
		{
			Name:   "cpuLoad",
			Type:   v1.MetricCumulative,
			Format: v1.FloatType,
			Units:  "count",
		},
	}
	timestamp1 := time.Now()
	timestamp2 := time.Now().Add(time.Minute)
	metrics := map[string][]v1.MetricVal{
		"qos": {
			{
				Timestamp: timestamp1,
				IntValue:  10,
			},
			{
				Timestamp: timestamp2,
				IntValue:  100,
			},
		},
		"cpuLoad": {
			{
				Timestamp:  timestamp1,
				FloatValue: 1.2,
			},
			{
				Timestamp:  timestamp2,
				FloatValue: 2.1,
			},
		},
	}
	cInfo := v2.ContainerInfo{
		Spec: v2.ContainerSpec{
			CustomMetrics: spec,
		},
		Stats: []*v2.ContainerStats{
			{
				CustomMetrics: metrics,
			},
		},
	}
	sb := &summaryBuilder{}
	assert.Contains(t, sb.containerInfoV2ToUserDefinedMetrics(&cInfo),
		kubestats.UserDefinedMetric{
			UserDefinedMetricDescriptor: kubestats.UserDefinedMetricDescriptor{
				Name:  "qos",
				Type:  kubestats.MetricGauge,
				Units: "per second",
			},
			Time:  unversioned.NewTime(timestamp2),
			Value: 100,
		},
		kubestats.UserDefinedMetric{
			UserDefinedMetricDescriptor: kubestats.UserDefinedMetricDescriptor{
				Name:  "cpuLoad",
				Type:  kubestats.MetricCumulative,
				Units: "count",
			},
			Time:  unversioned.NewTime(timestamp2),
			Value: 2.1,
		})
}
