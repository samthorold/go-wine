package app

import (
	"context"
	"time"

	"go-wine/internal/domain"
)

// LogTastingCommand is the input to logging a Tasting. Rating is a plain int
// here; the handler turns it into the validated domain.Rating value object.
type LogTastingCommand struct {
	DrinkerID  domain.ID
	WineID     domain.ID
	Vintage    *int
	Rating     int
	Note       string
	Companions []domain.ID
	DrunkOn    time.Time
}

// LogTastingHandler is a command-side use case: it goes through the domain
// because logging carries real invariants (rating range, referenced Drinker and
// Wine must exist).
type LogTastingHandler struct {
	drinkers domain.DrinkerRepository
	wines    domain.WineRepository
	tastings domain.TastingRepository
}

func NewLogTastingHandler(d domain.DrinkerRepository, w domain.WineRepository, t domain.TastingRepository) *LogTastingHandler {
	return &LogTastingHandler{drinkers: d, wines: w, tastings: t}
}

// Handle validates the command and persists a new Tasting, returning its ID.
func (h *LogTastingHandler) Handle(ctx context.Context, cmd LogTastingCommand) (domain.ID, error) {
	if _, err := h.drinkers.Get(ctx, cmd.DrinkerID); err != nil {
		return "", err
	}
	if _, err := h.wines.Get(ctx, cmd.WineID); err != nil {
		return "", err
	}

	rating, err := domain.NewRating(cmd.Rating)
	if err != nil {
		return "", err
	}

	drunkOn := cmd.DrunkOn
	if drunkOn.IsZero() {
		drunkOn = time.Now()
	}

	t, err := domain.NewTasting(cmd.DrinkerID, cmd.WineID, cmd.Vintage, rating, cmd.Note, cmd.Companions, drunkOn)
	if err != nil {
		return "", err
	}

	if err := h.tastings.Add(ctx, t); err != nil {
		return "", err
	}
	return t.ID, nil
}
