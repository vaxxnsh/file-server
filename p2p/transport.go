package p2p

import "net"

// Peer is a interface that represent remote node

type Peer interface {
	net.Conn
	Send([]byte) error
	CloseStream()
}

/*
	Transport is anything that handles the communication between
	between the nodes in the network. this can be of the form
	(TCP, UDP, websockets, ...)
*/

type Transport interface {
	Addr() string
	Dial(string) error
	ListenAndAccept()
	Consume() <-chan RPC
	Close() error
}
