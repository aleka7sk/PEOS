package core

import (
	"encoding/json"
	"fmt"
	"time"
)

// Timestamp is the PEOS timestamp value type used wherever a
// specification requires a recorded point in time (Artifact Revision
// recording, State Assignment, every immutable record family, Claim
// timestamps, and so on).
//
// This package never generates a Timestamp from the wall clock; every
// constructor takes a time.Time supplied by the caller. Determining "now"
// is the caller's responsibility, not this package's, so that record
// construction stays deterministic and testable.
//
// The zero Timestamp is invalid for any normative use. The original
// timezone offset supplied by the caller is preserved as given (this
// package does not canonicalize timestamps to UTC); two Timestamps
// representing the same instant in different offsets compare as equal
// but are not required to have an identical string form.
type Timestamp struct {
	t time.Time
}

// NewTimestamp validates t and returns a Timestamp. A zero time.Time is
// rejected.
func NewTimestamp(t time.Time) (Timestamp, error) {
	if t.IsZero() {
		return Timestamp{}, fmt.Errorf("core: NewTimestamp: %w", ErrInvalidTimestamp)
	}
	return Timestamp{t: t}, nil
}

// ParseTimestamp parses an RFC 3339 (with optional fractional seconds)
// string, as produced by Timestamp.String, and returns a Timestamp.
func ParseTimestamp(s string) (Timestamp, error) {
	parsed, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return Timestamp{}, fmt.Errorf("core: ParseTimestamp: %w: %v", ErrInvalidTimestamp, err)
	}
	return NewTimestamp(parsed)
}

// Time returns the underlying time.Time, with its original offset intact.
func (ts Timestamp) Time() time.Time { return ts.t }

// IsZero reports whether ts is the zero value.
func (ts Timestamp) IsZero() bool { return ts.t.IsZero() }

// String returns the RFC 3339 form with fractional seconds preserved to
// nanosecond precision (trailing zero fractional digits are omitted by
// time.RFC3339Nano, not truncated non-zero ones).
func (ts Timestamp) String() string { return ts.t.Format(time.RFC3339Nano) }

// Before reports whether ts represents an instant strictly before other,
// independent of either value's timezone offset.
func (ts Timestamp) Before(other Timestamp) bool { return ts.t.Before(other.t) }

// After reports whether ts represents an instant strictly after other,
// independent of either value's timezone offset.
func (ts Timestamp) After(other Timestamp) bool { return ts.t.After(other.t) }

// Equal reports whether ts and other represent the same instant,
// independent of either value's timezone offset.
func (ts Timestamp) Equal(other Timestamp) bool { return ts.t.Equal(other.t) }

// Compare returns -1 if ts is before other, 0 if they represent the same
// instant, and +1 if ts is after other.
func (ts Timestamp) Compare(other Timestamp) int { return ts.t.Compare(other.t) }

// MarshalJSON encodes ts using its RFC 3339 string form.
func (ts Timestamp) MarshalJSON() ([]byte, error) { return json.Marshal(ts.String()) }

// UnmarshalJSON decodes ts from its RFC 3339 string form.
func (ts *Timestamp) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("core: unmarshal Timestamp: %w", err)
	}
	parsed, err := ParseTimestamp(s)
	if err != nil {
		return err
	}
	*ts = parsed
	return nil
}
