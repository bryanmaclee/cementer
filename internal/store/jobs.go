package store

// Job persistence. The store is the SINGLE DB owner (axiom #4 / D2): these are
// ordinary synchronous methods on the one *sql.DB connection (SetMaxOpenConns(1)),
// serialized against the sample writeLoop by the single-connection pool + WAL +
// busy_timeout. Operator config is infrequent, so synchronous is correct. No second
// *sql.DB, and HTTP handlers call these methods rather than touch the database.
//
// Exactly one job is active (is_active=1) — the job that recording segments open
// under. SetActiveJob demotes the previous active in a transaction, mirroring the
// profile is_active invariant. Changing the active job while a segment is open is
// REFUSED (recording.go's ErrRecording) so an open segment stays bound to one job.

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Job is one unit of work, as sent on the wire (the /api/jobs* contract) and stored
// in the jobs table. The JSON tags ARE the client contract (mirrored by hand in
// web/src/types.ts — no codegen). ID, IsActive, CreatedAtUS and UpdatedAtUS are
// server-owned: a create/update body sets only the descriptive fields.
type Job struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Company     string `json:"company"`
	Well        string `json:"well"`
	CasingSize  string `json:"casingSize"`
	JobType     string `json:"jobType"`
	Location    string `json:"location"`
	Cementer    string `json:"cementer"`
	Notes       string `json:"notes"`
	IsActive    bool   `json:"isActive"`
	CreatedAtUS int64  `json:"createdAtUs"`
	UpdatedAtUS int64  `json:"updatedAtUs"`
}

// ErrJobNameRequired is returned when a create is missing the (only) required field.
var ErrJobNameRequired = errors.New("job name is required")

// ErrNoSuchJob is returned when an id does not match any job row.
var ErrNoSuchJob = errors.New("no such job")

// CreateJob inserts a new job (descriptive fields from j; the rest server-stamped)
// and returns its id. The new job is NOT made active — use SetActiveJob. name is
// required; the other fields may be empty.
func (s *Store) CreateJob(j Job) (int64, error) {
	if j.Name == "" {
		return 0, ErrJobNameRequired
	}
	now := time.Now().UnixMicro()
	res, err := s.db.Exec(
		`INSERT INTO jobs
		   (name, company, well, casing_size, job_type, location, cementer, notes,
		    is_active, created_at_us, updated_at_us)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)`,
		j.Name, j.Company, j.Well, j.CasingSize, j.JobType, j.Location, j.Cementer, j.Notes,
		now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("create job: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("create job id: %w", err)
	}
	return id, nil
}

// scanJob scans a full job row in the canonical column order.
func scanJob(sc interface{ Scan(...any) error }) (Job, error) {
	var j Job
	var active int
	err := sc.Scan(
		&j.ID, &j.Name, &j.Company, &j.Well, &j.CasingSize, &j.JobType,
		&j.Location, &j.Cementer, &j.Notes, &active, &j.CreatedAtUS, &j.UpdatedAtUS,
	)
	if err != nil {
		return Job{}, err
	}
	j.IsActive = active != 0
	return j, nil
}

const jobCols = `id, name, company, well, casing_size, job_type, location, cementer,
	notes, is_active, created_at_us, updated_at_us`

// ListJobs returns all jobs, most recently created first.
func (s *Store) ListJobs() ([]Job, error) {
	rows, err := s.db.Query(`SELECT ` + jobCols + ` FROM jobs ORDER BY created_at_us DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	defer rows.Close()

	jobs := make([]Job, 0)
	for rows.Next() {
		j, err := scanJob(rows)
		if err != nil {
			return nil, fmt.Errorf("scan job: %w", err)
		}
		jobs = append(jobs, j)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate jobs: %w", err)
	}
	return jobs, nil
}

// GetJob returns one job by id. ok is false (no error) when no such row exists.
func (s *Store) GetJob(id int64) (Job, bool, error) {
	row := s.db.QueryRow(`SELECT `+jobCols+` FROM jobs WHERE id = ?`, id)
	j, err := scanJob(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Job{}, false, nil
	}
	if err != nil {
		return Job{}, false, fmt.Errorf("get job: %w", err)
	}
	return j, true, nil
}

// UpdateJob updates a job's descriptive fields (the server-owned id/is_active/
// timestamps are untouched except updated_at). name is required. Returns ErrNoSuchJob
// when id matches nothing.
func (s *Store) UpdateJob(id int64, j Job) error {
	if j.Name == "" {
		return ErrJobNameRequired
	}
	now := time.Now().UnixMicro()
	res, err := s.db.Exec(
		`UPDATE jobs SET
		   name = ?, company = ?, well = ?, casing_size = ?, job_type = ?,
		   location = ?, cementer = ?, notes = ?, updated_at_us = ?
		 WHERE id = ?`,
		j.Name, j.Company, j.Well, j.CasingSize, j.JobType, j.Location, j.Cementer, j.Notes,
		now, id,
	)
	if err != nil {
		return fmt.Errorf("update job: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update job rows: %w", err)
	}
	if n == 0 {
		return ErrNoSuchJob
	}
	return nil
}

// ActiveJob returns the active job. ok is false (no error) when no job is active.
func (s *Store) ActiveJob() (Job, bool, error) {
	row := s.db.QueryRow(`SELECT ` + jobCols + ` FROM jobs WHERE is_active = 1 LIMIT 1`)
	j, err := scanJob(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Job{}, false, nil
	}
	if err != nil {
		return Job{}, false, fmt.Errorf("active job: %w", err)
	}
	return j, true, nil
}

// SetActiveJob makes the job with id active and demotes any other active job, in one
// transaction (the is_active=1 invariant). It REFUSES (ErrRecording) to change the
// active job while a segment is open — an open segment stays bound to one job; the
// operator must stop recording first. Returns ErrNoSuchJob when id matches nothing.
func (s *Store) SetActiveJob(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("set active job begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Confirm the target exists.
	var exists int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM jobs WHERE id = ?`, id).Scan(&exists); err != nil {
		return fmt.Errorf("set active job lookup: %w", err)
	}
	if exists == 0 {
		return ErrNoSuchJob
	}

	// Guard (axiom-consistent): if a DIFFERENT job has an open segment, refuse — the
	// open segment must stay bound to its job. Switching to the already-recording job
	// is a no-op switch and is allowed.
	openJobID, hasOpen, err := openSegmentJob(tx)
	if err != nil {
		return err
	}
	if hasOpen && openJobID != id {
		return ErrRecording
	}

	if _, err := tx.Exec(`UPDATE jobs SET is_active = 0 WHERE is_active = 1 AND id != ?`, id); err != nil {
		return fmt.Errorf("set active job demote: %w", err)
	}
	if _, err := tx.Exec(
		`UPDATE jobs SET is_active = 1, updated_at_us = ? WHERE id = ?`,
		time.Now().UnixMicro(), id,
	); err != nil {
		return fmt.Errorf("set active job promote: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("set active job commit: %w", err)
	}
	return nil
}

// JobPrintConfig returns the raw per-job print-override JSON for a job (the
// print_config column). An empty string means "no override" (the API layer then uses
// the company default verbatim). The store stays company-agnostic (axiom #4 / D2): it
// persists/returns the raw blob only — the default+override -> effective merge lives
// in the API layer. found is false (no error) when no such job exists.
func (s *Store) JobPrintConfig(id int64) (raw string, found bool, err error) {
	row := s.db.QueryRow(`SELECT print_config FROM jobs WHERE id = ?`, id)
	if err := row.Scan(&raw); errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	} else if err != nil {
		return "", false, fmt.Errorf("job print config: %w", err)
	}
	return raw, true, nil
}

// SetJobPrintConfig stores the raw per-job print-override JSON for a job. The caller
// (API layer) is responsible for validating/canonicalizing the JSON before it lands
// here; the store persists it as-is and bumps updated_at. Returns ErrNoSuchJob when id
// matches nothing.
func (s *Store) SetJobPrintConfig(id int64, raw string) error {
	res, err := s.db.Exec(
		`UPDATE jobs SET print_config = ?, updated_at_us = ? WHERE id = ?`,
		raw, time.Now().UnixMicro(), id,
	)
	if err != nil {
		return fmt.Errorf("set job print config: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("set job print config rows: %w", err)
	}
	if n == 0 {
		return ErrNoSuchJob
	}
	return nil
}
