package store

// Recording-segment persistence: markers over the always-on samples store (axioms
// #1 & #5). These methods ONLY insert/update rows in recording_segments — they never
// touch ingestion, the live readout, or stage volume. Samples are stored continuously
// whether or not a segment is open; a segment is just a time window over that store.
//
// The store is the SINGLE DB owner (axiom #4 / D2): synchronous methods on the one
// *sql.DB connection, serialized with the sample writeLoop by the 1-conn pool + WAL.
//
// Times are time.Now().UnixMicro() — the SAME clock/scale as samples.ts_us — so a
// Phase-4 chart can filter samples to [started_at_us, stopped_at_us). A segment with
// stopped_at_us NULL is the single open segment (recording in progress).

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Segment is one recording marker, as sent on the wire and stored in
// recording_segments. StoppedAtUS is a pointer so NULL (open) round-trips as JSON
// null. The JSON tags ARE the client contract (mirrored in web/src/types.ts).
type Segment struct {
	ID          int64  `json:"id"`
	JobID       int64  `json:"jobId"`
	StartedAtUS int64  `json:"startedAtUs"`
	StoppedAtUS *int64 `json:"stoppedAtUs"` // nil = open (recording in progress)
	CreatedAtUS int64  `json:"createdAtUs"`
}

// ErrNoActiveJob is returned by StartRecording when no job is active to open a
// segment under.
var ErrNoActiveJob = errors.New("no active job")

// ErrRecording is returned when an action is refused because a segment is already
// open: a second StartRecording, or switching the active job mid-recording.
var ErrRecording = errors.New("already recording")

// ErrNotRecording is returned by StopRecording when no segment is open.
var ErrNotRecording = errors.New("not recording")

// ErrNoSuchSegment is returned by AdjustSegment when an id matches nothing.
var ErrNoSuchSegment = errors.New("no such segment")

// ErrBadSegmentRange is returned when an adjust would make started_at > stopped_at.
var ErrBadSegmentRange = errors.New("started_at must be <= stopped_at")

// openSegmentJob reports the job id of the currently open segment (stopped_at_us
// NULL), if any. Shared by SetActiveJob's guard and the recording methods.
func openSegmentJob(q queryer) (jobID int64, ok bool, err error) {
	row := q.QueryRow(
		`SELECT job_id FROM recording_segments WHERE stopped_at_us IS NULL LIMIT 1`,
	)
	err = row.Scan(&jobID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("open segment lookup: %w", err)
	}
	return jobID, true, nil
}

// scanSegment scans a segment row in the canonical column order. stopped_at_us is
// NULLable.
func scanSegment(sc interface{ Scan(...any) error }) (Segment, error) {
	var s Segment
	var stopped sql.NullInt64
	if err := sc.Scan(&s.ID, &s.JobID, &s.StartedAtUS, &stopped, &s.CreatedAtUS); err != nil {
		return Segment{}, err
	}
	if stopped.Valid {
		v := stopped.Int64
		s.StoppedAtUS = &v
	}
	return s, nil
}

const segmentCols = `id, job_id, started_at_us, stopped_at_us, created_at_us`

// StartRecording opens a new segment under the active job, stamping started_at_us =
// now (a point on the samples timeline). It is axiom #1-safe: it inserts a marker
// row only — ingestion and the live readout are untouched, and it never resets stage
// volume (axiom #5). It returns ErrNoActiveJob when no job is active. If a segment is
// already open it does NOT open a second: it returns that open segment with
// ErrRecording, so the caller can surface "already recording" and show the open one.
func (s *Store) StartRecording() (Segment, error) {
	now := time.Now().UnixMicro()

	tx, err := s.db.Begin()
	if err != nil {
		return Segment{}, fmt.Errorf("start recording begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Already recording? Return the open segment with ErrRecording (no second open).
	if open, ok, err := openSegmentRow(tx); err != nil {
		return Segment{}, err
	} else if ok {
		return open, ErrRecording
	}

	var jobID int64
	err = tx.QueryRow(`SELECT id FROM jobs WHERE is_active = 1 LIMIT 1`).Scan(&jobID)
	if errors.Is(err, sql.ErrNoRows) {
		return Segment{}, ErrNoActiveJob
	}
	if err != nil {
		return Segment{}, fmt.Errorf("start recording active job: %w", err)
	}

	res, err := tx.Exec(
		`INSERT INTO recording_segments (job_id, started_at_us, stopped_at_us, created_at_us)
		 VALUES (?, ?, NULL, ?)`,
		jobID, now, now,
	)
	if err != nil {
		return Segment{}, fmt.Errorf("start recording insert: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Segment{}, fmt.Errorf("start recording id: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Segment{}, fmt.Errorf("start recording commit: %w", err)
	}
	return Segment{ID: id, JobID: jobID, StartedAtUS: now, StoppedAtUS: nil, CreatedAtUS: now}, nil
}

// openSegmentRow reads the full open segment row (stopped_at_us NULL), if any.
func openSegmentRow(q queryer) (Segment, bool, error) {
	row := q.QueryRow(
		`SELECT ` + segmentCols + ` FROM recording_segments WHERE stopped_at_us IS NULL LIMIT 1`,
	)
	seg, err := scanSegment(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Segment{}, false, nil
	}
	if err != nil {
		return Segment{}, false, fmt.Errorf("open segment row: %w", err)
	}
	return seg, true, nil
}

// StopRecording closes the open segment, stamping stopped_at_us = now. It is axiom
// #1-safe (marker-only; never gates ingestion/live or resets stage volume). Returns
// ErrNotRecording when no segment is open.
func (s *Store) StopRecording() (Segment, error) {
	now := time.Now().UnixMicro()

	tx, err := s.db.Begin()
	if err != nil {
		return Segment{}, fmt.Errorf("stop recording begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	open, ok, err := openSegmentRow(tx)
	if err != nil {
		return Segment{}, err
	}
	if !ok {
		return Segment{}, ErrNotRecording
	}

	if _, err := tx.Exec(
		`UPDATE recording_segments SET stopped_at_us = ? WHERE id = ?`, now, open.ID,
	); err != nil {
		return Segment{}, fmt.Errorf("stop recording update: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Segment{}, fmt.Errorf("stop recording commit: %w", err)
	}
	open.StoppedAtUS = &now
	return open, nil
}

// RecordingState reports whether a segment is open and, if so, the open segment id
// and its job id. recording=false leaves openSegmentID and jobID 0.
func (s *Store) RecordingState() (recording bool, openSegmentID int64, jobID int64, err error) {
	open, ok, err := openSegmentRow(s.db)
	if err != nil {
		return false, 0, 0, err
	}
	if !ok {
		return false, 0, 0, nil
	}
	return true, open.ID, open.JobID, nil
}

// GetSegment returns one segment by id. ok is false (no error) when no such row
// exists. Used by the adjust handler to echo the refreshed segment.
func (s *Store) GetSegment(id int64) (Segment, bool, error) {
	row := s.db.QueryRow(`SELECT `+segmentCols+` FROM recording_segments WHERE id = ?`, id)
	seg, err := scanSegment(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Segment{}, false, nil
	}
	if err != nil {
		return Segment{}, false, fmt.Errorf("get segment: %w", err)
	}
	return seg, true, nil
}

// ListSegments returns a job's segments, oldest first (chronological for the chart).
func (s *Store) ListSegments(jobID int64) ([]Segment, error) {
	rows, err := s.db.Query(
		`SELECT `+segmentCols+` FROM recording_segments WHERE job_id = ? ORDER BY started_at_us, id`,
		jobID,
	)
	if err != nil {
		return nil, fmt.Errorf("list segments: %w", err)
	}
	defer rows.Close()

	segs := make([]Segment, 0)
	for rows.Next() {
		seg, err := scanSegment(rows)
		if err != nil {
			return nil, fmt.Errorf("scan segment: %w", err)
		}
		segs = append(segs, seg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate segments: %w", err)
	}
	return segs, nil
}

// AdjustSegment nudges a segment's endpoints after the fact (axiom #5 adjustability).
// Both fields are optional (nil = leave unchanged); pass startedAtUS to move the
// start, stoppedAtUS to move/set the stop. Passing a non-nil stoppedAtUS on an open
// segment closes it; there is no way here to RE-OPEN a closed segment (set stop NULL)
// because the JSON contract carries values, not an explicit-null sentinel — re-open
// is a deferred nicety. The resulting [started, stopped] must be ordered when both
// are known, else ErrBadSegmentRange. Returns ErrNoSuchSegment for an unknown id.
func (s *Store) AdjustSegment(id int64, startedAtUS *int64, stoppedAtUS *int64) error {
	if startedAtUS == nil && stoppedAtUS == nil {
		return nil // nothing to change
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("adjust segment begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Read current endpoints so we can validate the resulting range.
	var curStart int64
	var curStop sql.NullInt64
	err = tx.QueryRow(
		`SELECT started_at_us, stopped_at_us FROM recording_segments WHERE id = ?`, id,
	).Scan(&curStart, &curStop)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNoSuchSegment
	}
	if err != nil {
		return fmt.Errorf("adjust segment lookup: %w", err)
	}

	newStart := curStart
	if startedAtUS != nil {
		newStart = *startedAtUS
	}
	newStopValid := curStop.Valid
	newStop := curStop.Int64
	if stoppedAtUS != nil {
		newStopValid = true
		newStop = *stoppedAtUS
	}
	if newStopValid && newStart > newStop {
		return ErrBadSegmentRange
	}

	sets := make([]string, 0, 2)
	args := make([]any, 0, 3)
	if startedAtUS != nil {
		sets = append(sets, "started_at_us = ?")
		args = append(args, *startedAtUS)
	}
	if stoppedAtUS != nil {
		sets = append(sets, "stopped_at_us = ?")
		args = append(args, *stoppedAtUS)
	}
	args = append(args, id)
	if _, err := tx.Exec(
		`UPDATE recording_segments SET `+joinComma(sets)+` WHERE id = ?`, args...,
	); err != nil {
		return fmt.Errorf("adjust segment update: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("adjust segment commit: %w", err)
	}
	return nil
}
