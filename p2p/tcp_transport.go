package p2p

import (
	"fmt"
	"net"
	"sync"
)

type TCPPeer struct {
	conn     net.Conn
	outbound bool
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandShakeFunc
	Decoder       Decoder
	OnPeer        func(peer Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listerner net.Listener
	rpcChan   chan RPC // channel to send RPCs to
	mu        sync.RWMutex
	peers     map[net.Addr]Peer
}

func NewTCPTransport(ops TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: ops,
		rpcChan:          make(chan RPC), // buffered channel for RPCs
	}
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

func (t *TCPTransport) ListenAndAccept() (err error) {
	t.listerner, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return
	}
	go t.startAcceptLoop()

	return
}

// Consume returns a read-only channel that can be used to consume RPCs
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcChan
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listerner.Accept()
		if err != nil {
			fmt.Printf("TCP accept error : %s\n", err)
		}
		fmt.Printf("new incoming connection %+v\n", conn)
		go t.handleConn(conn)
	}
}

type Temp struct{}

func (t *TCPTransport) handleConn(conn net.Conn) {
	var err error

	defer func() {
		fmt.Printf("Dropping peer connection: %s\n", err)
		conn.Close()

	}()

	peer := NewTCPPeer(conn, true)

	if err := t.HandshakeFunc(peer); err != nil {
		conn.Close()
		fmt.Printf("TCP handshake error : %s\n", err)
		return
	}

	if t.OnPeer != nil {
		if err := t.OnPeer(peer); err != nil {
			return
		}
	}

	rpc := RPC{}

	// Read Loop
	for {
		err = t.Decoder.Decode(conn, &rpc)
		if err != nil {
			fmt.Printf("TCP error : %s\n", err)
			return
		}
		rpc.From = conn.RemoteAddr()
		t.rpcChan <- rpc
		// fmt.Printf("message : %+v\n", rpc)
	}
}
