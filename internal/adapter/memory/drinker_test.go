package memory

import (
	"context"
	"errors"
	"testing"

	"go-wine/internal/domain"
)

func TestDrinkerRepo_SaveThenGet(t *testing.T) {
	repo := NewDrinkerRepo()
	d, _ := domain.NewDrinker("Sam")

	if err := repo.Save(context.Background(), d); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.Get(context.Background(), d.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "Sam" {
		t.Errorf("Name = %q, want Sam", got.Name)
	}
}

func TestDrinkerRepo_SaveUpsertsOnRename(t *testing.T) {
	repo := NewDrinkerRepo()
	d, _ := domain.NewDrinker("Sam")
	if err := repo.Save(context.Background(), d); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := d.Rename("Samuel"); err != nil {
		t.Fatalf("rename: %v", err)
	}
	if err := repo.Save(context.Background(), d); err != nil {
		t.Fatalf("Save (rename): %v", err)
	}

	got, err := repo.Get(context.Background(), d.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "Samuel" {
		t.Errorf("rename did not persist: Name = %q, want Samuel", got.Name)
	}
	all, _ := repo.List(context.Background())
	if len(all) != 1 {
		t.Errorf("rename must update in place, not insert: List len = %d, want 1", len(all))
	}
}

func TestDrinkerRepo_GetUnknownIsNotFound(t *testing.T) {
	repo := NewDrinkerRepo()
	if _, err := repo.Get(context.Background(), domain.NewID()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

// Compile-time assertion that DrinkerRepo satisfies the port.
var _ domain.DrinkerRepository = (*DrinkerRepo)(nil)
