package app_test

import (
	"context"
	"testing"

	"go-wine/internal/adapter/memory"
	"go-wine/internal/app"
	"go-wine/internal/domain"
)

func TestListVarieties_ReturnsViewsSortedByName(t *testing.T) {
	repo := memory.NewVarietyRepo()
	for _, name := range []string{"Shiraz", "Chardonnay", "Pinot Noir"} {
		v, _ := domain.NewVariety(name)
		repo.Save(v)
	}

	h := app.NewListVarietiesHandler(repo)
	views, err := h.Handle(context.Background())
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	got := make([]string, len(views))
	for i, v := range views {
		got[i] = v.Name
	}
	want := []string{"Chardonnay", "Pinot Noir", "Shiraz"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("names = %v, want %v", got, want)
		}
	}
}
