package kubernetes

import (
	"fmt"
	"net/http"

	"github.com/ugorji/go/codec"
)

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
var GetLocalPodUIDs = func(kubeletHost string) (map[string]struct{}, error) {
	url := fmt.Sprintf("http://%s/pods/", kubeletHost)
	resp, err := http.Get(url)
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
