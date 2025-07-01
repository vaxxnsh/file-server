package p2p

import "net"

// Peer is a interface that represent remote node

type Peer interface {
	net.Conn
	Send([]byte) error
}

/*
	Transport is anything that handles the communication between
	between the nodes in the network. this can be of the form
	(TCP, UDP, websockets, ...)
*/

type Transport interface {
	Dial(string) error
	ListenAndAccept()
	Consume() <-chan RPC
	Close() error
}
