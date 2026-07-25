package core

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestNewTimestampRejectsZero(t *testing.T) {
	_, err := NewTimestamp(time.Time{})
	if !errors.Is(err, ErrInvalidTimestamp) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTimestamp)
	}
}

func TestTimestampUTC(t *testing.T) {
	tm := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	ts, err := NewTimestamp(tm)
	if err != nil {
		t.Fatal(err)
	}
	if ts.Time().Location() != time.UTC {
		t.Errorf("location = %v, want UTC", ts.Time().Location())
	}
	if got := ts.String(); got != "2026-01-15T12:00:00Z" {
		t.Errorf("String() = %q, want %q", got, "2026-01-15T12:00:00Z")
	}
}

func TestTimestampNonUTCOffsetPreserved(t *testing.T) {
	loc := time.FixedZone("UTC+3", 3*60*60)
	tm := time.Date(2026, 1, 15, 15, 0, 0, 0, loc)
	ts, err := NewTimestamp(tm)
	if err != nil {
		t.Fatal(err)
	}
	if _, offset := ts.Time().Zone(); offset != 3*60*60 {
		t.Errorf("offset = %d, want %d", offset, 3*60*60)
	}
	if got := ts.String(); got != "2026-01-15T15:00:00+03:00" {
		t.Errorf("String() = %q, want %q", got, "2026-01-15T15:00:00+03:00")
	}
}

func TestTimestampNanosecondPrecisionPreserved(t *testing.T) {
	tm := time.Date(2026, 1, 15, 12, 0, 0, 123456789, time.UTC)
	ts, err := NewTimestamp(tm)
	if err != nil {
		t.Fatal(err)
	}
	roundTripped, err := ParseTimestamp(ts.String())
	if err != nil {
		t.Fatal(err)
	}
	if roundTripped.Time().Nanosecond() != 123456789 {
		t.Errorf("nanosecond = %d, want %d", roundTripped.Time().Nanosecond(), 123456789)
	}
}

func TestTimestampOrdering(t *testing.T) {
	earlier, err := NewTimestamp(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	later, err := NewTimestamp(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !earlier.Before(later) {
		t.Error("earlier.Before(later) = false, want true")
	}
	if !later.After(earlier) {
		t.Error("later.After(earlier) = false, want true")
	}
	if earlier.Compare(later) != -1 {
		t.Errorf("earlier.Compare(later) = %d, want -1", earlier.Compare(later))
	}

	// Ordering is independent of the recorded offset: the same instant
	// expressed in two different zones must compare equal.
	loc := time.FixedZone("UTC+2", 2*60*60)
	sameInstantDifferentZone, err := NewTimestamp(time.Date(2026, 1, 1, 2, 0, 0, 0, loc))
	if err != nil {
		t.Fatal(err)
	}
	if !earlier.Equal(sameInstantDifferentZone) {
		t.Error("same instant in different offsets did not compare equal")
	}
}

func TestTimestampJSONRoundTrip(t *testing.T) {
	original, err := NewTimestamp(time.Date(2026, 3, 4, 5, 6, 7, 890000000, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Timestamp
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(original) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, original)
	}
}

func TestTimestampJSONRejectsZero(t *testing.T) {
	var ts Timestamp
	if err := json.Unmarshal([]byte(`""`), &ts); err == nil {
		t.Error("Unmarshal empty string succeeded, want error")
	}
}

func TestParseTimestampRejectsMalformed(t *testing.T) {
	if _, err := ParseTimestamp("not-a-timestamp"); !errors.Is(err, ErrInvalidTimestamp) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTimestamp)
	}
}
