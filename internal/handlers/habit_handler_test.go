// Package handlers tests: unit tests for computeStreak (no database, no router).
package handlers

import (
	"testing"
)

func TestComputeStreak(t *testing.T) {
	tests := []struct {
		name           string
		completedDates []string
		today          string
		want           int
	}{
		{"empty dates", []string{}, "2024-01-15", 0},
		{"today only", []string{"2024-01-15"}, "2024-01-15", 1},
		{"today not done, yesterday done", []string{"2024-01-14"}, "2024-01-15", 1},
		{"three day streak including today", []string{"2024-01-13", "2024-01-14", "2024-01-15"}, "2024-01-15", 3},
		{"three day streak, today not done", []string{"2024-01-12", "2024-01-13", "2024-01-14"}, "2024-01-15", 3},
		{"gap breaks streak", []string{"2024-01-12", "2024-01-14", "2024-01-15"}, "2024-01-15", 2},
		{"future date in list ignored for streak", []string{"2024-01-14", "2024-01-15", "2024-01-16"}, "2024-01-15", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeStreak(tt.completedDates, tt.today)
			if got != tt.want {
				t.Errorf("computeStreak(%v, %q) = %d, want %d", tt.completedDates, tt.today, got, tt.want)
			}
		})
	}
}
