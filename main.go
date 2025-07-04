package main

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/vaxxnsh/file-server/p2p"
)

func OnPeer(peer p2p.Peer) error {
	fmt.Printf("Doing some logic with peer: %+v\n", peer)
	return nil
}

func makeServer(lisetenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":" + lisetenAddr,
		HandshakeFunc: p2p.NOHandShakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}

	t := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		StorageRoot:       lisetenAddr + "_Network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         t,
		BootstrapNodes:    nodes,
	}

	s := NewFileServer(fileServerOpts)

	t.OnPeer = s.OnPeer

	return s
}

func main() {
	s1 := makeServer("3000", "")
	s2 := makeServer("4000", ":3000")
	go func() {
		log.Fatal(s1.Start())
	}()

	time.Sleep(1 * time.Second)

	go s2.Start()

	time.Sleep(1 * time.Second)

	// data := bytes.NewReader([]byte("that a cool picture"))
	// s2.Store("coolPicture.jpg", data)
	// time.Sleep(5 * time.Millisecond)

	r, err := s2.Get("coolPicture.jpg")
	if err != nil {
		log.Fatal(err)
	}

	b, err := io.ReadAll(r)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Gotten bytes are : %s\n", string(b))
	select {}
}
