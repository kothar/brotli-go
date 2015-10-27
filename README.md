Go bindings for the Brotli compression library
===

See https://github.com/google/brotli for the upstream C/C++ source

Usage
---

To use the bindings, you just need to import the enc or dec package and call the Go wrapper 
functions `enc.CompressBuffer` or `dec.DecompressBuffer`

From the tests:
~~~
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
~~~

Bindings
---

This is a very basic Cgo wrapper for the enc and dec directories from the Brotli sources. I've made a few minor changes to get
things working with Go.

1. The default dictionary has been extracted to a separate 'shared' package to allow linking the enc and dec cgo modules if you use both. Otherwise there are duplicate symbols, as described in the dictionary.h header files.

2. The dictonary variable name for the dec package has been modified for the same reason, to avoid linker collisions.

TODO
---

* I haven't implemented stream compression yet - it will need a wrapper for the C++ classes

License
---

Brotli and these bindings are open-sourced under the Apache License, Version 2.0, see the LICENSE file.
