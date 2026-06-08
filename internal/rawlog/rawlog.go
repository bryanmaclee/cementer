// Package rawlog is durability layer 1: an append-only file that captures every
// raw line off the pump BEFORE it is parsed or stored. Even if the protocol
// changes, the parser has a bug, or the SQLite layer fails, the untouched byte
// stream is on disk and re-importable. This is the cheapest possible insurance.
package rawlog

import (
	"bufio"
	"os"
	"sync"
	"time"
)

// Writer appends raw lines to a file with O_APPEND and periodic fsync. It is safe
// for concurrent use, though in the pipeline only the ingest goroutine writes.
type Writer struct {
	mu     sync.Mutex
	f      *os.File
	bw     *bufio.Writer
	closed bool

	stop chan struct{}
	done chan struct{}
}

// Open opens (creating if needed) the raw log at path in append mode and starts a
// background flusher that fsyncs every flushEvery.
func Open(path string, flushEvery time.Duration) (*Writer, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	if flushEvery <= 0 {
		flushEvery = time.Second
	}
	w := &Writer{
		f:    f,
		bw:   bufio.NewWriter(f),
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	go w.flushLoop(flushEvery)
	return w, nil
}

// Append writes one raw line plus a newline. It buffers; durability to disk is
// guaranteed by the periodic fsync (and the final flush on Close).
func (w *Writer) Append(line []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return os.ErrClosed
	}
	if _, err := w.bw.Write(line); err != nil {
		return err
	}
	return w.bw.WriteByte('\n')
}

func (w *Writer) flushLoop(every time.Duration) {
	defer close(w.done)
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-w.stop:
			return
		case <-t.C:
			w.sync()
		}
	}
}

func (w *Writer) sync() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return
	}
	_ = w.bw.Flush()
	_ = w.f.Sync()
}

// Close flushes, fsyncs, and closes the file.
func (w *Writer) Close() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	w.closed = true
	w.mu.Unlock()

	close(w.stop)
	<-w.done

	w.mu.Lock()
	defer w.mu.Unlock()
	_ = w.bw.Flush()
	_ = w.f.Sync()
	return w.f.Close()
}
