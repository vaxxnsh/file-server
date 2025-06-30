package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/vaxxnsh/file-server/p2p"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.TCPTransport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts
	store    *Store
	peerLock sync.Mutex
	peers    map[string]p2p.Peer
	quitch   chan struct{}
}

type Payload struct {
	Key  string
	Data []byte
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storageOpts := StoreOpts{
		PathTransformFunc: opts.PathTransformFunc,
		Root:              opts.StorageRoot,
	}

	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storageOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("File server stopped")
		s.Transport.Close()
	}()

	for {
		select {
		case msg := <-s.Transport.Consume():
			fmt.Println(msg)
		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}

		go func(addr string) {
			fmt.Printf("[%s] attemping to connect with remote %s\n", s.Transport.ListenAddr, addr)
			if err := s.Transport.Dial(addr); err != nil {
				log.Println("dial error: ", err)
			}
		}(addr)
	}

	return nil
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p

	return nil
}

// func (s *FileServer) broadcast(p Payload) {

// 	for _,perr := range s.peers {

// 	}

// }

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.loop()

	return nil
}

func (s *FileServer) Stop() {
	close(s.quitch)
}
