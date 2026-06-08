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
CREATE INDEX IF NOT EXISTS idx_samples_ts ON samples(ts_us);`
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
