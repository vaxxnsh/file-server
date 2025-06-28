package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "someGoodPicture"
	pathname := CASPathTransformFunc(key)
	expected := "6f96b/0131b/a7047/77da4/41184/6af2c/c0e8e/fddd7"

	if pathname.Pathname != expected {
		t.Errorf("want %s have %s", expected, pathname)
	}
}

func TestStore(t *testing.T) {
	store := NewStore(StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	})
	bytes := bytes.NewReader([]byte("some jpeg file content"))

	if err := store.writeStream("pictures", bytes); err != nil {
		t.Error(err)
	}

	r, err := store.Read("pictures")
	if err != nil {
		t.Error(err)
	}
	b, _ := io.ReadAll(r)

	fmt.Println(string(b))

	if string(b) != "some jpeg file content" {
		t.Errorf("want %s have %s", "some jpeg file content", b)
	}
}
