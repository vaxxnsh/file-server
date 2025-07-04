package p2p

import (
	"fmt"
	"log"
	"net"
	"sync"
)

type TCPPeer struct {
	// underlying connection of the peer, i.e. TCP
	net.Conn
	outbound bool
	wg       *sync.WaitGroup
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandShakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listerner net.Listener
	rpcChan   chan RPC // channel to send RPCs to
}

func NewTCPTransport(ops TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: ops,
		rpcChan:          make(chan RPC, 1024), // buffered channel for RPCs
	}
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

func (t *TCPTransport) Addr() string {
	return t.ListenAddr
}

func (t *TCPPeer) CloseStream() {
	t.wg.Done()
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)

	return err
}

func (t *TCPTransport) ListenAndAccept() (err error) {
	t.listerner, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return
	}
	go t.startAcceptLoop()
	log.Printf("TCP transport listening on port : %s", t.ListenAddr)
	return
}

// Consume returns a read-only channel that can be used to consume RPCs
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcChan
}

func (t *TCPTransport) Close() error {
	return t.listerner.Close()
}

func (t *TCPTransport) Dial(addr string) error {
	fmt.Println(addr)
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		return err
	}

	go t.handleConn(conn, true)

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listerner.Accept()
		if err != nil {
			fmt.Printf("TCP accept error : %s\n", err)
		}
		fmt.Printf("new incoming connection %+v\n", conn)
		go t.handleConn(conn, false)
	}
}

type Temp struct{}

func (t *TCPTransport) handleConn(conn net.Conn, outBound bool) {
	var err error

	defer func() {
		fmt.Printf("Dropping peer connection: %s\n", err)
		conn.Close()

	}()

	peer := NewTCPPeer(conn, outBound)

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

	// Read Loop
	for {
		rpc := RPC{}
		err = t.Decoder.Decode(conn, &rpc)
		if err != nil {
			fmt.Printf("TCP error : %s\n", err)
			return
		}

		rpc.From = conn.RemoteAddr()

		if rpc.Stream {
			peer.wg.Add(1)
			fmt.Printf("[%s] incomming stream waiting ...\n", conn.RemoteAddr())
			peer.wg.Wait()
			fmt.Printf("[%s] stream closed, resuming read loop\n", conn.RemoteAddr())
			continue
		}

		t.rpcChan <- rpc
	}
}
