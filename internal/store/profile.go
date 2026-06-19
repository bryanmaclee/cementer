package store

// Pump-Profile persistence and the profile wire types (axiom #3: the pump
// self-describes). The store is the SINGLE DB owner (axiom #4 / D2): these are
// ordinary synchronous methods on the one *sql.DB connection (SetMaxOpenConns(1)),
// serialized against the sample writeLoop by the single-connection pool + WAL +
// busy_timeout. They are infrequent (operator config), so synchronous is correct.
// No second *sql.DB, and HTTP handlers must call these methods rather than touch
// the database directly.
//
// The store stays FORMAT-AGNOSTIC (axiom #2): it never imports internal/daqformat.
// The seed vocabulary is supplied by the caller (main) as a []Channel; the store
// only knows how to persist and read it.

import (
	"database/sql"
	"fmt"
	"time"
)

// Channel is one channel in a Pump Profile, as sent on the wire (hello/profile
// message) and stored in profile_channels. The JSON tags ARE the client contract
// (mirrored by hand in web/src/types.ts — no codegen). This shape carries display
// metadata only; the `enabled` flag is NOT part of the wire frame because that
// frame already lists enabled channels only. For the editor view (which needs the
// disabled rows too) see EditorChannel.
type Channel struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	Scope     string `json:"scope"`     // unit|aggregate|stage|job|meta
	UnitIndex int    `json:"unitIndex"` // 1-based when scope=="unit"; 0 otherwise
	Label     string `json:"label"`
	UoM       string `json:"uom"`
	Decimals  int    `json:"decimals"`
}

// Profile is the active Pump Profile as sent in the hello/profile WS frame. Its
// Channels are the ENABLED channels only, in sort_order.
type Profile struct {
	Name     string    `json:"name"`
	Units    int       `json:"units"`
	FormatID string    `json:"formatId"`
	Channels []Channel `json:"channels"`
}

// EditorChannel is a profile channel as the editor (GET /api/profile) sees it: the
// display metadata PLUS the persisted enabled flag and sort order, so the operator
// can see and toggle disabled channels. Disabled channels never reach the WS frame
// but must be visible (and editable) here.
type EditorChannel struct {
	Channel
	Enabled   bool `json:"enabled"`
	SortOrder int  `json:"sortOrder"`
}

// EditorProfile is the full active profile for the editor: ALL channels (enabled
// and disabled), in sort_order.
type EditorProfile struct {
	Name     string          `json:"name"`
	Units    int             `json:"units"`
	FormatID string          `json:"formatId"`
	Channels []EditorChannel `json:"channels"`
}

// SeedChannel is the input vocabulary a caller supplies to create/seed a profile.
// main builds these from the active DaqFormat's channel vocab; the store does not
// import daqformat. All seeded channels are enabled; sort_order follows the slice
// order.
type SeedChannel struct {
	ID        string
	Role      string
	Scope     string
	UnitIndex int
	Label     string
	UoM       string
	Decimals  int
}

// HasActiveProfile reports whether an active profile already exists. main uses it
// to make seeding idempotent (a second boot must not duplicate the profile).
func (s *Store) HasActiveProfile() (bool, error) {
	var n int
	row := s.db.QueryRow(`SELECT COUNT(*) FROM pump_profiles WHERE is_active = 1`)
	if err := row.Scan(&n); err != nil {
		return false, fmt.Errorf("count active profiles: %w", err)
	}
	return n > 0, nil
}

// SeedActiveProfile creates a new profile from the supplied vocabulary and makes it
// the sole active profile, in one transaction. It is intended for first-run seeding;
// callers should guard it with HasActiveProfile so a reboot does not duplicate the
// profile. All channels are created enabled, sort_order = slice index.
func (s *Store) SeedActiveProfile(name string, units int, formatID string, channels []SeedChannel) error {
	if units < 1 {
		units = 1
	}
	now := time.Now().UnixMicro()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("seed profile begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Demote any existing active profile so the is_active=1 invariant holds.
	if _, err := tx.Exec(`UPDATE pump_profiles SET is_active = 0 WHERE is_active = 1`); err != nil {
		return fmt.Errorf("seed profile demote: %w", err)
	}

	res, err := tx.Exec(
		`INSERT INTO pump_profiles (name, units, daq_format_id, is_active, created_at_us, updated_at_us)
		 VALUES (?, ?, ?, 1, ?, ?)`,
		name, units, formatID, now, now,
	)
	if err != nil {
		return fmt.Errorf("seed profile insert: %w", err)
	}
	profileID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("seed profile id: %w", err)
	}

	if err := insertChannels(tx, profileID, channels); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("seed profile commit: %w", err)
	}
	return nil
}

// insertChannels inserts the supplied vocabulary as enabled channels for a profile,
// sort_order following slice order. Used by both seed and reset.
func insertChannels(tx *sql.Tx, profileID int64, channels []SeedChannel) error {
	stmt, err := tx.Prepare(
		`INSERT INTO profile_channels
		   (profile_id, channel_id, role, scope, unit_index, label, uom, decimals, enabled, sort_order)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, ?)`,
	)
	if err != nil {
		return fmt.Errorf("prepare channel insert: %w", err)
	}
	defer stmt.Close()

	for i, c := range channels {
		decimals := c.Decimals
		if _, err := stmt.Exec(
			profileID, c.ID, c.Role, c.Scope, c.UnitIndex, c.Label, c.UoM, decimals, i,
		); err != nil {
			return fmt.Errorf("insert channel %q: %w", c.ID, err)
		}
	}
	return nil
}

// activeProfileMeta reads the active profile's id and header fields.
func (s *Store) activeProfileMeta(q queryer) (id int64, name string, units int, formatID string, err error) {
	row := q.QueryRow(
		`SELECT id, name, units, daq_format_id FROM pump_profiles WHERE is_active = 1 LIMIT 1`,
	)
	err = row.Scan(&id, &name, &units, &formatID)
	return
}

// queryer is the read subset shared by *sql.DB and *sql.Tx.
type queryer interface {
	QueryRow(query string, args ...any) *sql.Row
	Query(query string, args ...any) (*sql.Rows, error)
}

// ActiveProfile returns the active profile with its ENABLED channels only, ordered
// by sort_order — the shape sent in the hello/profile WS frame. ok is false when no
// active profile exists (should not happen after seeding).
func (s *Store) ActiveProfile() (p Profile, ok bool, err error) {
	id, name, units, formatID, err := s.activeProfileMeta(s.db)
	if err == sql.ErrNoRows {
		return Profile{}, false, nil
	}
	if err != nil {
		return Profile{}, false, fmt.Errorf("active profile: %w", err)
	}

	rows, err := s.db.Query(
		`SELECT channel_id, role, scope, unit_index, label, uom, decimals
		   FROM profile_channels
		  WHERE profile_id = ? AND enabled = 1
		  ORDER BY sort_order, channel_id`,
		id,
	)
	if err != nil {
		return Profile{}, false, fmt.Errorf("active profile channels: %w", err)
	}
	defer rows.Close()

	var chans []Channel
	for rows.Next() {
		var c Channel
		if err := rows.Scan(&c.ID, &c.Role, &c.Scope, &c.UnitIndex, &c.Label, &c.UoM, &c.Decimals); err != nil {
			return Profile{}, false, fmt.Errorf("scan profile channel: %w", err)
		}
		chans = append(chans, c)
	}
	if err := rows.Err(); err != nil {
		return Profile{}, false, fmt.Errorf("iterate profile channels: %w", err)
	}

	return Profile{Name: name, Units: units, FormatID: formatID, Channels: chans}, true, nil
}

// ActiveEditorProfile returns the active profile with ALL channels (enabled and
// disabled), in sort_order — the shape the editor GET returns so the operator can
// see and toggle disabled channels. ok is false when no active profile exists.
func (s *Store) ActiveEditorProfile() (p EditorProfile, ok bool, err error) {
	id, name, units, formatID, err := s.activeProfileMeta(s.db)
	if err == sql.ErrNoRows {
		return EditorProfile{}, false, nil
	}
	if err != nil {
		return EditorProfile{}, false, fmt.Errorf("active editor profile: %w", err)
	}

	rows, err := s.db.Query(
		`SELECT channel_id, role, scope, unit_index, label, uom, decimals, enabled, sort_order
		   FROM profile_channels
		  WHERE profile_id = ?
		  ORDER BY sort_order, channel_id`,
		id,
	)
	if err != nil {
		return EditorProfile{}, false, fmt.Errorf("editor profile channels: %w", err)
	}
	defer rows.Close()

	var chans []EditorChannel
	for rows.Next() {
		var c EditorChannel
		var enabled int
		if err := rows.Scan(
			&c.ID, &c.Role, &c.Scope, &c.UnitIndex, &c.Label, &c.UoM, &c.Decimals, &enabled, &c.SortOrder,
		); err != nil {
			return EditorProfile{}, false, fmt.Errorf("scan editor channel: %w", err)
		}
		c.Enabled = enabled != 0
		chans = append(chans, c)
	}
	if err := rows.Err(); err != nil {
		return EditorProfile{}, false, fmt.Errorf("iterate editor channels: %w", err)
	}

	return EditorProfile{Name: name, Units: units, FormatID: formatID, Channels: chans}, true, nil
}

// ChannelUpdate is a per-channel edit applied by UpdateActiveProfile. Pointer
// fields are optional (nil = leave unchanged), so a PUT can patch just the fields
// it sends. ChannelID identifies the row (within the active profile).
type ChannelUpdate struct {
	ChannelID string
	Enabled   *bool
	Label     *string
	UoM       *string
	Decimals  *int
	SortOrder *int
}

// UpdateActiveProfile updates the active profile's units (when units >= 1) and
// applies per-channel edits, in one transaction. An edit for an unknown channel_id
// is an error (the caller sent a bad id). Pass units <= 0 to leave units unchanged.
func (s *Store) UpdateActiveProfile(units int, edits []ChannelUpdate) error {
	now := time.Now().UnixMicro()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("update profile begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var id int64
	row := tx.QueryRow(`SELECT id FROM pump_profiles WHERE is_active = 1 LIMIT 1`)
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("update profile: no active profile")
		}
		return fmt.Errorf("update profile lookup: %w", err)
	}

	if units >= 1 {
		if _, err := tx.Exec(
			`UPDATE pump_profiles SET units = ?, updated_at_us = ? WHERE id = ?`,
			units, now, id,
		); err != nil {
			return fmt.Errorf("update profile units: %w", err)
		}
	} else {
		if _, err := tx.Exec(
			`UPDATE pump_profiles SET updated_at_us = ? WHERE id = ?`, now, id,
		); err != nil {
			return fmt.Errorf("touch profile: %w", err)
		}
	}

	for _, e := range edits {
		if err := applyChannelEdit(tx, id, e); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("update profile commit: %w", err)
	}
	return nil
}

// applyChannelEdit applies one ChannelUpdate. It uses a COALESCE-style approach:
// build the SET clause from only the supplied fields so unset fields are untouched.
func applyChannelEdit(tx *sql.Tx, profileID int64, e ChannelUpdate) error {
	sets := make([]string, 0, 5)
	args := make([]any, 0, 6)
	if e.Enabled != nil {
		sets = append(sets, "enabled = ?")
		v := 0
		if *e.Enabled {
			v = 1
		}
		args = append(args, v)
	}
	if e.Label != nil {
		sets = append(sets, "label = ?")
		args = append(args, *e.Label)
	}
	if e.UoM != nil {
		sets = append(sets, "uom = ?")
		args = append(args, *e.UoM)
	}
	if e.Decimals != nil {
		sets = append(sets, "decimals = ?")
		args = append(args, *e.Decimals)
	}
	if e.SortOrder != nil {
		sets = append(sets, "sort_order = ?")
		args = append(args, *e.SortOrder)
	}
	if len(sets) == 0 {
		return nil // nothing to change for this channel
	}

	query := "UPDATE profile_channels SET " + joinComma(sets) + " WHERE profile_id = ? AND channel_id = ?"
	args = append(args, profileID, e.ChannelID)
	res, err := tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("update channel %q: %w", e.ChannelID, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update channel %q rows: %w", e.ChannelID, err)
	}
	if n == 0 {
		return fmt.Errorf("update channel %q: no such channel in active profile", e.ChannelID)
	}
	return nil
}

// joinComma joins SET fragments with ", " (avoids importing strings for one call).
func joinComma(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	return out
}

// ResetActiveProfileChannels replaces the active profile's channels with a fresh
// copy of the supplied vocabulary (all enabled, sort_order = slice order) — the
// operator escape hatch. Done in one transaction. main supplies the vocab from the
// active DaqFormat (the store stays format-agnostic).
func (s *Store) ResetActiveProfileChannels(channels []SeedChannel) error {
	now := time.Now().UnixMicro()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("reset channels begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var id int64
	row := tx.QueryRow(`SELECT id FROM pump_profiles WHERE is_active = 1 LIMIT 1`)
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("reset channels: no active profile")
		}
		return fmt.Errorf("reset channels lookup: %w", err)
	}

	if _, err := tx.Exec(`DELETE FROM profile_channels WHERE profile_id = ?`, id); err != nil {
		return fmt.Errorf("reset channels delete: %w", err)
	}
	if err := insertChannels(tx, id, channels); err != nil {
		return err
	}
	if _, err := tx.Exec(
		`UPDATE pump_profiles SET updated_at_us = ? WHERE id = ?`, now, id,
	); err != nil {
		return fmt.Errorf("reset channels touch: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("reset channels commit: %w", err)
	}
	return nil
}
