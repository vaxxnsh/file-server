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

type TCPTransport struct {
	listenAddress string
	listerner     net.Listener
	mu            sync.RWMutex
	peers         map[net.Addr]Peer
}

func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		listenAddress: listenAddr,
	}
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

func (t *TCPTransport) ListenAndAccept() (err error) {
	t.listerner, err = net.Listen("tcp", t.listenAddress)
	if err != nil {
		return
	}
	go t.startAcceptLoop()

	return
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listerner.Accept()
		if err != nil {
			fmt.Printf("TCP accept error : %s\n", err)
		}
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)
	fmt.Printf("new incoming connection %+v\n", peer)
}
