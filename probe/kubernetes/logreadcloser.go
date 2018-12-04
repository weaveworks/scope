package kubernetes

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"sync"
)

type logReadCloser struct {
	labels      []string
	labelLength int
	readClosers []io.ReadCloser
	eof         []bool
	buffer      bytes.Buffer
	dataChannel chan []byte
	eofChannel  chan int
	wg          sync.WaitGroup
}

// NewLogReadCloser reads from multiple io.ReadCloser, where data is available,
// and annotates each line with the reader's label
func NewLogReadCloser(readClosersWithLabel map[io.ReadCloser]string) io.ReadCloser {
	n := len(readClosersWithLabel)
	labels := make([]string, n)
	readClosers := make([]io.ReadCloser, n)

	i := 0
	labelLength := 0
	for readCloser, label := range readClosersWithLabel {
		readClosers[i] = readCloser
		labels[i] = label
		labelLength = int(math.Max(float64(labelLength), float64(len(label))))
		i++
	}

	l := logReadCloser{
		readClosers: readClosers,
		labels:      labels,
		labelLength: labelLength,
		dataChannel: make(chan []byte),
		eofChannel:  make(chan int),
		eof:         make([]bool, len(readClosers)),
	}

	l.wg.Add(n)

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
	for !empty && l.buffer.Len() < len(p) && !l.isEOF() {
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
	for _, rc := range l.readClosers {
		err := rc.Close()
		if err != nil {
			return err
		}
	}

	l.wg.Wait()

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
	defer l.wg.Done()
	reader := bufio.NewReader(l.readClosers[idx])
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			if len(line) > 0 {
				l.dataChannel <- l.annotateLine(idx, line)
			}
			break
		}
		if err != nil {
			// error, exit
			break
		}
		l.dataChannel <- l.annotateLine(idx, line)
	}

	l.eofChannel <- idx
}

func (l *logReadCloser) annotateLine(idx int, line []byte) []byte {
	// do not annotate if it's the only reader
	if len(l.labels) == 1 {
		return line
	}
	return []byte(fmt.Sprintf("[%-*s] %v", l.labelLength, l.labels[idx], string(line)))
}

func (l *logReadCloser) isEOF() bool {
	for _, e := range l.eof {
		if !e {
			return false
		}
	}
	return true
}
