package store

import (
	"path/filepath"
	"testing"
	"time"
)

func TestCreateGetUpdateJobRoundTrip(t *testing.T) {
	st := openTestStore(t)

	in := Job{
		Name:       "Smith 4-12H",
		Company:    "Acme Energy",
		Well:       "Smith 4-12H",
		CasingSize: `9-5/8"`,
		JobType:    "surface",
		Location:   "Section 12",
		Cementer:   "B. Maclee",
		Notes:      "morning pour",
	}
	id, err := st.CreateJob(in)
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if id == 0 {
		t.Fatal("CreateJob returned id 0")
	}

	got, ok, err := st.GetJob(id)
	if err != nil || !ok {
		t.Fatalf("GetJob: ok=%v err=%v", ok, err)
	}
	if got.Name != in.Name || got.Company != in.Company || got.CasingSize != in.CasingSize ||
		got.JobType != in.JobType || got.Location != in.Location || got.Cementer != in.Cementer ||
		got.Notes != in.Notes {
		t.Fatalf("job fields not round-tripped: %+v", got)
	}
	if got.IsActive {
		t.Fatal("a newly created job must not be active")
	}
	if got.CreatedAtUS == 0 || got.UpdatedAtUS == 0 {
		t.Fatalf("timestamps not stamped: %+v", got)
	}

	// Update descriptive fields.
	upd := got
	upd.Company = "Acme Energy Services"
	upd.Notes = "afternoon pour"
	if err := st.UpdateJob(id, upd); err != nil {
		t.Fatalf("UpdateJob: %v", err)
	}
	got2, _, _ := st.GetJob(id)
	if got2.Company != "Acme Energy Services" || got2.Notes != "afternoon pour" {
		t.Fatalf("update not applied: %+v", got2)
	}
}

func TestCreateJobNameRequired(t *testing.T) {
	st := openTestStore(t)
	if _, err := st.CreateJob(Job{Name: ""}); err == nil {
		t.Fatal("expected ErrJobNameRequired for empty name")
	}
}

func TestGetJobMissing(t *testing.T) {
	st := openTestStore(t)
	if _, ok, err := st.GetJob(999); err != nil {
		t.Fatalf("GetJob errored: %v", err)
	} else if ok {
		t.Fatal("GetJob ok=true for missing id")
	}
}

func TestUpdateJobMissing(t *testing.T) {
	st := openTestStore(t)
	if err := st.UpdateJob(999, Job{Name: "x"}); err != ErrNoSuchJob {
		t.Fatalf("want ErrNoSuchJob, got %v", err)
	}
}

func TestListJobsNewestFirst(t *testing.T) {
	st := openTestStore(t)
	if _, err := st.CreateJob(Job{Name: "first"}); err != nil {
		t.Fatalf("create first: %v", err)
	}
	id2, err := st.CreateJob(Job{Name: "second"})
	if err != nil {
		t.Fatalf("create second: %v", err)
	}

	jobs, err := st.ListJobs()
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if len(jobs) != 2 {
		t.Fatalf("want 2 jobs, got %d", len(jobs))
	}
	// Newest (highest id) first.
	if jobs[0].ID != id2 {
		t.Fatalf("expected newest job first; got %+v", jobs[0])
	}
}

func TestSetActiveJobUniqueness(t *testing.T) {
	st := openTestStore(t)
	id1, _ := st.CreateJob(Job{Name: "A"})
	id2, _ := st.CreateJob(Job{Name: "B"})

	// No active job initially.
	if _, ok, err := st.ActiveJob(); err != nil {
		t.Fatalf("ActiveJob: %v", err)
	} else if ok {
		t.Fatal("no job should be active before SetActiveJob")
	}

	if err := st.SetActiveJob(id1); err != nil {
		t.Fatalf("SetActiveJob id1: %v", err)
	}
	a, ok, _ := st.ActiveJob()
	if !ok || a.ID != id1 {
		t.Fatalf("active job should be id1, got %+v ok=%v", a, ok)
	}

	// Switching to id2 demotes id1 — exactly one active.
	if err := st.SetActiveJob(id2); err != nil {
		t.Fatalf("SetActiveJob id2: %v", err)
	}
	var active int
	if err := st.db.QueryRow(`SELECT COUNT(*) FROM jobs WHERE is_active = 1`).Scan(&active); err != nil {
		t.Fatalf("count active: %v", err)
	}
	if active != 1 {
		t.Fatalf("want exactly one active job, got %d", active)
	}
	a2, _, _ := st.ActiveJob()
	if a2.ID != id2 {
		t.Fatalf("active job should be id2, got %+v", a2)
	}
}

func TestSetActiveJobMissing(t *testing.T) {
	st := openTestStore(t)
	if err := st.SetActiveJob(999); err != ErrNoSuchJob {
		t.Fatalf("want ErrNoSuchJob, got %v", err)
	}
}

// TestJobPrintConfigRoundTrip proves the per-job print override (the print_config
// JSON column) defaults to "" (no override) and round-trips a raw blob.
func TestJobPrintConfigRoundTrip(t *testing.T) {
	st := openTestStore(t)
	id, err := st.CreateJob(Job{Name: "Smith 4-12H"})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	// A fresh job has no override.
	raw, found, err := st.JobPrintConfig(id)
	if err != nil || !found {
		t.Fatalf("JobPrintConfig: found=%v err=%v", found, err)
	}
	if raw != "" {
		t.Fatalf("fresh job print_config should be empty, got %q", raw)
	}

	// Store a raw override and read it back verbatim.
	const ov = `{"title":"Surface Job","pageSize":"a4"}`
	if err := st.SetJobPrintConfig(id, ov); err != nil {
		t.Fatalf("SetJobPrintConfig: %v", err)
	}
	raw, found, err = st.JobPrintConfig(id)
	if err != nil || !found {
		t.Fatalf("JobPrintConfig after set: found=%v err=%v", found, err)
	}
	if raw != ov {
		t.Fatalf("override not round-tripped: got %q want %q", raw, ov)
	}
}

func TestJobPrintConfigMissing(t *testing.T) {
	st := openTestStore(t)
	if _, found, err := st.JobPrintConfig(999); err != nil {
		t.Fatalf("JobPrintConfig errored: %v", err)
	} else if found {
		t.Fatal("found=true for missing job")
	}
	if err := st.SetJobPrintConfig(999, "{}"); err != ErrNoSuchJob {
		t.Fatalf("want ErrNoSuchJob, got %v", err)
	}
}

// TestMigrationIdempotentAcrossReopen proves the print_config column migration is a
// no-op on reopen: a DB created, closed, and reopened keeps the stored override and
// does not error. This also exercises ADD COLUMN being skipped when present.
func TestMigrationIdempotentAcrossReopen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "migrate.db")

	st, err := Open(path, 50*time.Millisecond, nil)
	if err != nil {
		t.Fatalf("open 1: %v", err)
	}
	id, err := st.CreateJob(Job{Name: "Job 1"})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if err := st.SetJobPrintConfig(id, `{"pageSize":"a4"}`); err != nil {
		t.Fatalf("SetJobPrintConfig: %v", err)
	}
	if err := st.Close(); err != nil {
		t.Fatalf("close 1: %v", err)
	}

	// Reopen the same DB — initSchema + migrate run again and must be a no-op.
	st2, err := Open(path, 50*time.Millisecond, nil)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	t.Cleanup(func() { _ = st2.Close() })

	raw, found, err := st2.JobPrintConfig(id)
	if err != nil || !found {
		t.Fatalf("JobPrintConfig after reopen: found=%v err=%v", found, err)
	}
	if raw != `{"pageSize":"a4"}` {
		t.Fatalf("override lost across reopen: %q", raw)
	}
}

// TestSetActiveJobRejectedWhileRecording proves the axiom-consistent guard: you
// cannot switch the active job while a segment is open (the open segment stays bound
// to one job). Switching to the SAME already-recording job is allowed (no-op).
func TestSetActiveJobRejectedWhileRecording(t *testing.T) {
	st := openTestStore(t)
	id1, _ := st.CreateJob(Job{Name: "A"})
	id2, _ := st.CreateJob(Job{Name: "B"})

	if err := st.SetActiveJob(id1); err != nil {
		t.Fatalf("SetActiveJob id1: %v", err)
	}
	if _, err := st.StartRecording(); err != nil {
		t.Fatalf("StartRecording: %v", err)
	}

	// Switching to a different job mid-recording is refused.
	if err := st.SetActiveJob(id2); err != ErrRecording {
		t.Fatalf("want ErrRecording switching jobs mid-recording, got %v", err)
	}
	// The active job is still id1.
	a, _, _ := st.ActiveJob()
	if a.ID != id1 {
		t.Fatalf("active job changed despite refusal: %+v", a)
	}

	// Re-selecting the already-recording job is allowed.
	if err := st.SetActiveJob(id1); err != nil {
		t.Fatalf("re-selecting recording job should be allowed, got %v", err)
	}
}
