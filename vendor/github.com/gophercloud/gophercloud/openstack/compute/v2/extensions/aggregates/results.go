package aggregates

import "github.com/gophercloud/gophercloud/pagination"

// Aggregate represents a host aggregate in the OpenStack cloud.
type Aggregate struct {
	// The availability zone of the host aggregate.
	AvailabilityZone string `json:"availability_zone"`

	// A list of host ids in this aggregate.
	Hosts []string `json:"hosts"`

	// The ID of the host aggregate.
	ID int `json:"id"`

	// Metadata key and value pairs associate with the aggregate.
	Metadata map[string]string `json:"metadata"`

	// Name of the aggregate.
	Name string `json:"name"`
}

// AggregatesPage represents a single page of all Aggregates from a List
// request.
type AggregatesPage struct {
	pagination.SinglePageBase
}

// IsEmpty determines whether or not a page of Aggregates contains any results.
func (page AggregatesPage) IsEmpty() (bool, error) {
	aggregates, err := ExtractAggregates(page)
	return len(aggregates) == 0, err
}

// ExtractAggregates interprets a page of results as a slice of Aggregates.
func ExtractAggregates(p pagination.Page) ([]Aggregate, error) {
	var a struct {
		Aggregates []Aggregate `json:"aggregates"`
	}
	err := (p.(AggregatesPage)).ExtractInto(&a)
	return a.Aggregates, err
}
