# Go bindings for the Brotli compression library

[![GoDoc](https://godoc.org/gopkg.in/kothar/brotli-go.v0?status.svg)](https://godoc.org/gopkg.in/kothar/brotli-go.v0)
[![Build Status](https://travis-ci.org/kothar/brotli-go.svg)](https://travis-ci.org/kothar/brotli-go)

See <https://github.com/google/brotli> for the upstream C/C++ source, and
the `VERSION.md` file to find out the currently vendored version.

Usage
---

To use the bindings, you just need to import the enc or dec package and call the Go wrapper
functions `enc.CompressBuffer` or `dec.DecompressBuffer`

Naive compression + decompression example with no error handling:

```go
import (
	"gopkg.in/kothar/brotli-go.v0/dec"
	"gopkg.in/kothar/brotli-go.v0/enc"
)

func brotliRoundtrip(input []byte) []byte {
  // passing nil to get default *BrotliParams
  // careful, q=11 is the (extremely slow) default
  compressed, _ := enc.CompressBuffer(nil, input, make([]byte, 0))
  decompressed, _ := dec.DecompressBuffer(compressed, make([]byte, 0))
  return decompressed
}
```

For a more complete roundtrip example, read top-level file `brotli_test.go`

The `enc.BrotliParams` type lets you specify various Brotli parameters, such
as `quality`, `lgwin` (sliding window size), and `lgblock` (input block size).

```go
import (
	"gopkg.in/kothar/brotli-go.v0/enc"
)

func brotliFastCompress(input []byte) []byte {
  params := enc.NewBrotliParams()
  // brotli supports quality values from 0 to 11 included
  // 0 is the fastest, 11 is the most compressed but slowest
  params.SetQuality(0)
  compressed, _ := enc.CompressBuffer(params, input, make([]byte, 0))
  return compressed
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
  compressedWriter,_ := os.OpenFile("data.bin.bro", os.O_CREATE|os.O_WRONLY, 0644)

  brotliWriter := enc.NewBrotliWriter(nil, compressedWriter)
  // BrotliWriter will close writer passed as argument if it implements io.Closer
  defer brotliWriter.Close()

  fileReader, _ := os.Open("data.bin")
  defer fileReader.Close()

  io.Copy(brotliWriter,fileReader)
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

  decompressedWriter,_ := os.OpenFile("data.bin.unbro", os.O_CREATE|os.O_WRONLY, 0644)
  defer decompressedWriter.Close()
  io.Copy(decompressedWriter, brotliReader)
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

Brotli and these bindings are open-sourced under the MIT License - see the LICENSE file.
