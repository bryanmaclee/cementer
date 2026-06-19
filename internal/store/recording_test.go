package store

import "testing"

// activeJob creates a job and makes it active, returning its id.
func activeJob(t *testing.T, st *Store, name string) int64 {
	t.Helper()
	id, err := st.CreateJob(Job{Name: name})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if err := st.SetActiveJob(id); err != nil {
		t.Fatalf("SetActiveJob: %v", err)
	}
	return id
}

func TestStartRecordingRequiresActiveJob(t *testing.T) {
	st := openTestStore(t)
	// A job that is NOT active must not satisfy StartRecording.
	if _, err := st.CreateJob(Job{Name: "idle"}); err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if _, err := st.StartRecording(); err != ErrNoActiveJob {
		t.Fatalf("want ErrNoActiveJob, got %v", err)
	}
}

func TestStartStopRecordingLifecycle(t *testing.T) {
	st := openTestStore(t)
	jobID := activeJob(t, st, "Job 1")

	// Not recording initially.
	rec, openID, jID, err := st.RecordingState()
	if err != nil {
		t.Fatalf("RecordingState: %v", err)
	}
	if rec || openID != 0 || jID != 0 {
		t.Fatalf("should not be recording initially: rec=%v open=%d job=%d", rec, openID, jID)
	}

	// Start: open segment under the active job, stopped_at NULL.
	seg, err := st.StartRecording()
	if err != nil {
		t.Fatalf("StartRecording: %v", err)
	}
	if seg.JobID != jobID {
		t.Fatalf("segment bound to wrong job: %d want %d", seg.JobID, jobID)
	}
	if seg.StoppedAtUS != nil {
		t.Fatalf("open segment should have nil StoppedAtUS, got %v", *seg.StoppedAtUS)
	}
	if seg.StartedAtUS == 0 {
		t.Fatal("StartedAtUS not stamped")
	}

	// State reflects recording.
	rec, openID, jID, err = st.RecordingState()
	if err != nil {
		t.Fatalf("RecordingState: %v", err)
	}
	if !rec || openID != seg.ID || jID != jobID {
		t.Fatalf("state mismatch: rec=%v open=%d (want %d) job=%d (want %d)", rec, openID, seg.ID, jID, jobID)
	}

	// Stop: stopped_at set.
	closed, err := st.StopRecording()
	if err != nil {
		t.Fatalf("StopRecording: %v", err)
	}
	if closed.ID != seg.ID {
		t.Fatalf("stop closed wrong segment: %d want %d", closed.ID, seg.ID)
	}
	if closed.StoppedAtUS == nil {
		t.Fatal("stopped segment should have non-nil StoppedAtUS")
	}
	if *closed.StoppedAtUS < closed.StartedAtUS {
		t.Fatalf("stopped_at before started_at: %d < %d", *closed.StoppedAtUS, closed.StartedAtUS)
	}

	// Back to not recording.
	rec, _, _, _ = st.RecordingState()
	if rec {
		t.Fatal("should not be recording after stop")
	}
}

func TestDoubleStartReturnsOpenSegment(t *testing.T) {
	st := openTestStore(t)
	activeJob(t, st, "Job 1")

	first, err := st.StartRecording()
	if err != nil {
		t.Fatalf("first StartRecording: %v", err)
	}
	// Second start must NOT open a second segment; it returns the open one + ErrRecording.
	again, err := st.StartRecording()
	if err != ErrRecording {
		t.Fatalf("want ErrRecording on double-start, got %v", err)
	}
	if again.ID != first.ID {
		t.Fatalf("double-start returned a different segment: %d want %d", again.ID, first.ID)
	}

	// Exactly one open segment exists.
	var open int
	if err := st.db.QueryRow(`SELECT COUNT(*) FROM recording_segments WHERE stopped_at_us IS NULL`).Scan(&open); err != nil {
		t.Fatalf("count open: %v", err)
	}
	if open != 1 {
		t.Fatalf("double-start opened a second segment: %d open", open)
	}
}

func TestStopWhenNotRecording(t *testing.T) {
	st := openTestStore(t)
	activeJob(t, st, "Job 1")
	if _, err := st.StopRecording(); err != ErrNotRecording {
		t.Fatalf("want ErrNotRecording, got %v", err)
	}
}

func TestListSegmentsMultiplePerJob(t *testing.T) {
	st := openTestStore(t)
	jobID := activeJob(t, st, "Job 1")

	// Two complete segments.
	if _, err := st.StartRecording(); err != nil {
		t.Fatalf("start 1: %v", err)
	}
	if _, err := st.StopRecording(); err != nil {
		t.Fatalf("stop 1: %v", err)
	}
	if _, err := st.StartRecording(); err != nil {
		t.Fatalf("start 2: %v", err)
	}
	if _, err := st.StopRecording(); err != nil {
		t.Fatalf("stop 2: %v", err)
	}

	segs, err := st.ListSegments(jobID)
	if err != nil {
		t.Fatalf("ListSegments: %v", err)
	}
	if len(segs) != 2 {
		t.Fatalf("want 2 segments, got %d", len(segs))
	}
	// Chronological order.
	if segs[0].StartedAtUS > segs[1].StartedAtUS {
		t.Fatal("segments not in chronological order")
	}
	for _, s := range segs {
		if s.JobID != jobID {
			t.Fatalf("segment bound to wrong job: %+v", s)
		}
	}
}

func TestAdjustSegmentMovesStart(t *testing.T) {
	st := openTestStore(t)
	jobID := activeJob(t, st, "Job 1")
	seg, _ := st.StartRecording()
	if _, err := st.StopRecording(); err != nil {
		t.Fatalf("stop: %v", err)
	}

	// Nudge the start 10 seconds earlier (axiom #5 after-the-fact adjustment).
	earlier := seg.StartedAtUS - 10_000_000
	if err := st.AdjustSegment(seg.ID, &earlier, nil); err != nil {
		t.Fatalf("AdjustSegment: %v", err)
	}

	segs, _ := st.ListSegments(jobID)
	if len(segs) != 1 {
		t.Fatalf("want 1 segment, got %d", len(segs))
	}
	if segs[0].StartedAtUS != earlier {
		t.Fatalf("start not moved: got %d want %d", segs[0].StartedAtUS, earlier)
	}
}

func TestAdjustSegmentBadOrdering(t *testing.T) {
	st := openTestStore(t)
	activeJob(t, st, "Job 1")
	seg, _ := st.StartRecording()
	if _, err := st.StopRecording(); err != nil {
		t.Fatalf("stop: %v", err)
	}

	// Move the start AFTER the stop -> rejected.
	afterStop := *mustStop(t, st, seg.ID) + 1_000_000
	if err := st.AdjustSegment(seg.ID, &afterStop, nil); err != ErrBadSegmentRange {
		t.Fatalf("want ErrBadSegmentRange, got %v", err)
	}
}

func TestAdjustSegmentMissing(t *testing.T) {
	st := openTestStore(t)
	ts := int64(123)
	if err := st.AdjustSegment(999, &ts, nil); err != ErrNoSuchSegment {
		t.Fatalf("want ErrNoSuchSegment, got %v", err)
	}
}

// mustStop reads a segment's current stopped_at_us, failing if it is open.
func mustStop(t *testing.T, st *Store, id int64) *int64 {
	t.Helper()
	segs, err := st.ListSegments(jobIDOf(t, st, id))
	if err != nil {
		t.Fatalf("ListSegments: %v", err)
	}
	for _, s := range segs {
		if s.ID == id {
			if s.StoppedAtUS == nil {
				t.Fatalf("segment %d is open", id)
			}
			return s.StoppedAtUS
		}
	}
	t.Fatalf("segment %d not found", id)
	return nil
}

func jobIDOf(t *testing.T, st *Store, segID int64) int64 {
	t.Helper()
	var jobID int64
	if err := st.db.QueryRow(`SELECT job_id FROM recording_segments WHERE id = ?`, segID).Scan(&jobID); err != nil {
		t.Fatalf("job_id of segment %d: %v", segID, err)
	}
	return jobID
}
