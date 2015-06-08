package report_test

import (
	"github.com/weaveworks/scope/report"
)

var (
	clientHostID     = "client.host.com"
	clientHostName   = clientHostID
	clientHostNodeID = report.MakeHostNodeID(clientHostID)
	clientAddress    = "10.10.10.20"
	serverHostID     = "server.host.com"
	serverHostName   = serverHostID
	serverHostNodeID = report.MakeHostNodeID(serverHostID)
	serverAddress    = "10.10.10.1"
	unknownHostID    = ""              // by definition, we don't know it
	unknownAddress   = "172.16.93.112" // will be a pseudonode, no corresponding host

	client54001EndpointNodeID = report.MakeEndpointNodeID(clientHostID, clientAddress, "54001") // i.e. curl
	client54002EndpointNodeID = report.MakeEndpointNodeID(clientHostID, clientAddress, "54002") // also curl
	server80EndpointNodeID    = report.MakeEndpointNodeID(serverHostID, serverAddress, "80")    // i.e. apache
	unknown1EndpointNodeID    = report.MakeEndpointNodeID(unknownHostID, unknownAddress, "10001")
	unknown2EndpointNodeID    = report.MakeEndpointNodeID(unknownHostID, unknownAddress, "10002")
	unknown3EndpointNodeID    = report.MakeEndpointNodeID(unknownHostID, unknownAddress, "10003")

	clientAddressNodeID  = report.MakeAddressNodeID(clientHostID, clientAddress)
	serverAddressNodeID  = report.MakeAddressNodeID(serverHostID, serverAddress)
	unknownAddressNodeID = report.MakeAddressNodeID(unknownHostID, unknownAddress)
)
