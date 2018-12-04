package aggregates

import "github.com/gophercloud/gophercloud"

func aggregatesListURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL("os-aggregates")
}
