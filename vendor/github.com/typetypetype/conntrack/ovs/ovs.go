package ovs

import (
	"errors"
	"fmt"
	"strconv"
)

func FollowOvsFlows(bufferSize int, flags uint32) (<-chan *OvsFlowInfo, func(), error) {

	dpif, err := NewDpifGroups(0)

	if err != nil {
		fmt.Println(err)
		return nil, nil, err
	}

	dp, _, err := lookupDatapath(dpif, "ovs-system")

	res, stop, err := dp.FollowFlows()
	return res, func() { stop(); dpif.Close() }, err

}

func lookupDatapath(dpif *Dpif, name string) (*DatapathHandle, string, error) {
	dph, err := dpif.LookupDatapath(name)
	if err == nil {
		return &dph, name, nil
	}

	if !IsNoSuchDatapathError(err) {
		return nil, "", err
	}

	// If the name is a number, try to use it as an id
	ifindex, err := strconv.ParseUint(name, 10, 32)
	if err == nil {
		dp, err := dpif.LookupDatapathByID(DatapathID(ifindex))
		if err == nil {
			return &dp.Handle, dp.Name, nil
		}

		if !IsNoSuchDatapathError(err) {

			return nil, "", err
		}
	}

	return nil, "", errors.New(fmt.Sprintf("Cannot find datapath \"%s\"", name))
}
