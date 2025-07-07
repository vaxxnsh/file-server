package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type PathTransformFunc func(key string) PathKey

const DefaultRootName = "file-data"

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
	ID                string
	Root              string
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
	if len(opts.Root) == 0 {
		opts.Root = DefaultRootName
	}

	if len(opts.ID) == 0 {
		opts.ID = generateID()
	}

	return &Store{
		StoreOpts: opts,
	}
}

func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.Pathname, p.FileName)
}

func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.Pathname, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Delete(id, key string) error {
	pathKey := s.PathTransformFunc(key)

	defer func() {
		log.Printf("deleted [%s] from disk", pathKey.FileName)
	}()

	firstPathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FirstPathName())

	return os.RemoveAll(firstPathNameWithRoot)
}

func (s *Store) Has(id, key string) bool {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())

	_, err := os.Stat(fullPathWithRoot)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Read(id, key string) (int64, io.Reader, error) {
	return s.readStream(id, key)
}

func (s *Store) readStream(id, key string) (int64, io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())

	file, err := os.Open(fullPathWithRoot)

	if err != nil {
		return 0, nil, err
	}

	fi, err := file.Stat()

	if err != nil {
		return 0, nil, err
	}

	return fi.Size(), file, nil
}

func (s *Store) Write(id, key string, r io.Reader) (int64, error) {
	return s.writeStream(id, key, r)
}

func (s *Store) WriteDecrypt(encKey []byte, id, key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(id, key)

	if err != nil {
		return 0, err
	}
	n, err := copyDecrypt(encKey, r, f)

	if err != nil {
		return 0, err
	}

	return int64(n), nil
}

func (s *Store) openFileForWriting(id string, key string) (*os.File, error) {
	pathKey := s.PathTransformFunc(key)
	pathKeyWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.Pathname)
	if err := os.MkdirAll(pathKeyWithRoot, os.ModePerm); err != nil {
		return nil, err
	}

	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())

	return os.Create(fullPathWithRoot)
}

func (s *Store) writeStream(id, key string, r io.Reader) (int64, error) {

	f, err := s.openFileForWriting(id, key)

	if err != nil {
		return 0, err
	}

	return io.Copy(f, r)
}
