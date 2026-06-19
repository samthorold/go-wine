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

func TestVarietyRepo_SetAndGetCharacteristics(t *testing.T) {
	repo := NewVarietyRepo()
	v, _ := domain.NewVariety("Shiraz")
	repo.Save(v)

	c, _ := domain.NewCharacteristics(5, 4, 2, 1, 5, []string{"blackberry", "pepper"}, domain.ProvenanceDefault)
	if err := repo.SetCharacteristics(context.Background(), v.ID, c); err != nil {
		t.Fatalf("SetCharacteristics: %v", err)
	}

	got, err := repo.GetCharacteristics(context.Background(), v.ID)
	if err != nil {
		t.Fatalf("GetCharacteristics: %v", err)
	}
	if got.Body.Int() != 5 || got.Provenance != domain.ProvenanceDefault {
		t.Errorf("GetCharacteristics = %+v, want the saved bundle", got)
	}
}

func TestVarietyRepo_CharacteristicsAbsentIsZero(t *testing.T) {
	repo := NewVarietyRepo()
	v, _ := domain.NewVariety("Shiraz")
	repo.Save(v)

	got, err := repo.GetCharacteristics(context.Background(), v.ID)
	if err != nil {
		t.Fatalf("GetCharacteristics: %v", err)
	}
	if !got.IsZero() {
		t.Errorf("characteristics for an unseeded Variety should be the zero bundle; got %+v", got)
	}
}

func TestVarietyRepo_SetCharacteristicsUnknownVarietyIsNotFound(t *testing.T) {
	repo := NewVarietyRepo()
	c, _ := domain.NewCharacteristics(3, 3, 3, 3, 3, nil, domain.ProvenanceDefault)
	if err := repo.SetCharacteristics(context.Background(), domain.NewID(), c); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

// Compile-time assertion that VarietyRepo satisfies the port.
var _ domain.VarietyRepository = (*VarietyRepo)(nil)
