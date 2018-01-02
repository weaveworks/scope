package kubernetes

import (
	"bytes"
	"io"

	log "github.com/Sirupsen/logrus"
)

const (
	internalBufferSize = 1024
)

type logReadCloser struct {
	readClosers  []io.ReadCloser
	eof          []bool
	buffer       bytes.Buffer
	dataChannel  chan []byte
	stopChannels []chan struct{}
	eofChannel   chan int
}

// NewLogReadCloser takes multiple io.ReadCloser and reads where data is available.
func NewLogReadCloser(readClosers ...io.ReadCloser) io.ReadCloser {
	stopChannels := make([]chan struct{}, len(readClosers))
	for i := range readClosers {
		stopChannels[i] = make(chan struct{})
	}

	l := logReadCloser{
		readClosers:  readClosers,
		dataChannel:  make(chan []byte),
		stopChannels: stopChannels,
		eofChannel:   make(chan int),
		eof:          make([]bool, len(readClosers)),
	}

	for idx := range l.readClosers {
		go l.readInput(idx)
	}

	return &l
}

func (l *logReadCloser) Read(p []byte) (int, error) {
	if len(p) <= l.buffer.Len() {
		return l.readInternalBuffer(p)
	}

	// if there's data available to read, read it,
	// otherwise block
	byteCount := 0
	if l.buffer.Len() > 0 {
		n, err := l.readInternalBuffer(p)
		if err != nil {
			return n, err
		}
		byteCount += n
	} else {
		// block on read or EOF
		received := false
		for !received && !l.isEOF() {
			select {
			case data := <-l.dataChannel:
				l.buffer.Write(data)
				received = true
			case idx := <-l.eofChannel:
				l.eof[idx] = true
			}
		}
	}

	// check if there's more data to read, without blocking
	empty := false
	for !empty && l.buffer.Len() < len(p) {
		select {
		case data := <-l.dataChannel:
			l.buffer.Write(data)
		case idx := <-l.eofChannel:
			l.eof[idx] = true
		default:
			empty = true
		}
	}

	return l.readInternalBuffer(p[byteCount:])
}

func (l *logReadCloser) Close() error {
	for i, rc := range l.readClosers {
		err := rc.Close()
		if err != nil {
			return err
		}

		// synchronous stop:
		// the routines write to dataChannel which will be closed by this thread
		select {
		case <-l.stopChannels[i]:
			break
		}
		close(l.stopChannels[i])
	}

	close(l.dataChannel)
	close(l.eofChannel)
	return nil
}

func (l *logReadCloser) readInternalBuffer(p []byte) (int, error) {
	n, err := l.buffer.Read(p)
	if err == io.EOF && !l.isEOF() {
		return n, nil
	}

	return n, err
}

func (l *logReadCloser) readInput(idx int) {
	tmpBuffer := make([]byte, internalBufferSize)
	for {
		n, err := l.readClosers[idx].Read(tmpBuffer)
		if err == io.EOF {
			if n > 0 {
				l.dataChannel <- tmpBuffer[:n]
			}
			l.eofChannel <- idx
			break
		}
		if err != nil {
			log.Errorf("Failed to read: %v", err)
			break
		}
		l.dataChannel <- tmpBuffer[:n]
	}

	// signal the routine won't write to dataChannel
	l.stopChannels[idx] <- struct{}{}
}

func (l *logReadCloser) isEOF() bool {
	for _, e := range l.eof {
		if !e {
			return false
		}
	}
	return true
}
