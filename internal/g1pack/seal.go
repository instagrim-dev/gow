package g1pack

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	SealSchema   = "g1-pack-seal/1"
	MaxSealBytes = 64 << 10
	SealScope    = "metadata-only seal; protected execution remains unauthorized; custody remains unverified"
)

// Seal binds exact metadata bytes to the limited validation result available
// locally. It does not contain, seal, or validate protected task material.
type Seal struct {
	Schema         string     `json:"schema"`
	CreatedAt      string     `json:"created_at"`
	ManifestSHA256 string     `json:"manifest_sha256"`
	ManifestBytes  int        `json:"manifest_bytes"`
	PackID         string     `json:"pack_id"`
	Validation     Validation `json:"validation"`
	Scope          string     `json:"scope"`
}

// DecodeSeal refuses an ambiguous or incomplete receipt before inspection.
func DecodeSeal(raw []byte) (Seal, error) {
	var s Seal
	if len(raw) == 0 {
		return s, fmt.Errorf("G1 pack seal is empty")
	}
	if len(raw) > MaxSealBytes {
		return s, fmt.Errorf("G1 pack seal exceeds %d bytes", MaxSealBytes)
	}
	if !utf8.Valid(raw) {
		return s, fmt.Errorf("G1 pack seal must be valid UTF-8")
	}
	if err := strictSealKeys(raw); err != nil {
		return s, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&s); err != nil {
		return s, err
	}
	if _, err := d.Token(); err != io.EOF {
		return s, fmt.Errorf("G1 pack seal must contain exactly one JSON object")
	}
	if err := s.Validate(); err != nil {
		return s, err
	}
	return s, nil
}

func (s Seal) Validate() error {
	if s.Schema != SealSchema {
		return fmt.Errorf("unsupported G1 pack seal schema %q", s.Schema)
	}
	if _, err := time.Parse(time.RFC3339Nano, s.CreatedAt); err != nil {
		return fmt.Errorf("seal created_at must be RFC3339: %w", err)
	}
	if !sha256Hex.MatchString(s.ManifestSHA256) || s.ManifestBytes < 1 || strings.TrimSpace(s.PackID) == "" || len(s.PackID) > 256 {
		return fmt.Errorf("seal requires a lower-case manifest sha256, positive manifest_bytes, and pack_id")
	}
	if !s.Validation.StructurallyValid || s.Validation.Readiness != PreparedNotAuthorized || s.Validation.ProtectedExecutionAuthorized || s.Validation.CustodyVerified || !contains(s.Validation.Blockers, NoProtectedExecution) || !contains(s.Validation.Blockers, CustodyNotIndependently) {
		return fmt.Errorf("seal validation must retain the non-authorizing, custody-unverified boundary")
	}
	if s.Scope != SealScope {
		return fmt.Errorf("seal scope must retain the metadata-only boundary")
	}
	return nil
}

// MatchesManifest compares the receipt against exact input bytes and the
// decoded pack identity. A match proves only that this receipt names these
// metadata bytes; it does not prove custody, authorization, or task contents.
func (s Seal) MatchesManifest(m Manifest, raw []byte) bool {
	return s.ManifestSHA256 == Digest(raw) && s.ManifestBytes == len(raw) && s.PackID == m.PackID
}

func Digest(raw []byte) string {
	// Keep digest ownership with this package so seal inspection and creation
	// cannot drift in their binding algorithm.
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func strictSealKeys(raw []byte) error {
	allowed := map[string]map[string]bool{
		"root":       {"schema": true, "created_at": true, "manifest_sha256": true, "manifest_bytes": true, "pack_id": true, "validation": true, "scope": true},
		"validation": {"structurally_valid": true, "readiness": true, "protected_execution_authorized": true, "custody_verified": true, "blockers": true},
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var walk func(kind string, depth int) error
	walk = func(kind string, depth int) error {
		if depth > 4 {
			return fmt.Errorf("G1 pack seal JSON exceeds nesting ceiling")
		}
		tok, err := d.Token()
		if err != nil {
			return err
		}
		delim, ok := tok.(json.Delim)
		if !ok || delim != '{' {
			return fmt.Errorf("G1 pack seal %s must be an object", kind)
		}
		seen := map[string]bool{}
		for d.More() {
			t, err := d.Token()
			if err != nil {
				return err
			}
			key, ok := t.(string)
			if !ok || !allowed[kind][key] || seen[key] {
				return fmt.Errorf("unknown, case-variant, or duplicate G1 pack seal %s key %q", kind, key)
			}
			seen[key] = true
			if kind == "root" && key == "validation" {
				if err := walk("validation", depth+1); err != nil {
					return err
				}
				continue
			}
			if kind == "validation" && key == "blockers" {
				tok, err := d.Token()
				if err != nil {
					return err
				}
				a, ok := tok.(json.Delim)
				if !ok || a != '[' {
					return fmt.Errorf("G1 pack seal validation blockers must be an array")
				}
				for d.More() {
					tok, err := d.Token()
					if err != nil {
						return err
					}
					if _, ok := tok.(string); !ok {
						return fmt.Errorf("G1 pack seal validation blockers must contain strings")
					}
				}
				if _, err := d.Token(); err != nil {
					return err
				}
				continue
			}
			var discard any
			if err := d.Decode(&discard); err != nil {
				return err
			}
		}
		if len(seen) != len(allowed[kind]) {
			return fmt.Errorf("G1 pack seal %s is missing one or more required keys", kind)
		}
		_, err = d.Token()
		return err
	}
	if err := walk("root", 0); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return fmt.Errorf("G1 pack seal must contain exactly one JSON object")
	}
	return nil
}
