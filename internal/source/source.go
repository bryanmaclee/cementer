// Package source abstracts where raw ASCII lines come from, so the rest of the
// pipeline is identical whether data arrives from the real USB-serial port or
// from a replay file during development. Implementations: serialreader (production)
// and the Replay source below (development / tests, no pump required).
package source

import (
	"bufio"
	"context"
	"io"
	"os"
	"time"
)

// LineSource produces raw lines (newline stripped). Run blocks, delivering each
// line to emit, until the context is cancelled (live sources) or the input is
// exhausted (finite sources like a replay file), then returns.
type LineSource interface {
	Run(ctx context.Context, emit func(line []byte)) error
	Close() error
}

// Replay emits the lines of an io.Reader at a fixed cadence, simulating a live
// stream. With Loop set it restarts from the top when the input is exhausted, so a
// short sample file can drive an indefinitely long dev session.
type Replay struct {
	rc       io.ReadCloser
	interval time.Duration
	loop     bool
	reopen   func() (io.ReadCloser, error) // used to rewind when Loop is set
}

// NewReplayFile opens path and emits one line every interval. If loop is true it
// rewinds at EOF (so the file can be re-read forever).
func NewReplayFile(path string, interval time.Duration, loop bool) (*Replay, error) {
	open := func() (io.ReadCloser, error) { return os.Open(path) }
	rc, err := open()
	if err != nil {
		return nil, err
	}
	if interval <= 0 {
		interval = time.Second
	}
	return &Replay{rc: rc, interval: interval, loop: loop, reopen: open}, nil
}

// NewReplayReader emits lines from rc at interval (no looping).
func NewReplayReader(rc io.ReadCloser, interval time.Duration) *Replay {
	if interval <= 0 {
		interval = time.Second
	}
	return &Replay{rc: rc, interval: interval}
}

func (r *Replay) Run(ctx context.Context, emit func(line []byte)) error {
	t := time.NewTicker(r.interval)
	defer t.Stop()

	sc := bufio.NewScanner(r.rc)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for {
		if !sc.Scan() {
			if err := sc.Err(); err != nil {
				return err
			}
			if !r.loop || r.reopen == nil {
				return nil // finite input exhausted
			}
			// Rewind: close current and reopen for another pass.
			_ = r.rc.Close()
			nrc, err := r.reopen()
			if err != nil {
				return err
			}
			r.rc = nrc
			sc = bufio.NewScanner(r.rc)
			sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
			continue
		}

		// Copy the line: Scanner reuses its buffer on the next Scan.
		line := append([]byte(nil), sc.Bytes()...)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			emit(line)
		}
	}
}

func (r *Replay) Close() error {
	if r.rc != nil {
		return r.rc.Close()
	}
	return nil
}
