package core

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Extension is a namespaced container for Product-specific data attached
// to a PEOS construct. PEOS-owned fields are never placed inside an
// Extension; this container exists precisely so that Product contract
// data stays structurally separate from normative PEOS fields instead of
// being smuggled into them.
//
// The zero Extension (returned by NewExtension) and an Extension that has
// had entries added and later would be empty are treated identically as
// "no Product-specific data present" by IsZero; this package does not
// distinguish an explicitly-declared-empty extension container from an
// absent one.
//
// Payloads are stored and returned as defensive copies: mutating a
// []byte obtained from Get, or a json.RawMessage passed into With, after
// the call returns never affects the Extension's internal state.
type Extension struct {
	data map[string]json.RawMessage
}

// NewExtension returns an empty Extension.
func NewExtension() Extension { return Extension{} }

// With returns a copy of e with namespace set to payload. namespace must
// be non-empty and must not already be present in e; use a fresh
// Extension (or Without, if ever needed by a later packet) to replace an
// existing namespace's payload rather than silently overwriting it.
// payload must be non-empty and syntactically valid JSON.
func (e Extension) With(namespace string, payload json.RawMessage) (Extension, error) {
	ns, err := normalizeIdentityValue(namespace)
	if err != nil {
		return Extension{}, fmt.Errorf("core: Extension.With: %w", err)
	}
	if len(payload) == 0 || !json.Valid(payload) {
		return Extension{}, fmt.Errorf("core: Extension.With: payload must be non-empty, valid JSON")
	}
	if _, exists := e.data[ns]; exists {
		return Extension{}, fmt.Errorf("core: Extension.With: namespace %q: %w", ns, ErrDuplicateExtensionNamespace)
	}
	result := Extension{data: make(map[string]json.RawMessage, len(e.data)+1)}
	for k, v := range e.data {
		result.data[k] = copyRawMessage(v)
	}
	result.data[ns] = copyRawMessage(payload)
	return result, nil
}

// Get returns a defensive copy of the payload stored under namespace, and
// whether it was present.
func (e Extension) Get(namespace string) (json.RawMessage, bool) {
	v, ok := e.data[namespace]
	if !ok {
		return nil, false
	}
	return copyRawMessage(v), true
}

// Namespaces returns the extension's namespaces in sorted order.
func (e Extension) Namespaces() []string {
	names := make([]string, 0, len(e.data))
	for k := range e.data {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// IsZero reports whether e carries no Product-specific data.
func (e Extension) IsZero() bool { return len(e.data) == 0 }

func copyRawMessage(v json.RawMessage) json.RawMessage {
	cp := make(json.RawMessage, len(v))
	copy(cp, v)
	return cp
}

// MarshalJSON encodes e as a JSON object keyed by namespace. Go's
// encoding/json marshals map[string]* keys in sorted order, which keeps
// this output deterministic.
func (e Extension) MarshalJSON() ([]byte, error) {
	if e.data == nil {
		return json.Marshal(map[string]json.RawMessage{})
	}
	return json.Marshal(e.data)
}

// UnmarshalJSON decodes e from a JSON object keyed by namespace. An empty
// key is rejected.
func (e *Extension) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal Extension: %w", err)
	}
	result := Extension{data: make(map[string]json.RawMessage, len(raw))}
	for k, v := range raw {
		if k == "" {
			return fmt.Errorf("core: unmarshal Extension: empty namespace key")
		}
		result.data[k] = copyRawMessage(v)
	}
	*e = result
	return nil
}
