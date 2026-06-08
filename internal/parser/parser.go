// Package parser turns a raw ASCII line off the pump into a model.Reading.
//
// This is the ONE place the on-the-wire protocol is encoded. The exact field
// layout from the real cement unit is not yet known, so the parser is deliberately
// permissive: it never panics, skips blank/comment/garbage lines, and tolerates
// short lines and unparseable fields. Once a real capture exists, refine Config.
package parser

import (
	"strconv"
	"strings"
	"time"

	"github.com/bryanmaclee/cementer/internal/model"
)

// Config describes how to map positional ASCII fields to named channels.
type Config struct {
	// Delimiter splits a line into fields. Default ",".
	Delimiter string
	// Channels maps field index -> channel name. A field whose name is "" is
	// ignored (e.g. a leading timestamp column we don't use yet).
	Channels []string
}

// DefaultConfig is a sensible starting layout for development against the
// synthetic stream: comma-separated pressure,rate,density,volume.
func DefaultConfig() Config {
	return Config{
		Delimiter: ",",
		Channels:  []string{"pressure", "rate", "density", "volume"},
	}
}

// Parser is safe for use from a single ingest goroutine.
type Parser struct {
	cfg Config
	seq int64
}

func New(cfg Config) *Parser {
	if cfg.Delimiter == "" {
		cfg.Delimiter = ","
	}
	if len(cfg.Channels) == 0 {
		cfg.Channels = DefaultConfig().Channels
	}
	return &Parser{cfg: cfg}
}

// Parse converts one raw line into a Reading stamped with ts. ok is false for
// lines that carry no usable numeric data (blank, comment, or all-garbage) — the
// caller should skip those for storage/broadcast but they are already safe in the
// raw log regardless.
func (p *Parser) Parse(line []byte, ts time.Time) (model.Reading, bool) {
	s := strings.TrimSpace(string(line))
	if s == "" || strings.HasPrefix(s, "#") {
		return model.Reading{}, false
	}

	fields := strings.Split(s, p.cfg.Delimiter)
	values := make(map[string]float64, len(p.cfg.Channels))
	for i, raw := range fields {
		if i >= len(p.cfg.Channels) {
			break
		}
		name := p.cfg.Channels[i]
		if name == "" {
			continue
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
		if err != nil {
			// Tolerate an unparseable field; just omit that channel for this frame.
			continue
		}
		values[name] = v
	}

	if len(values) == 0 {
		return model.Reading{}, false
	}

	p.seq++
	return model.Reading{Seq: p.seq, TS: ts, Values: values}, true
}
