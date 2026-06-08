// Package serialreader is the production source.LineSource: it reads newline-
// delimited ASCII off a USB-serial port (the RS-232→USB adapter on the Pi) using
// go.bug.st/serial. It implements the same interface as the replay source, so the
// rest of the pipeline never knows whether data came from the pump or a file.
package serialreader

import (
	"bufio"
	"context"

	"go.bug.st/serial"
)

// Config is the serial port configuration. Port is a device path; prefer a stable
// /dev/serial/by-id/... path over /dev/ttyUSB0 so it survives replug/reboot.
type Config struct {
	Port     string
	BaudRate int
	DataBits int
	Parity   serial.Parity
	StopBits serial.StopBits
}

// DefaultConfig is 9600 8N1 — a common default. Override once the pump's settings
// are known.
func DefaultConfig(port string) Config {
	return Config{
		Port:     port,
		BaudRate: 9600,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
}

type Reader struct {
	cfg  Config
	port serial.Port
}

// Open opens the serial port with the given configuration.
func Open(cfg Config) (*Reader, error) {
	mode := &serial.Mode{
		BaudRate: cfg.BaudRate,
		DataBits: cfg.DataBits,
		Parity:   cfg.Parity,
		StopBits: cfg.StopBits,
	}
	port, err := serial.Open(cfg.Port, mode)
	if err != nil {
		return nil, err
	}
	return &Reader{cfg: cfg, port: port}, nil
}

// Run reads lines until ctx is cancelled or the port errors. A blocking
// bufio.Scanner read sits off the event path entirely, so a slow client can never
// stall serial reads (it's a dedicated goroutine in the pipeline).
func (r *Reader) Run(ctx context.Context, emit func(line []byte)) error {
	// Close the port when the context is cancelled so the blocking Scan unblocks.
	go func() {
		<-ctx.Done()
		_ = r.port.Close()
	}()

	sc := bufio.NewScanner(r.port)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		emit(append([]byte(nil), sc.Bytes()...))
	}
	if err := sc.Err(); err != nil && ctx.Err() == nil {
		return err
	}
	return ctx.Err()
}

func (r *Reader) Close() error {
	return r.port.Close()
}
