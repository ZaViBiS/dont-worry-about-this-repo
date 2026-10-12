package bot

import (
	"testing"
	"time"

	"ZaViBiS/dont-worry-about-this-repo/internal/db"
)

func TestGenerateChart(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("empty records returns error when no date range", func(t *testing.T) {
		_, _, err := generateChart(nil, loc)
		if err == nil {
			t.Fatal("expected error for empty records")
		}
	})

	t.Run("single record renders successfully for full month", func(t *testing.T) {
		now := time.Now().In(loc)
		records := []db.Record{
			{ID: 1, Timestamp: time.Date(now.Year(), now.Month(), 15, 12, 0, 0, 0, loc)},
		}
		data, total, err := generateChart(records, loc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(data) == 0 {
			t.Fatal("expected non-empty png bytes")
		}
		if total != 1 {
			t.Fatalf("expected total 1, got %d", total)
		}
	})

	t.Run("records in current month counted, records outside current month ignored", func(t *testing.T) {
		now := time.Now().In(loc)
		records := []db.Record{
			{ID: 1, Timestamp: time.Date(now.Year(), now.Month(), 1, 10, 0, 0, 0, loc)},
			{ID: 2, Timestamp: time.Date(now.Year(), now.Month(), 5, 10, 0, 0, 0, loc)},
			{ID: 3, Timestamp: time.Date(now.Year(), now.Month(), 20, 10, 0, 0, 0, loc)},
			{ID: 4, Timestamp: time.Date(now.Year(), now.Month()-1, 20, 10, 0, 0, 0, loc)}, // previous month
		}
		data, total, err := generateChart(records, loc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(data) == 0 {
			t.Fatal("expected non-empty png bytes")
		}
		if total != 3 {
			t.Fatalf("expected total 3, got %d", total)
		}
	})

	t.Run("explicit month date range renders successfully", func(t *testing.T) {
		startDate := time.Date(2025, time.December, 1, 0, 0, 0, 0, loc)
		endDate := time.Date(2025, time.December, 31, 0, 0, 0, 0, loc)
		records := []db.Record{
			{ID: 1, Timestamp: time.Date(2025, time.December, 10, 12, 0, 0, 0, loc)},
			{ID: 2, Timestamp: time.Date(2025, time.December, 20, 14, 0, 0, 0, loc)},
			{ID: 3, Timestamp: time.Date(2025, time.November, 20, 14, 0, 0, 0, loc)},
		}

		data, total, err := generateChart(records, loc, startDate, endDate)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(data) == 0 {
			t.Fatal("expected non-empty png bytes")
		}
		if total != 2 {
			t.Fatalf("expected total 2, got %d", total)
		}
	})
}

func TestParseDateRange(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}

	fixedNow := time.Date(2026, time.October, 12, 10, 0, 0, 0, loc)

	t.Run("no args defaults to current full month (1st to last day)", func(t *testing.T) {
		start, end, label, err := parseDateRange(nil, fixedNow, loc, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if label != "жовтень" {
			t.Errorf("expected label 'жовтень', got '%s'", label)
		}
		expectedStart := time.Date(2026, time.October, 1, 0, 0, 0, 0, loc)
		expectedEnd := time.Date(2026, time.October, 31, 0, 0, 0, 0, loc)
		if !start.Equal(expectedStart) || !end.Equal(expectedEnd) {
			t.Errorf("expected %v - %v, got %v - %v", expectedStart, expectedEnd, start, end)
		}
	})

	t.Run("specified month dec without year defaults to previous year if in future", func(t *testing.T) {
		start, end, label, err := parseDateRange([]string{"dec"}, fixedNow, loc, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if label != "грудень 2025" {
			t.Errorf("expected label 'грудень 2025', got '%s'", label)
		}
		expectedStart := time.Date(2025, time.December, 1, 0, 0, 0, 0, loc)
		expectedEnd := time.Date(2025, time.December, 31, 0, 0, 0, 0, loc)
		if !start.Equal(expectedStart) || !end.Equal(expectedEnd) {
			t.Errorf("expected %v - %v, got %v - %v", expectedStart, expectedEnd, start, end)
		}
	})

	t.Run("specified month in Ukrainian", func(t *testing.T) {
		start, end, label, err := parseDateRange([]string{"жовтень"}, fixedNow, loc, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if label != "жовтень" {
			t.Errorf("expected label 'жовтень', got '%s'", label)
		}
		expectedStart := time.Date(2026, time.October, 1, 0, 0, 0, 0, loc)
		expectedEnd := time.Date(2026, time.October, 31, 0, 0, 0, 0, loc)
		if !start.Equal(expectedStart) || !end.Equal(expectedEnd) {
			t.Errorf("expected %v - %v, got %v - %v", expectedStart, expectedEnd, start, end)
		}
	})

	t.Run("numeric month 10", func(t *testing.T) {
		start, end, label, err := parseDateRange([]string{"10"}, fixedNow, loc, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if label != "жовтень" {
			t.Errorf("expected label 'жовтень', got '%s'", label)
		}
		expectedStart := time.Date(2026, time.October, 1, 0, 0, 0, 0, loc)
		expectedEnd := time.Date(2026, time.October, 31, 0, 0, 0, 0, loc)
		if !start.Equal(expectedStart) || !end.Equal(expectedEnd) {
			t.Errorf("expected %v - %v, got %v - %v", expectedStart, expectedEnd, start, end)
		}
	})

	t.Run("month with explicit year dec 2026", func(t *testing.T) {
		start, end, label, err := parseDateRange([]string{"dec", "2026"}, fixedNow, loc, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if label != "грудень" {
			t.Errorf("expected label 'грудень', got '%s'", label)
		}
		expectedStart := time.Date(2026, time.December, 1, 0, 0, 0, 0, loc)
		expectedEnd := time.Date(2026, time.December, 31, 0, 0, 0, 0, loc)
		if !start.Equal(expectedStart) || !end.Equal(expectedEnd) {
			t.Errorf("expected %v - %v, got %v - %v", expectedStart, expectedEnd, start, end)
		}
	})

	t.Run("format YYYY-MM", func(t *testing.T) {
		start, end, label, err := parseDateRange([]string{"2025-12"}, fixedNow, loc, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if label != "грудень 2025" {
			t.Errorf("expected label 'грудень 2025', got '%s'", label)
		}
		expectedStart := time.Date(2025, time.December, 1, 0, 0, 0, 0, loc)
		expectedEnd := time.Date(2025, time.December, 31, 0, 0, 0, 0, loc)
		if !start.Equal(expectedStart) || !end.Equal(expectedEnd) {
			t.Errorf("expected %v - %v, got %v - %v", expectedStart, expectedEnd, start, end)
		}
	})

	t.Run("format MM.YYYY", func(t *testing.T) {
		start, end, label, err := parseDateRange([]string{"12.2025"}, fixedNow, loc, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if label != "грудень 2025" {
			t.Errorf("expected label 'грудень 2025', got '%s'", label)
		}
		expectedStart := time.Date(2025, time.December, 1, 0, 0, 0, 0, loc)
		expectedEnd := time.Date(2025, time.December, 31, 0, 0, 0, 0, loc)
		if !start.Equal(expectedStart) || !end.Equal(expectedEnd) {
			t.Errorf("expected %v - %v, got %v - %v", expectedStart, expectedEnd, start, end)
		}
	})

	t.Run("invalid month returns error", func(t *testing.T) {
		_, _, _, err := parseDateRange([]string{"invalid_month"}, fixedNow, loc, nil)
		if err == nil {
			t.Fatal("expected error for invalid month")
		}
	})
}
