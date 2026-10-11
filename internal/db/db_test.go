package db

import (
	"context"
	"os"
	"testing"

	"ZaViBiS/dont-worry-about-this-repo/internal/config"
)

func TestDBInitAndOperations(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	dbPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer func() {
		_ = os.Remove(dbPath)
	}()

	cfg := config.Config{DatabasePath: dbPath}
	database, err := DBInit(cfg)
	if err != nil {
		t.Fatalf("DBInit failed: %v", err)
	}
	defer func() {
		_ = database.Close()
	}()

	ctx := context.Background()

	// Initial count should be 0
	records, err := GetAll(ctx, database)
	if err != nil {
		t.Fatalf("GetAll empty failed: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected 0 records, got %d", len(records))
	}

	// Add records
	r1, err := Add(ctx, database)
	if err != nil {
		t.Fatalf("Add 1 failed: %v", err)
	}
	if r1.ID != 1 {
		t.Fatalf("expected ID 1, got %d", r1.ID)
	}

	r2, err := Add(ctx, database)
	if err != nil {
		t.Fatalf("Add 2 failed: %v", err)
	}
	if r2.ID != 2 {
		t.Fatalf("expected ID 2, got %d", r2.ID)
	}

	records, err = GetAll(ctx, database)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	if records[0].ID != 1 || records[1].ID != 2 {
		t.Fatalf("unexpected record IDs: %v", records)
	}

	// Delete record 1
	deleted, err := Delete(ctx, database, 1)
	if err != nil {
		t.Fatalf("Delete record 1 failed: %v", err)
	}
	if !deleted {
		t.Fatal("expected record 1 to be deleted")
	}

	// Verify only record 2 remains
	records, err = GetAll(ctx, database)
	if err != nil {
		t.Fatalf("GetAll after delete failed: %v", err)
	}
	if len(records) != 1 || records[0].ID != 2 {
		t.Fatalf("expected only record 2, got %v", records)
	}

	// Attempt to delete already deleted record 1
	deleted, err = Delete(ctx, database, 1)
	if err != nil {
		t.Fatalf("Delete non-existing record failed: %v", err)
	}
	if deleted {
		t.Fatal("expected false for non-existing record")
	}
}
