package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Oleg2210/gophermart/internal/domain/user"
)

func TestMemoryRepositoryCreateSuccess(t *testing.T) {
	repo := NewMemoryRepository()

	u, err := repo.Create(context.Background(), "oleg", "hashed-pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if u.Login != "oleg" {
		t.Errorf("expected login 'oleg', got %s", u.Login)
	}

	if u.HashedPassword != "hashed-pass" {
		t.Errorf("unexpected hashed password")
	}

	if u.ID == "" {
		t.Errorf("expected non-empty ID")
	}
}

func TestMemoryRepositoryCreateDuplicateLogin(t *testing.T) {
	repo := NewMemoryRepository()

	_, err := repo.Create(context.Background(), "oleg", "pass1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = repo.Create(context.Background(), "oleg", "pass2")
	if !errors.Is(err, user.ErrLoginAlreadyExists) {
		t.Fatalf("expected ErrLoginAlreadyExists, got %v", err)
	}
}

func TestMemoryRepositoryGetByLoginSuccess(t *testing.T) {
	repo := NewMemoryRepository()

	created, _ := repo.Create(context.Background(), "oleg", "pass")

	got, err := repo.GetByLogin(context.Background(), "oleg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("expected same user ID")
	}
}

func TestMemoryRepositoryGetByLoginNotFound(t *testing.T) {
	repo := NewMemoryRepository()

	_, err := repo.GetByLogin(context.Background(), "unknown")
	if !errors.Is(err, user.ErrLoginDoesNotExist) {
		t.Fatalf("expected ErrLoginDoesNotExist, got %v", err)
	}
}

func TestMemoryRepositoryCreateContextCanceled(t *testing.T) {
	repo := NewMemoryRepository()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.Create(ctx, "oleg", "pass")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestMemoryRepositoryGetByLoginContextCanceled(t *testing.T) {
	repo := NewMemoryRepository()

	ctx, _ := context.WithTimeout(context.Background(), time.Nanosecond)
	time.Sleep(time.Millisecond)

	_, err := repo.GetByLogin(ctx, "oleg")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.DeadlineExceeded, got %v", err)
	}
}
