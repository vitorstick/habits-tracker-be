// Package handlers contains HTTP handlers for the habit tracker API.
// Each handler receives (w http.ResponseWriter, r *http.Request) - the standard Go HTTP signature.
package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"habit-tracker-be/internal/database"
	"habit-tracker-be/internal/models"

	"github.com/go-chi/chi/v5"
)

// DefaultUserID is used for v1 without auth. When you add Supabase Auth (Phase 6),
// replace this with the user_id from the JWT. Ensure you have at least one user
// in the `users` table (e.g. INSERT INTO users (id, email) VALUES (1, 'dev@example.com');).
const DefaultUserID = 1

// GetHabits returns habits for a specific date (defaults to today).
// Query param: date=YYYY-MM-DD
func GetHabits(w http.ResponseWriter, r *http.Request) {
	log.Println("[GetHabits] Handling GET /api/habits")

	// Parse optional date parameter; default to current local date.
	targetDateStr := r.URL.Query().Get("date")
	if targetDateStr == "" {
		targetDateStr = time.Now().Format("2006-01-02")
	}
	targetTime, err := time.Parse("2006-01-02", targetDateStr)
	if err != nil {
		log.Printf("[GetHabits] Invalid date parameter %q: %v", targetDateStr, err)
		http.Error(w, "Invalid date format (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	// Query habits for the default user (including description, frequency_details, locked).
	rows, err := database.DB.Query(r.Context(),
		`SELECT id, title, description, icon, color, frequency, frequency_details, locked
		 FROM habits WHERE user_id = $1 ORDER BY id`,
		DefaultUserID)
	if err != nil {
		log.Printf("[GetHabits] DB query error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// First pass: collect all habits and their IDs.
	var habits []models.Habit
	var habitIDs []int
	for rows.Next() {
		var h models.Habit
		var description *string
		if err := rows.Scan(&h.ID, &h.Title, &description, &h.Icon, &h.Color, &h.Frequency, &h.FrequencyDetails, &h.Locked); err != nil {
			log.Printf("[GetHabits] Row scan error for habit: %v", err)
			continue
		}
		if description != nil {
			h.Description = *description
		}
		habits = append(habits, h)
		habitIDs = append(habitIDs, h.ID)
	}

	if err := rows.Err(); err != nil {
		log.Printf("[GetHabits] Rows iteration error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Batch fetch all logs for all habits in a single query (fixes N+1 problem).
	logsByHabit, err := fetchCompletedDatesBatch(r.Context(), habitIDs)
	if err != nil {
		log.Printf("[GetHabits] Failed to batch fetch logs: %v", err)
		// Continue with empty logs rather than failing the entire request.
		logsByHabit = make(map[int][]string)
	}

	// Second pass: attach logs, compute status/streak, and filter by frequency.
	var filteredHabits []models.Habit
	for _, h := range habits {
		h.CompletedDates = logsByHabit[h.ID]
		if h.CompletedDates == nil {
			h.CompletedDates = []string{}
		}

		// Compute status relative to targetDateStr.
		if h.Locked {
			h.Status = "locked"
		} else {
			h.Status = "pending"
			for _, d := range h.CompletedDates {
				if d == targetDateStr {
					h.Status = "completed"
					break
				}
			}
		}

		// Compute streak relative to targetDateStr.
		h.Streak = computeStreak(h.CompletedDates, targetDateStr)

		// Filter: only include habits that are due on targetTime.
		if !habitDueOnDate(h, targetTime) {
			continue
		}
		filteredHabits = append(filteredHabits, h)
	}
	habits = filteredHabits

	// Return empty array instead of null when no habits (better for frontend).
	if habits == nil {
		habits = []models.Habit{}
	}

	log.Printf("[GetHabits] Returning %d habits", len(habits))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(habits); err != nil {
		log.Printf("[GetHabits] JSON encode error: %v", err)
	}
}

// fetchCompletedDates returns all completed_at dates for a habit as YYYY-MM-DD strings.
func fetchCompletedDates(ctx context.Context, habitID int) ([]string, error) {
	rows, err := database.DB.Query(ctx,
		"SELECT completed_at::text FROM habit_logs WHERE habit_id = $1 ORDER BY completed_at DESC",
		habitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			continue
		}
		// PostgreSQL may return "2024-01-15T00:00:00Z" - trim to date part if needed.
		if len(d) >= 10 {
			d = d[:10]
		}
		dates = append(dates, d)
	}
	return dates, rows.Err()
}

// fetchCompletedDatesBatch returns all completed_at dates for multiple habits in a single query.
// Returns a map of habitID -> []dates. This prevents N+1 queries when fetching logs for many habits.
func fetchCompletedDatesBatch(ctx context.Context, habitIDs []int) (map[int][]string, error) {
	if len(habitIDs) == 0 {
		return make(map[int][]string), nil
	}

	rows, err := database.DB.Query(ctx,
		`SELECT habit_id, completed_at::text
		 FROM habit_logs
		 WHERE habit_id = ANY($1)
		 ORDER BY habit_id, completed_at DESC`,
		habitIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logsByHabit := make(map[int][]string)
	for rows.Next() {
		var habitID int
		var d string
		if err := rows.Scan(&habitID, &d); err != nil {
			continue
		}
		// PostgreSQL may return "2024-01-15T00:00:00Z" - trim to date part if needed.
		if len(d) >= 10 {
			d = d[:10]
		}
		logsByHabit[habitID] = append(logsByHabit[habitID], d)
	}
	return logsByHabit, rows.Err()
}

// habitDueOnDate returns true if the habit should be shown on the given date based on frequency and frequencyDetails.
// - daily: always true.
// - weekly: true only if date's weekday is in frequencyDetails.days.
// - monthly: true only if date's day of month equals frequencyDetails.dayOfMonth.
func habitDueOnDate(h models.Habit, date time.Time) bool {
	switch h.Frequency {
	case "daily":
		return true
	case "weekly":
		return habitDueOnDateWeekly(h.FrequencyDetails, date)
	case "monthly":
		return habitDueOnDateMonthly(h.FrequencyDetails, date)
	default:
		return true
	}
}

func habitDueOnDateWeekly(details *models.FrequencyDetails, date time.Time) bool {
	if details == nil || len(*details) == 0 {
		return true
	}
	var parsed struct {
		Days []interface{} `json:"days"` // ["monday","wednesday"] or [0,1,2]
	}
	if err := json.Unmarshal(*details, &parsed); err != nil || len(parsed.Days) == 0 {
		return true
	}
	weekday := date.Weekday() // time.Sunday=0, Monday=1, ..., Saturday=6
	for _, d := range parsed.Days {
		switch v := d.(type) {
		case string:
			if weekdayStringMatches(weekday, v) {
				return true
			}
		case float64:
			if int(v) == int(weekday) {
				return true
			}
		}
	}
	return false
}

var weekdayNames = []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}

func weekdayStringMatches(weekday time.Weekday, s string) bool {
	name := weekdayNames[weekday]
	return len(s) > 0 && strings.EqualFold(s, name)
}

func habitDueOnDateMonthly(details *models.FrequencyDetails, date time.Time) bool {
	if details == nil || len(*details) == 0 {
		return true
	}
	var parsed struct {
		DayOfMonth *int `json:"dayOfMonth"`
	}
	if err := json.Unmarshal(*details, &parsed); err != nil || parsed.DayOfMonth == nil {
		return true
	}
	day := date.Day() // 1-31
	return day == *parsed.DayOfMonth
}

// computeStreak returns the current streak: consecutive days completed up to today.
// If today is completed, streak counts from today backwards; otherwise from yesterday.
func computeStreak(completedDates []string, today string) int {
	if len(completedDates) == 0 {
		return 0
	}
	// Build a set of completed dates for quick lookup.
	set := make(map[string]bool)
	for _, d := range completedDates {
		set[d] = true
	}

	// Decide end date: if today is completed, count from today; else from yesterday.
	end, _ := time.Parse("2006-01-02", today)
	if !set[today] {
		end = end.AddDate(0, 0, -1)
	}

	streak := 0
	for {
		key := end.Format("2006-01-02")
		if !set[key] {
			break
		}
		streak++
		end = end.AddDate(0, 0, -1)
	}
	return streak
}

// CreateHabit creates a new habit from the JSON body.
func CreateHabit(w http.ResponseWriter, r *http.Request) {
	log.Println("[CreateHabit] Handling POST /api/habits")

	var req models.CreateHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[CreateHabit] Invalid JSON body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Apply defaults if not provided (Go zero value for string is "").
	if req.Color == "" {
		req.Color = "#58cc02"
	}
	if req.Frequency == "" {
		req.Frequency = "daily"
	}
	locked := false
	if req.Locked != nil {
		locked = *req.Locked
	}

	// Optional description: store NULL in DB when empty so frontend can distinguish omit vs "".
	var desc *string
	if req.Description != "" {
		desc = &req.Description
	}

	if req.FrequencyDetails != nil {
		log.Printf("[CreateHabit] FrequencyDetails: %s", string(*req.FrequencyDetails))
	}

	var id int
	err := database.DB.QueryRow(r.Context(),
		`INSERT INTO habits (user_id, title, description, icon, color, frequency, frequency_details, locked)
		 VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8) RETURNING id`,
		DefaultUserID, req.Title, desc, req.Icon, req.Color, req.Frequency,
		req.FrequencyDetails, locked).Scan(&id)
	if err != nil {
		log.Printf("[CreateHabit] INSERT error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[CreateHabit] Created habit id=%d title=%q", id, req.Title)
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{"id": id, "title": req.Title}); err != nil {
		log.Printf("[CreateHabit] JSON encode error: %v", err)
	}
}

// ToggleHabitLog toggles a completion log for a habit on a given date.
// If the log exists, it is deleted; otherwise it is inserted.
// Query param: date=YYYY-MM-DD (defaults to today).
func ToggleHabitLog(w http.ResponseWriter, r *http.Request) {
	habitIDStr := chi.URLParam(r, "id")
	log.Printf("[ToggleHabitLog] Handling POST /api/habits/%s/log", habitIDStr)

	habitID, err := strconv.Atoi(habitIDStr)
	if err != nil || habitID <= 0 {
		log.Printf("[ToggleHabitLog] Invalid habit id: %q", habitIDStr)
		http.Error(w, "Invalid habit id", http.StatusBadRequest)
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}
	// Basic date validation: must be YYYY-MM-DD.
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		log.Printf("[ToggleHabitLog] Invalid date: %q", dateStr)
		http.Error(w, "Invalid date (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	var exists bool
	err = database.DB.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM habit_logs WHERE habit_id = $1 AND completed_at = $2::date)",
		habitID, dateStr).Scan(&exists)
	if err != nil {
		log.Printf("[ToggleHabitLog] EXISTS query error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if exists {
		_, err = database.DB.Exec(r.Context(),
			"DELETE FROM habit_logs WHERE habit_id = $1 AND completed_at = $2::date",
			habitID, dateStr)
		if err != nil {
			log.Printf("[ToggleHabitLog] DELETE error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("[ToggleHabitLog] Removed log habit_id=%d date=%s", habitID, dateStr)
	} else {
		_, err = database.DB.Exec(r.Context(),
			"INSERT INTO habit_logs (habit_id, completed_at) VALUES ($1, $2::date)",
			habitID, dateStr)
		if err != nil {
			log.Printf("[ToggleHabitLog] INSERT error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("[ToggleHabitLog] Added log habit_id=%d date=%s", habitID, dateStr)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Toggled"}); err != nil {
		log.Printf("[ToggleHabitLog] JSON encode error: %v", err)
	}
}

// DeleteHabit deletes a habit and its logs (via CASCADE) for the default user.
func DeleteHabit(w http.ResponseWriter, r *http.Request) {
	habitIDStr := chi.URLParam(r, "id")
	log.Printf("[DeleteHabit] Handling DELETE /api/habits/%s", habitIDStr)

	habitID, err := strconv.Atoi(habitIDStr)
	if err != nil || habitID <= 0 {
		log.Printf("[DeleteHabit] Invalid habit id: %q", habitIDStr)
		http.Error(w, "Invalid habit id", http.StatusBadRequest)
		return
	}

	result, err := database.DB.Exec(r.Context(),
		"DELETE FROM habits WHERE id = $1 AND user_id = $2",
		habitID, DefaultUserID)
	if err != nil {
		log.Printf("[DeleteHabit] DELETE error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		log.Printf("[DeleteHabit] Habit %d not found or not owned by user", habitID)
		http.Error(w, "Habit not found", http.StatusNotFound)
		return
	}

	log.Printf("[DeleteHabit] Deleted habit id=%d", habitID)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Habit deleted"}); err != nil {
		log.Printf("[DeleteHabit] JSON encode error: %v", err)
	}
}
