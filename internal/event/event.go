package event

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

const MaxBody = 256 * 1024

// CoreEvent is a signed field custody message posted by a drill crew.
type CoreEvent struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Interval describes a drilled core depth range awaiting lab custody transfer.
type Interval struct {
	HoleID    string  `json:"hole_id"`
	FromDepth float64 `json:"from_depth_m"`
	ToDepth   float64 `json:"to_depth_m"`
	BoxLabel  string  `json:"box_label"`
	Lithology string  `json:"lithology,omitempty"`
}

// GammaScan is a downhole natural-gamma reading attached to a core interval.
type GammaScan struct {
	HoleID    string  `json:"hole_id"`
	DepthM    float64 `json:"depth_m"`
	CPS       float64 `json:"cps"`
	ProbeID   string  `json:"probe_id"`
	Operator  string  `json:"operator,omitempty"`
	Notes     string  `json:"notes,omitempty"`
}

// Envelope is retained as an alias for older call sites; prefer CoreEvent.
type Envelope = CoreEvent

func Parse(body []byte) (CoreEvent, error) {
	if len(body) == 0 {
		return CoreEvent{}, fmt.Errorf("empty body")
	}
	if len(body) > MaxBody {
		return CoreEvent{}, fmt.Errorf("body %d exceeds %d bytes", len(body), MaxBody)
	}
	var env CoreEvent
	if err := json.Unmarshal(body, &env); err != nil {
		return CoreEvent{}, fmt.Errorf("json: %v", err)
	}
	if err := ValidateType(env.Type); err != nil {
		return CoreEvent{}, err
	}
	if len(env.Payload) == 0 || string(env.Payload) == "null" {
		return CoreEvent{}, fmt.Errorf("payload is required")
	}
	if !json.Valid(env.Payload) {
		return CoreEvent{}, fmt.Errorf("payload is not valid json")
	}
	switch {
	case strings.HasPrefix(env.Type, "core.interval"):
		if _, err := DecodeInterval(env.Payload); err != nil {
			return CoreEvent{}, err
		}
	case strings.HasPrefix(env.Type, "core.gamma"):
		if _, err := DecodeGammaScan(env.Payload); err != nil {
			return CoreEvent{}, err
		}
	}
	return env, nil
}

func DecodeInterval(raw json.RawMessage) (Interval, error) {
	var iv Interval
	if err := json.Unmarshal(raw, &iv); err != nil {
		return Interval{}, fmt.Errorf("interval payload: %w", err)
	}
	if strings.TrimSpace(iv.HoleID) == "" {
		return Interval{}, fmt.Errorf("interval hole_id is required")
	}
	if iv.ToDepth < iv.FromDepth {
		return Interval{}, fmt.Errorf("interval to_depth_m must be >= from_depth_m")
	}
	if strings.TrimSpace(iv.BoxLabel) == "" {
		return Interval{}, fmt.Errorf("interval box_label is required")
	}
	return iv, nil
}

func DecodeGammaScan(raw json.RawMessage) (GammaScan, error) {
	var g GammaScan
	if err := json.Unmarshal(raw, &g); err != nil {
		return GammaScan{}, fmt.Errorf("gamma payload: %w", err)
	}
	if strings.TrimSpace(g.HoleID) == "" {
		return GammaScan{}, fmt.Errorf("gamma hole_id is required")
	}
	if g.DepthM < 0 {
		return GammaScan{}, fmt.Errorf("gamma depth_m must be >= 0")
	}
	if g.CPS < 0 {
		return GammaScan{}, fmt.Errorf("gamma cps must be >= 0")
	}
	if strings.TrimSpace(g.ProbeID) == "" {
		return GammaScan{}, fmt.Errorf("gamma probe_id is required")
	}
	return g, nil
}

func ValidateType(t string) error {
	t = strings.TrimSpace(t)
	if t == "" {
		return fmt.Errorf("event type is required")
	}
	if len(t) > 128 {
		return fmt.Errorf("event type too long")
	}
	parts := strings.Split(t, ".")
	if len(parts) < 1 {
		return fmt.Errorf("event type is empty")
	}
	for _, p := range parts {
		if p == "" {
			return fmt.Errorf("event type has empty segment")
		}
		for _, r := range p {
			if unicode.IsUpper(r) {
				return fmt.Errorf("event type must be lowercase")
			}
			ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-'
			if !ok {
				return fmt.Errorf("event type has illegal character %q", r)
			}
		}
	}
	return nil
}

func MatchPrefix(eventType, prefix string) bool {
	if prefix == "" {
		return true
	}
	if prefix == eventType {
		return true
	}
	if strings.HasSuffix(prefix, ".") {
		return strings.HasPrefix(eventType, prefix)
	}
	return eventType == prefix || strings.HasPrefix(eventType, prefix+".")
}
