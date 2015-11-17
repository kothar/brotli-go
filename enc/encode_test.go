package enc

import (
	"bytes"
	"testing"
)

func TestBufferSizes(T *testing.T) {

	input1 := bytes.Repeat([]byte("The quick brown fox jumps over the lazy dog. "), 100000)
	output1 := make([]byte, len(input1)*2)
	_, err := CompressBuffer(nil, input1, output1)
	if err != nil {
		T.Error(err)
	}

	output2 := make([]byte, len(input1))
	_, err = CompressBuffer(nil, input1, output2)
	if err != nil {
		T.Error(err)
	}

	output3 := make([]byte, len(input1)/2)
	_, err = CompressBuffer(nil, input1, output3)
	if err != nil {
		T.Error(err)
	}

	_, err = CompressBuffer(nil, input1, nil)
	if err != nil {
		T.Error(err)
	}
}
