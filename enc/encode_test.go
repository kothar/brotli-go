package enc

import (
	"bytes"
	"strings"
	"testing"

	"gopkg.in/kothar/brotli-go.v0/dec"
)

const (
	testQuality int = 9
)

func TestBufferSizes(T *testing.T) {
	params := NewBrotliParams()
	params.SetQuality(testQuality)

	input1 := []byte(strings.Repeat("The quick brown fox jumps over the lazy dog", 100000))
	T.Logf("q=%d, inputSize=%d\n", params.Quality(), len(input1))

	output1 := make([]byte, len(input1)*2)
	_, err := CompressBuffer(params, input1, output1)
	if err != nil {
		T.Error(err)
	}

	output2 := make([]byte, len(input1))
	_, err = CompressBuffer(params, input1, output2)
	if err != nil {
		T.Error(err)
	}

	output3 := make([]byte, len(input1)/2)
	_, err = CompressBuffer(params, input1, output3)
	if err != nil {
		T.Error(err)
	}

	_, err = CompressBuffer(params, input1, nil)
	if err != nil {
		T.Error(err)
	}
}

func TestStreamEncode(T *testing.T) {
	params := NewBrotliParams()
	params.SetQuality(testQuality)

	input1 := []byte(strings.Repeat("The quick brown fox jumps over the lazy dog", 100000))
	inputSize := len(input1)
	T.Logf("q=%d, inputSize=%d\n", params.Quality(), inputSize)

	for lgwin := 16; lgwin <= 22; lgwin++ {
		params.SetLgwin(lgwin)
		compressor := newBrotliCompressor(params)
		defer compressor.free()
		blockSize := compressor.getInputBlockSize()

		// compress the entire data in one go
		fullBufferOutput, err := CompressBuffer(params, input1, make([]byte, 0))
		if err != nil {
			T.Error(err)
		}

		// then using the low-level stream interface
		streamBuffer := new(bytes.Buffer)
		rounds := 0
		pos := 0
		for pos < inputSize {
			rounds++
			copySize := blockSize
			remaining := inputSize - pos
			if remaining < copySize {
				copySize = remaining
			}
			compressor.copyInputToRingBuffer(input1[pos : pos+copySize])
			pos += copySize

			output, err := compressor.writeBrotliData(pos >= inputSize, false)
			if err != nil {
				T.Error(err)
			}
			streamBuffer.Write(output)
		}

		fullStreamOutput := streamBuffer.Bytes()
		if !bytes.Equal(fullStreamOutput, fullBufferOutput) {
			T.Fatalf("for lgwin %d, stream compression didn't give same result as buffer compression", params.Lgwin())
		}

		// then using the high-level Writer interface
		writerBuffer := new(bytes.Buffer)
		writer := NewBrotliWriter(params, writerBuffer)
		writer.Write(input1)
		writer.Close()

		fullWriterOutput := writerBuffer.Bytes()
		if !bytes.Equal(fullWriterOutput, fullBufferOutput) {
			T.Fatalf("for lgwin %d, stream writer compression didn't give same result as buffer compression", params.Lgwin())
		}

		outputSize := len(fullStreamOutput)
		T.Logf("lgwin=%d, rounds=%d, output=%d (%.4f%% of input size)\n", params.Lgwin(), rounds, outputSize, float32(outputSize)*100.0/float32(inputSize))
	}
}

func TestFlush(T *testing.T) {
	params := NewBrotliParams()
	params.SetQuality(testQuality)

	input1 := []byte(strings.Repeat("The quick brown fox jumps over the lazy dog", 10000))
	inputSize := len(input1)

	for _, blockSize := range []int{324, 1623, 6125, 21126} {
		T.Logf("q=%d, inputSize=%d, blockSize=%d\n", params.Quality(), inputSize, blockSize)

		writerBuffer := new(bytes.Buffer)
		writer := NewBrotliWriter(params, writerBuffer)
		outputLength := 0

		// Write small blocks
		for pos := 0; pos < len(input1); pos += blockSize {
			end := pos + blockSize
			if end > len(input1) {
				end = len(input1)
			}
			if _, err := writer.Write(input1[pos:end]); err != nil {
				T.Error(err)
			}

			newOutputLength := writerBuffer.Len()
			if newOutputLength == outputLength {
				// Nothing was written, attempt to flush
				if err := writer.Flush(); err != nil {
					T.Error(err)
				}
				newOutputLength = writerBuffer.Len()
				if newOutputLength == outputLength {
					T.Error("Flush did not produce any additional output")
				}
			}
			outputLength = newOutputLength
		}
		if err := writer.Close(); err != nil {
			T.Error(err)
		}

		// Check the output is valid
		decoded, err := dec.DecompressBuffer(writerBuffer.Bytes(), nil)
		if err != nil {
			T.Error(err)
		}
		if !bytes.Equal(decoded, input1) {
			T.Errorf("Flushed output does not decode to same bytes as input")
		}
	}
}
