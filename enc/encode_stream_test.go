package enc

import (
	"log"
	"strings"
	"testing"
)

func TestStreamEncode(T *testing.T) {
	input1 := []byte(strings.Repeat("The quick brown fox jumps over the lazy dog", 1000))
	inputSize := int64(len(input1))

	compressor := NewBrotliCompressor(nil)
	blockSize := compressor.GetInputBlockSize()

	fullOutput := make([]byte, 0)

	pos := int64(0)
	for pos < inputSize {
		copySize := blockSize
		remaining := inputSize - pos
		if remaining < copySize {
			copySize = remaining
		}
		compressor.CopyInputToRingBuffer(input1[pos : pos+copySize])
		pos += copySize

		output := compressor.WriteBrotliData(pos >= inputSize, false)
		fullOutput = append(fullOutput, output...)
	}

	fullOutput2, _ := CompressBuffer(nil, input1, make([]byte, 0))
	log.Println("input size = ", inputSize, "output size 1 = ", len(fullOutput), "output size 2 = ", len(fullOutput2))
}
