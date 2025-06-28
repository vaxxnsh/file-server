package p2p

// Peer is a interface that represent remote node

type Peer any

/*
	Transport is anything that handles the communication between
	between the nodes in the network. this can be of the form
	(TCP, UDP, websockets, ...)
*/

type Transport interface {
	ListenAndAccept()
	Consume() <-chan RPC
}
