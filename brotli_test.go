package brotli

import (
	"bytes"
	"testing"

	"github.com/kothar/brotli-go/dec"
	"github.com/kothar/brotli-go/enc"
)

func TestRoundtrip(T *testing.T) {
	T.Log("Compressing test string")
	s := []byte("Hello Hello Hello, Hello Hello Hello")
	T.Logf("Original: %s\n", s)

	params := enc.NewBrotliParams()
	buffer1 := make([]byte, len(s)+100)
	encoded, cerr := enc.CompressBuffer(params, s, buffer1)
	if cerr != nil {
		T.Error(cerr)
	}
	T.Logf("Compressed: %v\n", encoded)

	buffer2 := make([]byte, len(s))
	decoded, derr := dec.DecompressBuffer(encoded, buffer2)
	if derr != nil {
		T.Error(derr)
	}
	T.Logf("Decompressed: %s\n", decoded)

	if !bytes.Equal(s, decoded) {
		T.Error("Decoded output does not match original input")
	} else {
		T.Log("Decoded output matches original input")
	}
}
