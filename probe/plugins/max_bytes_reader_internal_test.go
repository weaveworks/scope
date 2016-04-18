package plugins

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strings"
	"testing"
)

func TestMaxBytesReaderReturnsAllDataIfSmaller(t *testing.T) {
	result, err := ioutil.ReadAll(MaxBytesReader(ioutil.NopCloser(strings.NewReader("some data")), 1024, errors.New("test error")))
	if err != nil {
		t.Error(err)
	}
	if string(result) != "some data" {
		t.Errorf("Expected %q, got %q", "some data", string(result))
	}
}

func TestMaxBytesReaderReturnsNilIfNil(t *testing.T) {
	result := MaxBytesReader(nil, 1024, errors.New("test error"))
	if result != nil {
		t.Errorf("Expected nil, got: %q", result)
	}
}

func TestMaxBytesReaderReturnsErrorIfLarger(t *testing.T) {
	input := &bytes.Buffer{}
	for i := int64(0); i <= 1024; i++ {
		input.WriteByte(byte(i))
	}

	result, err := ioutil.ReadAll(MaxBytesReader(ioutil.NopCloser(input), 1024, errors.New("test error")))
	if err.Error() != "test error" {
		t.Errorf("Expected error to be %q, got: %q", "test error", err.Error())
	}
	if len(result) != 1024 {
		t.Errorf("Expected result length to be 1024, but got: %d", len(result))
	}
}

func TestMaxBytesReaderReturnsErrorIfLargerAndMassiveBufferGiven(t *testing.T) {
	input := &bytes.Buffer{}
	for i := int64(0); i <= 1024; i++ {
		input.WriteByte(byte(i))
	}

	buffer := make([]byte, 1024+2)
	reader := MaxBytesReader(ioutil.NopCloser(input), 1024, errors.New("test error"))

	// First read is scoped down to the maximum
	readCount, err := reader.Read(buffer)
	if err != nil {
		t.Error(err)
	}
	if readCount != 1024 {
		t.Errorf("Expected result length to be 1024, but got: %d", readCount)
	}

	// Second read returns an error
	readCount, err = reader.Read(buffer)
	if err.Error() != "test error" {
		t.Errorf("Expected error to be %q, got: %q", "test error", err.Error())
	}
	if readCount != 0 {
		t.Errorf("Expected result length to be 0, but got: %d", readCount)
	}
}

type testReadCloser struct {
	closeError error
}

func (c testReadCloser) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (c testReadCloser) Close() error {
	return c.closeError
}

func TestMaxBytesReaderPassesThroughErrorsWhenClosing(t *testing.T) {
	readcloser := testReadCloser{errors.New("test error")}
	err := MaxBytesReader(readcloser, 1024, errors.New("overflow")).Close()
	if err == nil || err.Error() != "test error" {
		t.Errorf("Expected error to be %q, got: %q", "test error", err)
	}
}
