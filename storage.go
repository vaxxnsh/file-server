package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type PathTransformFunc func(key string) PathKey

func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashHex := hex.EncodeToString(hash[:])
	blockSize := 5
	sliceLen := len(hashHex) / blockSize

	paths := make([]string, sliceLen)

	for i := range sliceLen {
		start, end := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashHex[start:end]
		paths[i] = hashHex[start:end]
	}

	return PathKey{
		FileName: hashHex,
		Pathname: strings.Join(paths, "/"),
	}
}

var DefaultPathTransformFunc = func(key string) string {
	return key
}

type StoreOpts struct {
	PathTransformFunc PathTransformFunc
}

type Store struct {
	StoreOpts
}

type PathKey struct {
	Pathname string
	FileName string
}

func NewStore(opts StoreOpts) *Store {
	return &Store{
		StoreOpts: opts,
	}
}

func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.Pathname, p.FileName)
}

func (s *Store) Read(key string) (io.Reader, error) {
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)

	return buf, err
}

func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)

	return os.Open(pathKey.FullPath())
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathKey := s.PathTransformFunc(key)

	if err := os.MkdirAll(pathKey.Pathname, os.ModePerm); err != nil {
		return err
	}

	filepath := pathKey.FullPath()

	f, err := os.Create(filepath)

	if err != nil {
		return err
	}
	n, err := io.Copy(f, r)

	if err != nil {
		return err
	}
	log.Printf("Wrote %d bytes to %s", n, filepath)
	return nil
}
