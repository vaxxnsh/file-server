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
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOHandShakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}
	tr := p2p.NewTCPTransport(tcpOpts)

	go func() {
		msg := <-tr.Consume()
		fmt.Printf("Received message: %+v\n", msg)
	}()

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}
	select {}
}
