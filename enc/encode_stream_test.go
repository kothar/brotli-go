package enc

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestStreamEncode(T *testing.T) {
	input1 := []byte(strings.Repeat("The quick brown fox jumps over the lazy dog", 100000))
	inputSize := int64(len(input1))
	log.Printf("inputSize=%d\n", inputSize)
	params := NewBrotliParams()

	for lgwin := 16; lgwin <= 22; lgwin += 1 {
		params.SetLgwin(lgwin)
		compressor := NewBrotliCompressor(params)
		blockSize := compressor.GetInputBlockSize()

		fullOutput := make([]byte, 0)

		rounds := 0
		pos := int64(0)
		for pos < inputSize {
			rounds++
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

		fullOutput2, _ := CompressBuffer(params, input1, make([]byte, 0))

		if !bytes.Equal(fullOutput, fullOutput2) {
			T.Fatal("for lgwin %d, stream compression didn't give same result as buffer compression", params.Lgwin())
		}

		outputSize := len(fullOutput)
		log.Printf("lgwin=%d, rounds=%d, output=%d (%.4f%% of input size)\n", params.Lgwin(), rounds, outputSize, float32(outputSize)*100.0/float32(inputSize))
	}
}
