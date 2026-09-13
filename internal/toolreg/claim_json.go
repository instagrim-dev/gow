package toolreg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// encoding/json accepts duplicate and case-variant struct keys. Refuse both
// before decoding so the exact bound input has one interpretation.
// StrictKeys is shared by every strict data-only input format; callers own
// their allowed key set and nesting ceiling.
func StrictKeys(raw []byte, label string, keys []string, maxDepth int) error {
	allowed := map[string]bool{}
	for _, k := range keys {
		allowed[k] = true
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var value func(int) error
	value = func(depth int) error {
		if depth > maxDepth {
			return fmt.Errorf("%s JSON exceeds nesting ceiling", label)
		}
		token, err := d.Token()
		if err != nil {
			return err
		}
		delim, compound := token.(json.Delim)
		if !compound {
			return nil
		}
		switch delim {
		case '{':
			seen := map[string]bool{}
			for d.More() {
				t, err := d.Token()
				if err != nil {
					return err
				}
				key, ok := t.(string)
				if !ok || !allowed[key] || seen[key] {
					return fmt.Errorf("unknown, case-variant, or duplicate %s key %q", label, key)
				}
				seen[key] = true
				if err := value(depth + 1); err != nil {
					return err
				}
			}
		case '[':
			for d.More() {
				if err := value(depth + 1); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unexpected delimiter %q", delim)
		}
		_, err = d.Token()
		return err
	}
	if err := value(0); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return fmt.Errorf("%s must contain exactly one JSON object", label)
	}
	return nil
}
