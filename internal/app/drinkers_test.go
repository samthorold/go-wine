package app_test

import (
	"context"
	"errors"
	"testing"

	"go-wine/internal/adapter/memory"
	"go-wine/internal/app"
	"go-wine/internal/domain"
)

func TestCreateDrinker_PersistsAndReturnsID(t *testing.T) {
	drinkers := memory.NewDrinkerRepo()
	h := app.NewCreateDrinkerHandler(drinkers)

	id, err := h.Handle(context.Background(), app.CreateDrinkerCommand{Name: "Sam"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := drinkers.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Sam" {
		t.Errorf("Name = %q, want Sam", got.Name)
	}
}

func TestCreateDrinker_RejectsEmptyName(t *testing.T) {
	h := app.NewCreateDrinkerHandler(memory.NewDrinkerRepo())

	if _, err := h.Handle(context.Background(), app.CreateDrinkerCommand{Name: ""}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("empty name: err = %v, want ErrValidation", err)
	}
}

func TestRenameDrinker_UpdatesExisting(t *testing.T) {
	drinkers := memory.NewDrinkerRepo()
	d, _ := domain.NewDrinker("Sam")
	_ = drinkers.Save(context.Background(), d)
	h := app.NewRenameDrinkerHandler(drinkers)

	if err := h.Handle(context.Background(), app.RenameDrinkerCommand{ID: d.ID, Name: "Samuel"}); err != nil {
		t.Fatalf("rename: %v", err)
	}

	got, _ := drinkers.Get(context.Background(), d.ID)
	if got.Name != "Samuel" {
		t.Errorf("Name = %q, want Samuel", got.Name)
	}
}

func TestRenameDrinker_RejectsEmptyName(t *testing.T) {
	drinkers := memory.NewDrinkerRepo()
	d, _ := domain.NewDrinker("Sam")
	_ = drinkers.Save(context.Background(), d)
	h := app.NewRenameDrinkerHandler(drinkers)

	if err := h.Handle(context.Background(), app.RenameDrinkerCommand{ID: d.ID, Name: ""}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("empty name: err = %v, want ErrValidation", err)
	}
	got, _ := drinkers.Get(context.Background(), d.ID)
	if got.Name != "Sam" {
		t.Errorf("a rejected rename must not persist: Name = %q, want Sam", got.Name)
	}
}

func TestRenameDrinker_UnknownIsNotFound(t *testing.T) {
	h := app.NewRenameDrinkerHandler(memory.NewDrinkerRepo())

	if err := h.Handle(context.Background(), app.RenameDrinkerCommand{ID: domain.NewID(), Name: "Sam"}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unknown drinker: err = %v, want ErrNotFound", err)
	}
}
