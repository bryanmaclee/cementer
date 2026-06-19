// Package store is durability layer 2: the canonical, queryable SQLite store.
//
// All writes funnel through a single writer goroutine that batch-commits in WAL
// mode — this serializes writes (no "database is locked"), and the COMMIT is the
// durability point. After each batch commits, the readings in it are handed to
// onCommit (which the pipeline wires to the hub), so what live clients see always
// equals what is already stored. Clients are never in this write path.
package store

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/bryanmaclee/cementer/internal/model"

	_ "modernc.org/sqlite"
)

const maxBatch = 512

// Store owns the SQLite database and its single writer goroutine.
type Store struct {
	db            *sql.DB
	in            chan model.Reading
	onCommit      func(model.Reading)
	batchInterval time.Duration

	wg     sync.WaitGroup
	closed chan struct{}
}

// Open opens (creating if needed) the SQLite database at path with WAL mode and
// starts the writer goroutine. onCommit is called once per reading AFTER its batch
// is durably committed; it may be nil.
func Open(path string, batchInterval time.Duration, onCommit func(model.Reading)) (*Store, error) {
	dsn := fmt.Sprintf(
		"file:%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)",
		path,
	)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// One connection = a single serialized writer; reads are rare at this scale.
	db.SetMaxOpenConns(1)

	if err := initSchema(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if batchInterval <= 0 {
		batchInterval = 250 * time.Millisecond
	}

	s := &Store{
		db:            db,
		in:            make(chan model.Reading, 4096),
		onCommit:      onCommit,
		batchInterval: batchInterval,
		closed:        make(chan struct{}),
	}
	s.wg.Add(1)
	go s.writeLoop()
	return s, nil
}

func initSchema(db *sql.DB) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS samples (
    id      INTEGER PRIMARY KEY,
    ts_us   INTEGER NOT NULL,   -- unix microseconds
    channel TEXT    NOT NULL,
    value   REAL    NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_samples_ts ON samples(ts_us);

-- Pump Profile (axiom #3 — the pump self-describes). Exactly one row has
-- is_active=1: the pump this Pi is. profile_channels declares which of the
-- format's channels this physical pump actually has, with any label/uom/decimals
-- overrides. See docs/design/data-model.md and the phase3 scope.
CREATE TABLE IF NOT EXISTS pump_profiles (
    id            INTEGER PRIMARY KEY,
    name          TEXT    NOT NULL,
    units         INTEGER NOT NULL DEFAULT 1,     -- number of pumping units
    daq_format_id TEXT    NOT NULL,               -- references the code preset, e.g. "intellisense"
    is_active     INTEGER NOT NULL DEFAULT 0,     -- exactly one row = 1 (the pump this Pi is)
    created_at_us INTEGER NOT NULL,
    updated_at_us INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS profile_channels (
    id          INTEGER PRIMARY KEY,
    profile_id  INTEGER NOT NULL REFERENCES pump_profiles(id) ON DELETE CASCADE,
    channel_id  TEXT    NOT NULL,                 -- e.g. "unit1.pressure" (matches the DaqFormat field map)
    role        TEXT    NOT NULL,                 -- pressure|rate|density|volume|meta|...
    scope       TEXT    NOT NULL,                 -- unit|aggregate|stage|job|meta
    unit_index  INTEGER NOT NULL DEFAULT 0,       -- 1-based when scope=unit; 0 otherwise
    label       TEXT    NOT NULL,
    uom         TEXT    NOT NULL DEFAULT '',
    decimals    INTEGER NOT NULL DEFAULT 2,
    enabled     INTEGER NOT NULL DEFAULT 1,       -- the pump physically has this channel
    sort_order  INTEGER NOT NULL DEFAULT 0,
    UNIQUE(profile_id, channel_id)
);
CREATE INDEX IF NOT EXISTS idx_profile_channels_profile ON profile_channels(profile_id);

-- Job (the unit of work recordings attach to). Exactly one row has is_active=1:
-- the job whose recording segments new starts open under. TEXT fields default ''
-- so a job needs only a name. See docs/design/data-model.md and the phase3 scope
-- (D8 job fields).
CREATE TABLE IF NOT EXISTS jobs (
    id            INTEGER PRIMARY KEY,
    name          TEXT    NOT NULL,
    company       TEXT    NOT NULL DEFAULT '',    -- operator/company
    well          TEXT    NOT NULL DEFAULT '',    -- well / location name
    casing_size   TEXT    NOT NULL DEFAULT '',    -- e.g. "9-5/8\""
    job_type      TEXT    NOT NULL DEFAULT '',    -- surface / intermediate / production / squeeze
    location      TEXT    NOT NULL DEFAULT '',    -- field / lease
    cementer      TEXT    NOT NULL DEFAULT '',    -- foreman / crew lead
    notes         TEXT    NOT NULL DEFAULT '',
    is_active     INTEGER NOT NULL DEFAULT 0,     -- exactly one row = 1 (the active job)
    created_at_us INTEGER NOT NULL,
    updated_at_us INTEGER NOT NULL
);

-- Recording segment: a MARKER over the always-on samples store (axioms #1 & #5).
-- started_at_us / stopped_at_us are points on the SAME timeline as samples.ts_us
-- (time.Now().UnixMicro), so Phase-4 charts can filter samples to a segment range.
-- stopped_at_us NULL = open (recording in progress). A job has many segments.
-- Start/stop/adjust ONLY insert/update rows here — they never gate ingestion, the
-- live readout, or stage volume.
CREATE TABLE IF NOT EXISTS recording_segments (
    id            INTEGER PRIMARY KEY,
    job_id        INTEGER NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    started_at_us INTEGER NOT NULL,               -- a point on the samples.ts_us timeline
    stopped_at_us INTEGER,                          -- NULL = open (recording in progress)
    created_at_us INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_segments_job ON recording_segments(job_id);`
	_, err := db.Exec(ddl)
	return err
}

// Submit queues a reading for durable storage. It blocks only if the writer has
// fallen far behind (backpressure) — never drops, because the store is the source
// of truth.
func (s *Store) Submit(r model.Reading) {
	select {
	case <-s.closed:
		// Shutting down; drop silently (the raw log still has it).
	case s.in <- r:
	}
}

func (s *Store) writeLoop() {
	defer s.wg.Done()
	t := time.NewTicker(s.batchInterval)
	defer t.Stop()

	batch := make([]model.Reading, 0, maxBatch)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := s.commit(batch); err != nil {
			// Data is already safe in the raw log; log and keep ingesting.
			fmt.Printf("store: commit failed (%d readings): %v\n", len(batch), err)
		} else if s.onCommit != nil {
			for _, r := range batch {
				s.onCommit(r)
			}
		}
		batch = batch[:0]
	}

	for {
		select {
		case r := <-s.in:
			batch = append(batch, r)
			if len(batch) >= maxBatch {
				flush()
			}
		case <-t.C:
			flush()
		case <-s.closed:
			// Drain whatever is buffered, commit it, and exit. We never close s.in,
			// so a late Submit can't panic on a closed channel.
			for {
				select {
				case r := <-s.in:
					batch = append(batch, r)
					if len(batch) >= maxBatch {
						flush()
					}
				default:
					flush()
					return
				}
			}
		}
	}
}

func (s *Store) commit(batch []model.Reading) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO samples (ts_us, channel, value) VALUES (?, ?, ?)`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, r := range batch {
		tsUS := r.TS.UnixMicro()
		for ch, v := range r.Values {
			if _, err := stmt.Exec(tsUS, ch, v); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
	}
	return tx.Commit()
}

// Stats is a lightweight snapshot used for verification and the debug endpoint.
type Stats struct {
	Rows     int64     `json:"rows"`
	LatestTS time.Time `json:"latest_ts"`
}

func (s *Store) Stats() (Stats, error) {
	var st Stats
	var latest sql.NullInt64
	row := s.db.QueryRow(`SELECT COUNT(*), MAX(ts_us) FROM samples`)
	if err := row.Scan(&st.Rows, &latest); err != nil {
		return st, err
	}
	if latest.Valid {
		st.LatestTS = time.UnixMicro(latest.Int64)
	}
	return st, nil
}

// Close stops the writer (flushing any buffered batch) and closes the database.
// The caller must ensure no more Submit calls happen after Close begins (cancel
// the source first).
func (s *Store) Close() error {
	close(s.closed)
	s.wg.Wait()
	return s.db.Close()
}
