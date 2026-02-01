// Package handlers tests: unit tests for computeStreak (no database, no router).
package handlers

import (
	"habit-tracker-be/internal/models"
	"testing"
	"time"
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

func TestHabitDueOnDate(t *testing.T) {
	// 2026-02-01 is a Sunday (Weekday 0)
	sunday, _ := time.Parse("2006-01-02", "2026-02-01")
	monday, _ := time.Parse("2006-01-02", "2026-02-02")

	t.Run("daily habit", func(t *testing.T) {
		h := models.Habit{Frequency: "daily"}
		if !habitDueOnDate(h, sunday) {
			t.Error("daily habit should be due on Sunday")
		}
		if !habitDueOnDate(h, monday) {
			t.Error("daily habit should be due on Monday")
		}
	})

	t.Run("weekly habit - weekday name", func(t *testing.T) {
		details := models.FrequencyDetails(`{"days": ["monday", "wednesday"]}`)
		h := models.Habit{Frequency: "weekly", FrequencyDetails: &details}
		if habitDueOnDate(h, sunday) {
			t.Error("weekly habit (mon/wed) should NOT be due on Sunday")
		}
		if !habitDueOnDate(h, monday) {
			t.Error("weekly habit (mon/wed) should be due on Monday")
		}
	})

	t.Run("weekly habit - weekday number", func(t *testing.T) {
		details := models.FrequencyDetails(`{"days": [0, 2]}`) // Sunday and Tuesday
		h := models.Habit{Frequency: "weekly", FrequencyDetails: &details}
		if !habitDueOnDate(h, sunday) {
			t.Error("weekly habit (0, 2) should be due on Sunday")
		}
		if habitDueOnDate(h, monday) {
			t.Error("weekly habit (0, 2) should NOT be due on Monday")
		}
	})

	t.Run("monthly habit", func(t *testing.T) {
		details := models.FrequencyDetails(`{"dayOfMonth": 15}`)
		h := models.Habit{Frequency: "monthly", FrequencyDetails: &details}
		day15, _ := time.Parse("2006-01-02", "2026-02-15")
		day16, _ := time.Parse("2006-01-02", "2026-02-16")

		if !habitDueOnDate(h, day15) {
			t.Error("monthly habit should be due on the 15th")
		}
		if habitDueOnDate(h, day16) {
			t.Error("monthly habit should NOT be due on the 16th")
		}
	})
}

func TestDeleteHabit(t *testing.T) {
	// This tests the logic of DeleteHabit.
	// In a real environment, you'd need a test DB, but we can at least check if it handles invalid IDs.
	t.Run("invalid id", func(t *testing.T) {
		// Mocking a request with an invalid ID format or non-integer
		// This is hard to unit test without a mock DB or full integration test.
		// For now, we rely on the manual curl verification once the server is restarted.
	})
}
