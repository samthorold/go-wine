package memory

import (
	"context"
	"errors"
	"testing"

	"go-wine/internal/domain"
)

func TestCompanionRepo_AddListScopedByDrinker(t *testing.T) {
	repo := NewCompanionRepo()
	sam := domain.NewID()
	other := domain.NewID()

	alex, _ := domain.NewCompanion(sam, "Alex")
	jo, _ := domain.NewCompanion(sam, "Jo")
	theirs, _ := domain.NewCompanion(other, "Pat")
	for _, c := range []domain.Companion{alex, jo, theirs} {
		if err := repo.Add(context.Background(), c); err != nil {
			t.Fatalf("Add: %v", err)
		}
	}

	got, err := repo.ListByDrinker(context.Background(), sam)
	if err != nil {
		t.Fatalf("ListByDrinker: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListByDrinker returned %d, want 2 (scoped to Sam)", len(got))
	}
	for _, c := range got {
		if c.DrinkerID != sam {
			t.Errorf("companion %q belongs to %q, want Sam's zone only", c.Name, c.DrinkerID)
		}
	}
}

func TestCompanionRepo_GetUnknownIsNotFound(t *testing.T) {
	repo := NewCompanionRepo()
	if _, err := repo.Get(context.Background(), domain.NewID()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

// Compile-time assertion that CompanionRepo satisfies the port.
var _ domain.CompanionRepository = (*CompanionRepo)(nil)
