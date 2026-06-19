package store

import "testing"

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
