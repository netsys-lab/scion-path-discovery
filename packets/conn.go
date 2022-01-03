package packets

import (
	"net"

	"github.com/scionproto/scion/go/lib/snet"
)

// TODO: Remove to configurable options
const (
	PACKET_SIZE = 1400
)

func ConnTypeToString(connType int) string {
	switch connType {
	case ConnectionTypes.Bidirectional:
		return "bidirectional"

	case ConnectionTypes.Outgoing:
		return "outgoing"

	case ConnectionTypes.Incoming:
		return "incoming"
	}

	return ""
}

var ConnectionTypes = newConnectionTypes()

func newConnectionTypes() *connectionTypes {
	return &connectionTypes{
		Incoming:      1,
		Outgoing:      2,
		Bidirectional: 3,
	}
}

type connectionTypes struct {
	Incoming      int
	Outgoing      int
	Bidirectional int
}

var ConnectionStates = newConnectionStates()

func newConnectionStates() *connectionStates {
	return &connectionStates{
		Pending: 1,
		Open:    2,
		Closed:  3,
	}
}

type connectionStates struct {
	Pending int
	Open    int
	Closed  int
}

type BasicConn struct {
	state int
}

func (c *BasicConn) GetState() int {
	return c.state
}

type UDPConn interface {
	net.Conn
	Listen(snet.UDPAddr) error
	Dial(snet.UDPAddr, *snet.Path) error
	GetState() int
	GetMetrics() *PathMetrics
	GetPath() *snet.Path
	GetRemote() *snet.UDPAddr
	SetLocal(snet.UDPAddr)
	WriteStream([]byte) (int, error)
	ReadStream([]byte) (int, error)
	GetType() int
	GetId() string
	SetId(string)
	MarkAsClosed() error
}

type TransportConstructor func() UDPConn
