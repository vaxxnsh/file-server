package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/vaxxnsh/file-server/p2p"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         *p2p.TCPTransport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts
	store    *Store
	peerLock sync.Mutex
	peers    map[string]p2p.Peer
	quitch   chan struct{}
}
type Message struct {
	From    string
	Payload any
}

type MessageStoreFile struct {
	Key  string
	Size int64
}

type MessageGetFile struct {
	Key string
}

func (s *FileServer) handleMessage(from string, msg *Message) error {

	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.handleMessageGetFile(from, v)
	default:
		fmt.Printf("unknown message type: %T\n", v)
	}
	return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	if !s.store.Has(msg.Key) {
		return fmt.Errorf("file (%s) does not exist on disk", msg.Key)
	}
	r, err := s.store.Read(msg.Key)

	if err != nil {
		return err
	}

	peer, ok := s.peers[from]

	if !ok {
		return fmt.Errorf("peer %s not in the peer map", from)
	}

	n, err := io.Copy(peer, r)

	if err != nil {
		return err
	}

	fmt.Printf("wrote %d bytes over the network to %s\n", n, from)
	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]

	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer map", from)
	}
	n, err := s.store.Write(msg.Key, io.LimitReader(peer, int64(msg.Size)))
	if err != nil {
		return err
	}
	peer.(*p2p.TCPPeer).Wg.Done()
	log.Printf("Wrote %d bytes to disk", n)
	return nil
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.Has(key) {
		return s.store.Read(key)
	}

	fmt.Printf("don't have file locally fetching from network")

	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}

	for _, peer := range s.peers {
		fmt.Println("recving stream from peer : ", peer)
		fileBuffer := new(bytes.Buffer)
		n, err := io.CopyN(fileBuffer, peer, 22)
		if err != nil {
			return nil, err
		}
		fmt.Println("received bytes over the network: ", n)
		fmt.Println(fileBuffer.String())
	}

	return nil, nil
}

func (s *FileServer) Store(key string, r io.Reader) error {

	fileBuffer := new(bytes.Buffer)

	tee := io.TeeReader(r, fileBuffer)

	size, err := s.store.Write(key, tee)

	if err != nil {
		return err
	}

	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return err
	}

	time.Sleep(3 * time.Second)

	for _, peer := range s.peers {
		n, err := io.Copy(peer, fileBuffer)
		if err != nil {
			return err
		}
		fmt.Printf("recieved and written %d bytes", n)
	}

	return nil

	// return s.broadcast(&Message{
	// 	From:    "todo",
	// 	Payload: p,
	// })
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("File server stopped")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("decoding error : ", err)
			}

			if err := s.handleMessage(rpc.From.String(), &msg); err != nil {
				log.Println("handle message err : ", err)
			}
		case <-s.quitch:
			return
		}
	}
}

// func (s *FileServer) stream(msg *Message) error {
// 	peers := []io.Writer{}
// 	for _, peer := range s.peers {
// 		peers = append(peers, peer)
// 	}
// 	mw := io.MultiWriter(peers...)

// 	return gob.NewEncoder(mw).Encode(msg)
// }

func (s *FileServer) broadcast(msg *Message) error {
	msgBuf := new(bytes.Buffer)

	if err := gob.NewEncoder(msgBuf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		if err := peer.Send(msgBuf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	if len(s.BootstrapNodes) != 0 {
		s.bootstrapNetwork()
	}

	s.loop()

	return nil
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p

	log.Printf("connected with remote : %s", p.RemoteAddr())

	return nil
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

func init() {
	gob.Register(MessageGetFile{})
	gob.Register(MessageStoreFile{})
}
