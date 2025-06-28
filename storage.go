package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
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
		Original: hashHex,
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
	Original string
}

func NewStore(opts StoreOpts) *Store {
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathname := s.PathTransformFunc(key)

	if err := os.MkdirAll(pathname.Pathname, os.ModePerm); err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	io.Copy(buf, r)

	filenameBytes := md5.Sum(buf.Bytes())
	filename := hex.EncodeToString(filenameBytes[:])
	pathnameWithFilename := pathname.Pathname + "/" + filename

	f, err := os.Create(pathnameWithFilename)

	if err != nil {
		return err
	}
	n, err := io.Copy(f, buf)

	if err != nil {
		return err
	}
	log.Printf("Wrote %d bytes to %s", n, filename)
	return nil
}
