package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestCopyEncryptDecrypt(t *testing.T) {
	payload := "foo not Bar"
	src := bytes.NewReader([]byte(payload))
	dst := new(bytes.Buffer)
	key := newEncryptionKey()

	_, err := copyEncrypt(key, src, dst)

	if err != nil {
		t.Error(err)
	}

	fmt.Println(dst.String())

	out := new(bytes.Buffer)
	nw, err := copyDecrypt(key, dst, out)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("decryted : ", out.String())

	if nw != 16+len(payload) {
		t.Error("s")
	}

	if out.String() != payload {
		t.Errorf("decryption failed")
	}
}
