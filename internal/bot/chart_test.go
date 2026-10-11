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

	t.Run("empty records returns error", func(t *testing.T) {
		_, err := generateChart(nil, loc)
		if err == nil {
			t.Fatal("expected error for empty records")
		}
	})

	t.Run("single record renders successfully", func(t *testing.T) {
		records := []db.Record{
			{ID: 1, Timestamp: time.Now()},
		}
		data, err := generateChart(records, loc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(data) == 0 {
			t.Fatal("expected non-empty png bytes")
		}
	})

	t.Run("multiple records across days renders successfully", func(t *testing.T) {
		now := time.Now()
		records := []db.Record{
			{ID: 1, Timestamp: now.Add(-5 * 24 * time.Hour)},
			{ID: 2, Timestamp: now.Add(-3 * 24 * time.Hour)},
			{ID: 3, Timestamp: now.Add(-3 * 24 * time.Hour)},
			{ID: 4, Timestamp: now.Add(-1 * 24 * time.Hour)},
			{ID: 5, Timestamp: now},
		}
		data, err := generateChart(records, loc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(data) == 0 {
			t.Fatal("expected non-empty png bytes")
		}
	})

	t.Run("records spanning over 40 days renders successfully capped to 30", func(t *testing.T) {
		now := time.Now()
		records := []db.Record{
			{ID: 1, Timestamp: now.Add(-40 * 24 * time.Hour)},
			{ID: 2, Timestamp: now.Add(-10 * 24 * time.Hour)},
			{ID: 3, Timestamp: now},
		}
		data, err := generateChart(records, loc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(data) == 0 {
			t.Fatal("expected non-empty png bytes")
		}
	})
}
