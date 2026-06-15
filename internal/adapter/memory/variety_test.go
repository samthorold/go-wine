package memory

import (
	"context"
	"errors"
	"testing"

	"go-wine/internal/domain"
)

func TestVarietyRepo_SaveListAndGet(t *testing.T) {
	repo := NewVarietyRepo()
	v, _ := domain.NewVariety("Shiraz")
	repo.Save(v)

	list, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || list[0].Name != "Shiraz" {
		t.Fatalf("List = %+v, want one Shiraz", list)
	}

	got, err := repo.Get(context.Background(), v.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != v {
		t.Errorf("Get = %+v, want %+v", got, v)
	}
}

func TestVarietyRepo_GetUnknownIsNotFound(t *testing.T) {
	repo := NewVarietyRepo()
	if _, err := repo.Get(context.Background(), domain.NewID()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

// Compile-time assertion that VarietyRepo satisfies the port.
var _ domain.VarietyRepository = (*VarietyRepo)(nil)
