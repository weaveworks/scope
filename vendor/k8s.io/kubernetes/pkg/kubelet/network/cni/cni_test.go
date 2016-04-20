// +build linux

/*
Copyright 2014 The Kubernetes Authors All rights reserved.

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

package cni

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"testing"
	"text/template"

	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"

	docker "github.com/fsouza/go-dockerclient"
	cadvisorapi "github.com/google/cadvisor/info/v1"

	"k8s.io/kubernetes/cmd/kubelet/app/options"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/record"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	containertest "k8s.io/kubernetes/pkg/kubelet/container/testing"
	"k8s.io/kubernetes/pkg/kubelet/dockertools"
	"k8s.io/kubernetes/pkg/kubelet/network"
	nettest "k8s.io/kubernetes/pkg/kubelet/network/testing"
	proberesults "k8s.io/kubernetes/pkg/kubelet/prober/results"
	utiltesting "k8s.io/kubernetes/pkg/util/testing"
)

func installPluginUnderTest(t *testing.T, testVendorCNIDirPrefix, testNetworkConfigPath, vendorName string, plugName string) {
	pluginDir := path.Join(testNetworkConfigPath, plugName)
	err := os.MkdirAll(pluginDir, 0777)
	if err != nil {
		t.Fatalf("Failed to create plugin config dir: %v", err)
	}
	pluginConfig := path.Join(pluginDir, plugName+".conf")
	f, err := os.Create(pluginConfig)
	if err != nil {
		t.Fatalf("Failed to install plugin")
	}
	networkConfig := fmt.Sprintf("{ \"name\": \"%s\", \"type\": \"%s\" }", plugName, vendorName)

	_, err = f.WriteString(networkConfig)
	if err != nil {
		t.Fatalf("Failed to write network config file (%v)", err)
	}
	f.Close()

	vendorCNIDir := fmt.Sprintf(VendorCNIDirTemplate, testVendorCNIDirPrefix, vendorName)
	err = os.MkdirAll(vendorCNIDir, 0777)
	if err != nil {
		t.Fatalf("Failed to create plugin dir: %v", err)
	}
	pluginExec := path.Join(vendorCNIDir, vendorName)
	f, err = os.Create(pluginExec)

	const execScriptTempl = `#!/bin/bash
read ignore
env > {{.OutputEnv}}
echo "%@" >> {{.OutputEnv}}
export $(echo ${CNI_ARGS} | sed 's/;/ /g') &> /dev/null
mkdir -p {{.OutputDir}} &> /dev/null
echo -n "$CNI_COMMAND $CNI_NETNS $K8S_POD_NAMESPACE $K8S_POD_NAME $K8S_POD_INFRA_CONTAINER_ID" >& {{.OutputFile}}
echo -n "{ \"ip4\": { \"ip\": \"10.1.0.23/24\" } }"
`
	execTemplateData := &map[string]interface{}{
		"OutputFile": path.Join(pluginDir, plugName+".out"),
		"OutputEnv":  path.Join(pluginDir, plugName+".env"),
		"OutputDir":  pluginDir,
	}

	tObj := template.Must(template.New("test").Parse(execScriptTempl))
	buf := &bytes.Buffer{}
	if err := tObj.Execute(buf, *execTemplateData); err != nil {
		t.Fatalf("Error in executing script template - %v", err)
	}
	execScript := buf.String()
	_, err = f.WriteString(execScript)
	if err != nil {
		t.Fatalf("Failed to write plugin exec - %v", err)
	}

	err = f.Chmod(0777)
	if err != nil {
		t.Fatalf("Failed to set exec perms on plugin")
	}

	f.Close()
}

func tearDownPlugin(tmpDir string) {
	err := os.RemoveAll(tmpDir)
	if err != nil {
		fmt.Printf("Error in cleaning up test: %v", err)
	}
}

type fakeNetworkHost struct {
	kubeClient clientset.Interface
}

func NewFakeHost(kubeClient clientset.Interface) *fakeNetworkHost {
	host := &fakeNetworkHost{kubeClient: kubeClient}
	return host
}

func (fnh *fakeNetworkHost) GetPodByName(name, namespace string) (*api.Pod, bool) {
	return nil, false
}

func (fnh *fakeNetworkHost) GetKubeClient() clientset.Interface {
	return nil
}

func (nh *fakeNetworkHost) GetRuntime() kubecontainer.Runtime {
	dm, fakeDockerClient := newTestDockerManager()
	fakeDockerClient.SetFakeRunningContainers([]*docker.Container{
		{
			ID:    "test_infra_container",
			State: docker.State{Pid: 12345},
		},
	})
	return dm
}

func newTestDockerManager() (*dockertools.DockerManager, *dockertools.FakeDockerClient) {
	fakeDocker := dockertools.NewFakeDockerClient()
	fakeRecorder := &record.FakeRecorder{}
	containerRefManager := kubecontainer.NewRefManager()
	networkPlugin, _ := network.InitNetworkPlugin([]network.NetworkPlugin{}, "", nettest.NewFakeHost(nil))
	dockerManager := dockertools.NewFakeDockerManager(
		fakeDocker,
		fakeRecorder,
		proberesults.NewManager(),
		containerRefManager,
		&cadvisorapi.MachineInfo{},
		options.GetDefaultPodInfraContainerImage(),
		0, 0, "",
		containertest.FakeOS{},
		networkPlugin,
		nil,
		nil,
		nil)

	return dockerManager, fakeDocker
}

func TestCNIPlugin(t *testing.T) {
	// install some random plugin
	pluginName := fmt.Sprintf("test%d", rand.Intn(1000))
	vendorName := fmt.Sprintf("test_vendor%d", rand.Intn(1000))

	tmpDir := utiltesting.MkTmpdirOrDie("cni-test")
	testNetworkConfigPath := path.Join(tmpDir, "plugins", "net", "cni")
	testVendorCNIDirPrefix := tmpDir
	defer tearDownPlugin(tmpDir)
	installPluginUnderTest(t, testVendorCNIDirPrefix, testNetworkConfigPath, vendorName, pluginName)

	np := probeNetworkPluginsWithVendorCNIDirPrefix(path.Join(testNetworkConfigPath, pluginName), testVendorCNIDirPrefix)
	plug, err := network.InitNetworkPlugin(np, "cni", NewFakeHost(nil))
	if err != nil {
		t.Fatalf("Failed to select the desired plugin: %v", err)
	}

	err = plug.SetUpPod("podNamespace", "podName", "test_infra_container")
	if err != nil {
		t.Errorf("Expected nil: %v", err)
	}
	outputEnv := path.Join(testNetworkConfigPath, pluginName, pluginName+".env")
	eo, eerr := ioutil.ReadFile(outputEnv)
	outputFile := path.Join(testNetworkConfigPath, pluginName, pluginName+".out")
	output, err := ioutil.ReadFile(outputFile)
	if err != nil {
		t.Errorf("Failed to read output file %s: %v (env %s err %v)", outputFile, err, eo, eerr)
	}
	expectedOutput := "ADD /proc/12345/ns/net podNamespace podName test_infra_container"
	if string(output) != expectedOutput {
		t.Errorf("Mismatch in expected output for setup hook. Expected '%s', got '%s'", expectedOutput, string(output))
	}
	err = plug.TearDownPod("podNamespace", "podName", "test_infra_container")
	if err != nil {
		t.Errorf("Expected nil: %v", err)
	}
	output, err = ioutil.ReadFile(path.Join(testNetworkConfigPath, pluginName, pluginName+".out"))
	expectedOutput = "DEL /proc/12345/ns/net podNamespace podName test_infra_container"
	if string(output) != expectedOutput {
		t.Errorf("Mismatch in expected output for setup hook. Expected '%s', got '%s'", expectedOutput, string(output))
	}
}
