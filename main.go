package main

import (
	"fmt"
	"log"

	"github.com/vaxxnsh/file-server/p2p"
)

func OnPeer(peer p2p.Peer) error {
	fmt.Printf("Doing some logic with peer: %+v\n", peer)
	return nil
}

func main() {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOHandShakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}

	t := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		StorageRoot:       "3000_Network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         *t,
	}

	s := NewFileServer(fileServerOpts)

	t.OnPeer = s.OnPeer

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
