package p2p

import "net"

const (
	IncomingStream = 0x2
	IcomingMessage = 0x1
)

// Message represent any arbitrary data that is being sent over
// each transport between two nodes in the network
type RPC struct {
	From    net.Addr
	Payload []byte
	Stream  bool
}
