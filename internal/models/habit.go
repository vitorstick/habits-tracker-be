// Package models holds structs that mirror the database tables and API request/response shapes.
// The `json:"..."` tags tell Go's encoding/json how to serialize field names in JSON.
package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// FrequencyDetails holds flexible frequency options from the frontend (e.g. weekly days or monthly day).
// Stored as JSONB in the DB; the frontend can send any shape, e.g.:
//   {"type":"weekly","days":["monday","wednesday"]} or {"type":"monthly","dayOfMonth":15}
// We pass it through as raw JSON so the backend doesn't constrain the UI.
type FrequencyDetails json.RawMessage

// Scan implements sql.Scanner so we can read JSONB from PostgreSQL into this type.
func (f *FrequencyDetails) Scan(value interface{}) error {
	if value == nil {
		*f = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*f = append((*f)[0:0], v...)
		return nil
	case string:
		*f = []byte(v)
		return nil
	default:
		return fmt.Errorf("frequency_details: unsupported type %T", value)
	}
}

// Value implements driver.Valuer for optional INSERT/UPDATE of JSONB.
// Implemented on pointer so we can pass *FrequencyDetails from handlers and get NULL when nil.
func (f *FrequencyDetails) Value() (driver.Value, error) {
	if f == nil || len(*f) == 0 {
		return nil, nil
	}
	return []byte(*f), nil
}

// MarshalJSON outputs raw JSON so the API returns an object, not a string.
func (f *FrequencyDetails) MarshalJSON() ([]byte, error) {
	if f == nil || len(*f) == 0 {
		return []byte("null"), nil
	}
	return json.RawMessage(*f).MarshalJSON()
}

// UnmarshalJSON accepts whatever the frontend sends (object or null).
func (f *FrequencyDetails) UnmarshalJSON(data []byte) error {
	*f = FrequencyDetails(data)
	return nil
}

// Habit represents a single habit as returned by the API.
// It includes both DB columns and computed fields (Status, Streak, CompletedDates).
type Habit struct {
	ID               int              `json:"id"`
	Title            string           `json:"title"`
	Description      string           `json:"description,omitempty"`
	Icon             string           `json:"icon"`
	Color            string           `json:"color"`
	Frequency        string           `json:"frequency"`
	FrequencyDetails *FrequencyDetails `json:"frequencyDetails,omitempty"`
	Locked           bool             `json:"locked"`
	Status           string           `json:"status"`           // Computed: "completed" or "pending" (or "locked" if Locked)
	Streak           int              `json:"streak"`           // Computed: consecutive days completed
	CompletedDates   []string         `json:"completedDates"`   // Dates when the habit was completed (YYYY-MM-DD)
	CreatedAt        time.Time        `json:"-"`               // `-` means omit from JSON output
}

// CreateHabitRequest is the JSON body we expect when creating a new habit.
// frequencyDetails is flexible: frontend can send e.g. { "type": "weekly", "days": [1,3,5] } or { "type": "monthly", "dayOfMonth": 15 }.
// locked is optional; if omitted, defaults to false.
type CreateHabitRequest struct {
	Title            string           `json:"title"`
	Description      string           `json:"description"`
	Icon             string           `json:"icon"`
	Color            string           `json:"color"`
	Frequency        string           `json:"frequency"`
	FrequencyDetails *FrequencyDetails `json:"frequencyDetails,omitempty"`
	Locked           *bool            `json:"locked,omitempty"`
}
