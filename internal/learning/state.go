package learning

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	StatusActive   = "active"
	StatusComplete = "complete"
	StatusSkipped  = "skipped"
)

type State struct {
	Active *Route `json:"active,omitempty"`
}

type Route struct {
	FullName     string `json:"full_name"`
	Slug         string `json:"slug"`
	ObsidianPath string `json:"obsidian_path"`
	GeneratedAt  string `json:"generated_at"`
	CurrentDay   int    `json:"current_day"`
	Day1Due      string `json:"day1_due"`
	Day2Due      string `json:"day2_due"`
	Status       string `json:"status"`
}

func LoadState(path string) (State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return State{}, nil
		}
		return State{}, fmt.Errorf("read learning state: %w", err)
	}
	if len(data) == 0 {
		return State{}, nil
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}, fmt.Errorf("parse learning state: %w", err)
	}
	return state, nil
}

func SaveState(path string, state State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create learning state dir: %w", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal learning state: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write learning state: %w", err)
	}
	return nil
}

func (s State) NeedsNewProject(now time.Time) bool {
	if s.Active == nil {
		return true
	}
	if s.Active.Status != "" && s.Active.Status != StatusActive {
		return true
	}
	return s.CurrentRouteDay(now) == 0
}

func (s State) CurrentRouteDay(now time.Time) int {
	if s.Active == nil {
		return 0
	}
	if s.Active.Status != "" && s.Active.Status != StatusActive {
		return 0
	}
	today := dateString(now)
	if today <= s.Active.Day1Due {
		return 1
	}
	if today <= s.Active.Day2Due {
		return 2
	}
	return 0
}

func dateString(t time.Time) string {
	if t.IsZero() {
		t = time.Now().UTC()
	}
	return t.Format("2006-01-02")
}
