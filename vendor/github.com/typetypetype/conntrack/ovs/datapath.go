package ovs

import (
	"fmt"
	"syscall"
)

// Datapaths are identified by the Ifindex of their netdev.
type DatapathID int32

type datapathInfo struct {
	ifindex DatapathID
	name    string
}

func (dpif *Dpif) parseDatapathInfo(msg *NlMsgParser, cmd int) (res datapathInfo, err error) {
	_, ovshdr, err := dpif.checkNlMsgHeaders(msg, DATAPATH, cmd)
	if err != nil {
		return
	}

	res.ifindex = ovshdr.datapathID()
	attrs, err := msg.TakeAttrs()
	if err != nil {
		return
	}

	res.name, err = attrs.GetString(OVS_DP_ATTR_NAME)
	return
}

type DatapathHandle struct {
	Dpif    *Dpif
	Ifindex DatapathID
}

func (dp DatapathHandle) ID() DatapathID {
	return dp.Ifindex
}

func (dp DatapathHandle) Reopen() (DatapathHandle, error) {
	dpif, err := dp.Dpif.Reopen()
	return DatapathHandle{Dpif: dpif, Ifindex: dp.Ifindex}, err
}

func (dpif *Dpif) CreateDatapath(name string) (DatapathHandle, error) {
	var features uint32 = OVS_DP_F_UNALIGNED | OVS_DP_F_VPORT_PIDS

	req := NewNlMsgBuilder(RequestFlags, dpif.families[DATAPATH].id)
	req.PutGenlMsghdr(OVS_DP_CMD_NEW, OVS_DATAPATH_VERSION)
	req.putOvsHeader(0)
	req.PutStringAttr(OVS_DP_ATTR_NAME, name)
	req.PutUint32Attr(OVS_DP_ATTR_UPCALL_PID, 0)
	req.PutUint32Attr(OVS_DP_ATTR_USER_FEATURES, features)

	resp, err := dpif.sock.Request(req)
	if err != nil {
		return DatapathHandle{}, err
	}

	dpi, err := dpif.parseDatapathInfo(resp, OVS_DP_CMD_NEW)
	if err != nil {
		return DatapathHandle{}, err
	}

	return DatapathHandle{Dpif: dpif, Ifindex: dpi.ifindex}, nil
}

func (dpif *Dpif) LookupDatapath(name string) (DatapathHandle, error) {
	req := NewNlMsgBuilder(RequestFlags, dpif.families[DATAPATH].id)
	req.PutGenlMsghdr(OVS_DP_CMD_GET, OVS_DATAPATH_VERSION)
	req.putOvsHeader(0)
	req.PutStringAttr(OVS_DP_ATTR_NAME, name)

	resp, err := dpif.sock.Request(req)
	if err != nil {
		return DatapathHandle{}, err
	}

	dpi, err := dpif.parseDatapathInfo(resp, OVS_DP_CMD_GET)
	if err != nil {
		return DatapathHandle{}, err
	}

	return DatapathHandle{Dpif: dpif, Ifindex: dpi.ifindex}, nil
}

type Datapath struct {
	Handle DatapathHandle
	Name   string
}

func (dpif *Dpif) LookupDatapathByID(ifindex DatapathID) (Datapath, error) {
	req := NewNlMsgBuilder(RequestFlags, dpif.families[DATAPATH].id)
	req.PutGenlMsghdr(OVS_DP_CMD_GET, OVS_DATAPATH_VERSION)
	req.putOvsHeader(ifindex)

	resp, err := dpif.sock.Request(req)
	if err != nil {
		return Datapath{}, err
	}

	dpi, err := dpif.parseDatapathInfo(resp, OVS_DP_CMD_GET)
	if err != nil {
		return Datapath{}, err
	}

	return Datapath{
		Handle: DatapathHandle{Dpif: dpif, Ifindex: ifindex},
		Name:   dpi.name,
	}, nil
}

func IsNoSuchDatapathError(err error) bool {
	return err == NetlinkError(syscall.ENODEV)
}

func (dpif *Dpif) EnumerateDatapaths() (map[string]DatapathHandle, error) {
	res := make(map[string]DatapathHandle)

	req := NewNlMsgBuilder(DumpFlags, dpif.families[DATAPATH].id)
	req.PutGenlMsghdr(OVS_DP_CMD_GET, OVS_DATAPATH_VERSION)
	req.putOvsHeader(0)

	consumer := func(resp *NlMsgParser) error {
		dpi, err := dpif.parseDatapathInfo(resp, OVS_DP_CMD_GET)
		if err != nil {
			return err
		}
		res[dpi.name] = DatapathHandle{Dpif: dpif, Ifindex: dpi.ifindex}
		return nil
	}

	err := dpif.sock.RequestMulti(req, consumer)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (dp DatapathHandle) Delete() error {
	req := NewNlMsgBuilder(RequestFlags, dp.Dpif.families[DATAPATH].id)
	req.PutGenlMsghdr(OVS_DP_CMD_DEL, OVS_DATAPATH_VERSION)
	req.putOvsHeader(dp.Ifindex)

	_, err := dp.Dpif.sock.Request(req)
	if err != nil {
		return err
	}

	dp.Dpif = nil
	dp.Ifindex = 0
	return nil
}

func (dp DatapathHandle) checkNlMsgHeaders(msg *NlMsgParser, family int, cmd int) error {
	_, ovshdr, err := dp.Dpif.checkNlMsgHeaders(msg, family, cmd)
	if err != nil {
		return err
	}

	if ovshdr.datapathID() != dp.Ifindex {
		return fmt.Errorf("wrong datapath Ifindex received (got %d, expected %d)", ovshdr.datapathID(), dp.Ifindex)
	}

	return nil
}
