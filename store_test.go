package main

import (
	"bytes"
	"testing"
)

func TestStore(t *testing.T) {
	store := NewStore(StoreOpts{
		PathTransformFunc: DefaultPathTransformFunc,
	})
	bytes := bytes.NewReader([]byte("some jpeg file content"))

	if err := store.writeStream("test.txt", bytes); err != nil {
		t.Error(err)
	}
}
