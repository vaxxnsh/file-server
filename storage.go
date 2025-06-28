package main

import (
	"io"
	"log"
	"os"
)

type PathTransformFunc func(key string) string

var DefaultPathTransformFunc = func(key string) string {
	return key
}

type StoreOpts struct {
	PathTransformFunc PathTransformFunc
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathname := s.PathTransformFunc(key)

	if err := os.MkdirAll(pathname, os.ModePerm); err != nil {
		return err
	}

	filename := "somefilename"
	pathnameWithFilename := pathname + "/" + filename

	f, err := os.Create(pathnameWithFilename)

	if err != nil {
		return err
	}
	n, err := io.Copy(f, r)

	if err != nil {
		return err
	}
	log.Printf("Wrote %d bytes to %s", n, filename)
	return nil
}
