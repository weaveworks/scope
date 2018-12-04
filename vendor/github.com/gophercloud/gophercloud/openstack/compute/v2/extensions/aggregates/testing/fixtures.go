package testing

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/aggregates"
	th "github.com/gophercloud/gophercloud/testhelper"
	"github.com/gophercloud/gophercloud/testhelper/client"
)

// AggregateListBody is sample response to the List call
const AggregateListBody = `
{
    "aggregates": [
        {
            "name": "test-aggregate1",
            "availability_zone": null,
            "deleted": false,
            "created_at": "2017-12-22T10:12:06.000000",
            "updated_at": null,
            "hosts": [],
            "deleted_at": null,
            "id": 1,
            "metadata": {}
        },
        {
            "name": "test-aggregate2",
            "availability_zone": "test-az",
            "deleted": false,
            "created_at": "2017-12-22T10:16:07.000000",
            "updated_at": null,
            "hosts": [
                "cmp0"
            ],
            "deleted_at": null,
            "id": 4,
            "metadata": {
                "availability_zone": "test-az"
            }
        }
    ]
}
`

// First aggregate from the AggregateListBody
var FirstFakeAggregate = aggregates.Aggregate{
	AvailabilityZone: "",
	Hosts:            []string{},
	ID:               1,
	Metadata:         map[string]string{},
	Name:             "test-aggregate1",
}

// Second aggregate from the AggregateListBody
var SecondFakeAggregate = aggregates.Aggregate{
	AvailabilityZone: "test-az",
	Hosts:            []string{"cmp0"},
	ID:               4,
	Metadata:         map[string]string{"availability_zone": "test-az"},
	Name:             "test-aggregate2",
}

// HandleListSuccessfully configures the test server to respond to a List request.
func HandleListSuccessfully(t *testing.T) {
	th.Mux.HandleFunc("/os-aggregates", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "GET")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)

		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, AggregateListBody)
	})
}
