package postgres_test

import (
	"context"
	"testing"

	"avito/internal/domain"
	"avito/internal/repository/postgres"
)

func TestTeamRepository_CreateAndGet(t *testing.T) {
	if testDB == nil {
		t.Skip("Database not available")
	}
	truncateTables(t)

	ctx := context.Background()
	repo := postgres.NewTeamRepository(testDB.DB)

	team := &domain.Team{
		Name: "Test Team Alpha",
	}

	err := repo.Create(ctx, team)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if team.ID == 0 {
		t.Error("Expected team ID to be set after creation")
	}

	gotTeam, err := repo.Get(ctx, team.Name)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if gotTeam.ID != team.ID {
		t.Errorf("Get() ID = %v, want %v", gotTeam.ID, team.ID)
	}
	if gotTeam.Name != team.Name {
		t.Errorf("Get() Name = %v, want %v", gotTeam.Name, team.Name)
	}

	exists, err := repo.Exists(ctx, team.Name)
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Exists() = false, want true")
	}
}

func TestTeamRepository_Duplicate(t *testing.T) {
	if testDB == nil {
		t.Skip("Database not available")
	}
	truncateTables(t)
	ctx := context.Background()
	repo := postgres.NewTeamRepository(testDB.DB)

	team := &domain.Team{Name: "Unique Team"}

	if err := repo.Create(ctx, team); err != nil {
		t.Fatalf("First Create() error = %v", err)
	}

	err := repo.Create(ctx, team)
	if err == nil {
		t.Fatal("Expected error on duplicate team name, got nil")
	}
	if err != domain.ErrTeamExists {
		t.Errorf("Expected ErrTeamExists, got %v", err)
	}
}
