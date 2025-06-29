package main

import (
	"fmt"
	"log"

	"github.com/vaxxnsh/file-server/p2p"
)

type FileServerOpts struct {
	ListenAddr        string
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.TCPTransport
	TransportOpts     p2p.TCPTransportOpts
}

type FileServer struct {
	FileServerOpts
	store  *Store
	quitch chan struct{}
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
	}
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("File server stopped")
		s.Stop()
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
