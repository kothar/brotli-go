# Go bindings for the Brotli compression library

See <https://github.com/google/brotli> for the upstream C/C++ source, and
the `VERSION.md` file to find out the currently vendored version.

Usage
---

To use the bindings, you just need to import the enc or dec package and call the Go wrapper
functions `enc.CompressBuffer` or `dec.DecompressBuffer`

```go
import (
	"gopkg.in/kothar/brotli-go.v0/dec"
	"gopkg.in/kothar/brotli-go.v0/enc"
)
```

From the tests:

```go
import (
	"bytes"
	"testing"

	"gopkg.in/kothar/brotli-go.v0/dec"
	"gopkg.in/kothar/brotli-go.v0/enc"
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
```

Advanced usage (streaming API)
---

When the data set is too large to fit in-memory, `CompressBuffer` and
`DecompressBuffer` are not a viable option.

`brotli-go` also exposes a streaming interface both for encoding:

```go
import (
	"gopkg.in/kothar/brotli-go.v0/enc"
)

func main() {
  compressedWriter := os.OpenFile("data.bin.bro", os.O_CREATE|os.O_WRONLY, 0644)

  // passing nil to get default params — careful, q=11 is the (extremely slow) default
  brotliWriter := enc.NewBrotliWriter(nil, compressedWriter)
  // BrotliWriter will close writer passed as argument if it implements io.Closer
  defer brotliWriter.Close()

  fileReader, _ := os.Open("data.bin")
  defer fileReader.Close()

  io.Copy(fileReader, brotliWriter)
}
```

..and for decoding:

```go
import (
	"gopkg.in/kothar/brotli-go.v0/dec"
)

func main() {
  archiveReader, _ := os.Open("data.bin.bro")

  brotliReader := dec.NewBrotliReader(archiveReader)
  defer brotliReader.Close()

  decompressedWriter := os.OpenFile("data.bin.unbro", os.O_CREATE|os.O_WRONLY, 0644)
  defer decompressedWriter.Close()
  io.Copy(brotliReader, decompressedWriter)
}
```

Bindings
---

This is a very basic Cgo wrapper for the enc and dec directories from the Brotli sources. I've made a few minor changes to get
things working with Go.

1. The default dictionary has been extracted to a separate 'shared' package to allow linking the enc and dec cgo modules if you use both. Otherwise there are duplicate symbols, as described in the dictionary.h header files.

2. The dictionary variable name for the dec package has been modified for the same reason, to avoid linker collisions.

Links
---

  * brotli streaming decompression written in pure go: <https://github.com/dsnet/compress>

License
---

Brotli and these bindings are open-sourced under the Apache License, Version 2.0, see the LICENSE file.
