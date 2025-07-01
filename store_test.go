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
	s := NewStore(StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	})
	// id := generateID()
	// id := "randomId"
	// defer teardown(t, s)

	for i := range 1 {
		key := fmt.Sprintf("foo_%d", i)
		data := []byte("some jpg bytes")

		if err := s.Write(key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}

		if ok := s.Has(key); !ok {
			t.Errorf("expected to have key %s", key)
		}

		r, err := s.Read(key)
		if err != nil {
			t.Error(err)
		}

		b, _ := io.ReadAll(r)
		if string(b) != string(data) {
			t.Errorf("want %s have %s", data, b)
		}

		if err := s.Delete(key); err != nil {
			t.Error(err)
		}
	}
}
