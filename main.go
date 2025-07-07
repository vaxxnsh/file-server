package main

import (
	"bytes"
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
		EncKey:            newEncryptionKey(),
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
	s3 := makeServer("5000", ":3000", ":4000")
	go func() {
		log.Fatal(s1.Start())
	}()

	time.Sleep(1 * time.Second)

	go func() {
		log.Fatal(s2.Start())
	}()

	time.Sleep(1 * time.Second)

	go s3.Start()

	time.Sleep(1 * time.Second)

	for i := range 1 {
		key := fmt.Sprintf("coolPicture_%d.jpg", i)
		data := bytes.NewReader([]byte("that a cool picture"))
		s3.Store(key, data)
		time.Sleep(5 * time.Millisecond)

		if err := s3.store.Delete(s3.ID, key); err != nil {
			log.Fatal(err)
		}

		r, err := s3.Get(key)
		if err != nil {
			log.Fatal(err)
		}

		b, err := io.ReadAll(r)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Gotten bytes are : %s\n", string(b))
		time.Sleep(1 * time.Second)
	}
	select {}
}
