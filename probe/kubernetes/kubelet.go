package kubernetes

import (
	"net/http"

	"github.com/ugorji/go/codec"
)

// KubeletURL is just exported for testing
var KubeletURL = "http://localhost:10255"

// Intentionally not using the full kubernetes library DS
// to make parsing faster and more tolerant to schema changes
type podList struct {
	Items []struct {
		Metadata struct {
			UID string `json:"uid"`
		} `json:"metadata"`
	} `json:"items"`
}

// GetLocalPodUIDs obtains the UID of the pods run locally (it's just exported for testing)
var GetLocalPodUIDs = func() (map[string]struct{}, error) {
	resp, err := http.Get(KubeletURL + "/pods/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var localPods podList
	if err := codec.NewDecoder(resp.Body, &codec.JsonHandle{}).Decode(&localPods); err != nil {
		return nil, err
	}
	result := make(map[string]struct{}, len(localPods.Items))
	for _, pod := range localPods.Items {
		result[pod.Metadata.UID] = struct{}{}
	}
	return result, nil
}
