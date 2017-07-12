package tracer

type Callback interface {
	TCPEventV4(TcpV4)
	TCPEventV6(TcpV6)
	LostV4(uint64)
	LostV6(uint64)
}
