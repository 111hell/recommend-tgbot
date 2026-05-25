package recommend

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type History struct {
	Items []HistoryItem `json:"items"`
}

type HistoryItem struct {
	FullName      string     `json:"full_name"`
	RecommendedAt string     `json:"recommended_at"`
	Score         float64    `json:"score"`
	Categories    []Category `json:"categories"`
}

func LoadHistory(path string) (History, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return History{Items: []HistoryItem{}}, nil
		}
		return History{}, fmt.Errorf("read history: %w", err)
	}
	if len(data) == 0 {
		return History{Items: []HistoryItem{}}, nil
	}

	var history History
	if err := json.Unmarshal(data, &history); err != nil {
		return History{}, fmt.Errorf("parse history: %w", err)
	}
	if history.Items == nil {
		history.Items = []HistoryItem{}
	}
	return history, nil
}

func SaveHistory(path string, history History) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create history dir: %w", err)
	}
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal history: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write history: %w", err)
	}
	return nil
}

func (h History) WasRecommendedWithin(fullName string, now time.Time, days int) bool {
	for _, item := range h.Items {
		if item.FullName != fullName {
			continue
		}
		recommendedAt, err := time.Parse("2006-01-02", item.RecommendedAt)
		if err != nil {
			continue
		}
		if now.Sub(recommendedAt).Hours()/24 <= float64(days) {
			return true
		}
	}
	return false
}

func (h *History) Add(recommendations []Recommendation, now time.Time) {
	date := now.Format("2006-01-02")
	for _, recommendation := range recommendations {
		h.Items = append(h.Items, HistoryItem{
			FullName:      recommendation.Repository.FullName,
			RecommendedAt: date,
			Score:         recommendation.Score,
			Categories:    recommendation.Categories,
		})
	}
}
