package socket

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/netsys-lab/scion-path-discovery/packets"
	"github.com/scionproto/scion/go/lib/snet"
)

var _ UnderlaySocket = (*SCIONSocket)(nil)

type DialPacket struct {
	Addr snet.UDPAddr
}

// TODO: extend this further. It may be useful to use more than
// one native UDP socket due to performance limitations
//type Socket interface {
//	net.Conn
//}

type SCIONSocket struct {
	listenConns          []packets.UDPConn
	local                string
	localAddr            *snet.UDPAddr
	transportConstructor packets.TransportConstructor
	dialConns            []packets.UDPConn
}

func NewSCIONSocket(local string, transportConstructor packets.TransportConstructor) *SCIONSocket {
	s := SCIONSocket{
		local:                local,
		transportConstructor: transportConstructor,
		listenConns:          make([]packets.UDPConn, 0),
		dialConns:            make([]packets.UDPConn, 0),
	}

	return &s
}

func (s *SCIONSocket) Listen() error {
	lAddr, err := snet.ParseUDPAddr(s.local)
	if err != nil {
		return err
	}

	s.localAddr = lAddr
	conn := s.transportConstructor()
	s.listenConns = append(s.listenConns, conn)
	return conn.Listen(*s.localAddr)
}

func (s *SCIONSocket) WaitForDialIn() (*snet.UDPAddr, error) {
	// TODO: Close
	bts := make([]byte, packets.PACKET_SIZE)
	// We assume that the first conn here is always the one that was initialized by listen()
	// Other cons could be added due to handshakes (QUIC specific)
	fmt.Printf("Waiting for input on %s", s.local)
	_, err := s.listenConns[0].Read(bts)
	fmt.Println("Read something")
	if err != nil {
		return nil, err
	}
	p := DialPacket{}
	network := bytes.NewBuffer(bts) // Stand-in for a network connection
	dec := gob.NewDecoder(network)
	err = dec.Decode(&p)
	if err != nil {
		return nil, err
	}

	addr := p.Addr

	return &addr, nil
}

func (s *SCIONSocket) Dial(remote snet.UDPAddr, path snet.Path, options DialOptions) (packets.UDPConn, error) {
	// appnet.SetPath(&remote, path)
	// fmt.Printf("Dialing to %s via %s\n", remote.String(), remote.Path)
	conn := s.transportConstructor()
	conn.SetLocal(*s.localAddr)
	err := conn.Dial(remote, &path)
	if err != nil {
		return nil, err
	}

	if options.SendAddrPacket {
		var network bytes.Buffer
		enc := gob.NewEncoder(&network) // Will write to network.
		p := DialPacket{
			Addr: remote,
		}

		err := enc.Encode(p)
		conn.Write(network.Bytes())
		if err != nil {
			return nil, err
		}
	}

	s.dialConns = append(s.dialConns, conn)

	return conn, nil
}

func (s *SCIONSocket) DialAll(remote snet.UDPAddr, path []snet.Path, options DialOptions) ([]packets.UDPConn, error) {
	conns := make([]packets.UDPConn, 0)
	for _, v := range path {
		conn, err := s.Dial(remote, v, options)
		if err != nil {
			return nil, err
		}
		conns = append(conns, conn)
	}
	fmt.Println("Dial all#1")

	return conns, nil
}

func (s *SCIONSocket) GetListenConnections() []packets.UDPConn {
	return s.listenConns
}

func (s *SCIONSocket) GetDialConnections() []packets.UDPConn {
	return s.dialConns
}

func (s *SCIONSocket) CloseAll() []error {
	errors := make([]error, 0)
	for _, con := range s.dialConns {
		err := con.Close()
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}