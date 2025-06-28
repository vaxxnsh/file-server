package p2p

import "errors"

type HandShakeFunc func(any) error

// Error invalid handshake is returned if
// the handshake between local and remote node couldn't be established
var ErrorInvalidHandshake = errors.New("invalid handshake")

func NOHandShakeFunc(any) error {
	return nil
}
