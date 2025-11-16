package postgres_test

import (
	"context"
	"log"
	"os"
	"testing"

	"avito/internal/repository/postgres"

	_ "github.com/lib/pq"
)

var testDB *postgres.DB

func TestMain(m *testing.M) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5433/pr_reviewer?sslmode=disable"
	}

	cfg := postgres.Config{
		URL:            connStr,
		MaxConnections: 5,
		MaxIdle:        2,
	}

	var err error
	testDB, err = postgres.NewDB(cfg)
	if err != nil {
		log.Printf("Skipping repository tests: database not available at %s: %v", connStr, err)
		os.Exit(0)
	}

	if err := testDB.Ping(); err != nil {
		log.Printf("Skipping repository tests: database ping failed: %v", err)
		os.Exit(0)
	}

	code := m.Run()

	testDB.Close()
	os.Exit(code)
}

func truncateTables(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	queries := []string{
		"TRUNCATE TABLE batch_deactivate_tasks CASCADE",
		"TRUNCATE TABLE pr_reviewers CASCADE",
		"TRUNCATE TABLE pull_requests CASCADE",
		"TRUNCATE TABLE users CASCADE",
		"TRUNCATE TABLE teams CASCADE",
	}

	for _, q := range queries {
		if _, err := testDB.ExecContext(ctx, q); err != nil {
			t.Fatalf("Failed to truncate table: %v", err)
		}
	}
}
