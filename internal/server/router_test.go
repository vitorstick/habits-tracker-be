// Package server_test runs HTTP tests against the router (no import cycle with handlers).
package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"habit-tracker-be/internal/database"
	"habit-tracker-be/internal/server"
)

func initDBForTest(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration tests")
	}
	database.Connect(dbURL)
}

func TestGetHabits_Integration(t *testing.T) {
	initDBForTest(t)
	defer database.DB.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/habits", nil)
	rec := httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /api/habits status = %d, want 200", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var habits []map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&habits); err != nil {
		t.Errorf("response body is not JSON array: %v", err)
	}
}

func TestCreateHabit_Integration(t *testing.T) {
	initDBForTest(t)
	defer database.DB.Close()

	body := []byte(`{"title":"Test Habit","color":"#ff0000"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/habits", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("POST /api/habits status = %d, want 201. body: %s", rec.Code, rec.Body.Bytes())
	}
	var res map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("response body is not JSON: %v", err)
	}
	if _, ok := res["id"]; !ok {
		t.Errorf("response missing id")
	}
	if res["title"] != "Test Habit" {
		t.Errorf("response title = %v, want Test Habit", res["title"])
	}
}

func TestCreateHabit_InvalidJSON_Returns400(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/habits", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("POST /api/habits with invalid JSON status = %d, want 400", rec.Code)
	}
}

func TestToggleHabitLog_InvalidID_Returns400(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/habits/notanid/log", nil)
	rec := httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("POST /api/habits/notanid/log status = %d, want 400", rec.Code)
	}
}

func TestToggleHabitLog_InvalidDate_Returns400(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/habits/1/log?date=invalid", nil)
	rec := httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("POST /api/habits/1/log?date=invalid status = %d, want 400", rec.Code)
	}
}
