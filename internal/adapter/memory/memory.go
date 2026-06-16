// Package memory holds in-memory implementations of the domain repository
// ports. They back fast, containerless unit tests; the Postgres adapter is the
// production counterpart exercised by integration tests.
package memory

import (
	"context"
	"sync"

	"go-wine/internal/domain"
)

// DrinkerRepo is an in-memory domain.DrinkerRepository.
type DrinkerRepo struct {
	mu   sync.RWMutex
	data map[domain.ID]domain.Drinker
}

func NewDrinkerRepo() *DrinkerRepo {
	return &DrinkerRepo{data: make(map[domain.ID]domain.Drinker)}
}

func (r *DrinkerRepo) Save(_ context.Context, d domain.Drinker) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[d.ID] = d
	return nil
}

func (r *DrinkerRepo) Get(_ context.Context, id domain.ID) (domain.Drinker, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.data[id]
	if !ok {
		return domain.Drinker{}, domain.ErrNotFound
	}
	return d, nil
}

func (r *DrinkerRepo) List(_ context.Context) ([]domain.Drinker, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Drinker, 0, len(r.data))
	for _, d := range r.data {
		out = append(out, d)
	}
	return out, nil
}

// WineRepo is an in-memory domain.WineRepository.
type WineRepo struct {
	mu   sync.RWMutex
	data map[domain.ID]domain.Wine
}

func NewWineRepo() *WineRepo {
	return &WineRepo{data: make(map[domain.ID]domain.Wine)}
}

func (r *WineRepo) Save(w domain.Wine) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[w.ID] = w
}

// SetComposition replaces the stored Wine's Composition, keeping the Wine and
// its Composition together in one aggregate, matching the Postgres adapter's
// transactional write.
func (r *WineRepo) SetComposition(_ context.Context, wineID domain.ID, c domain.Composition) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	w, ok := r.data[wineID]
	if !ok {
		return domain.ErrNotFound
	}
	w.Composition = c
	r.data[wineID] = w
	return nil
}

func (r *WineRepo) Get(_ context.Context, id domain.ID) (domain.Wine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.data[id]
	if !ok {
		return domain.Wine{}, domain.ErrNotFound
	}
	return w, nil
}

func (r *WineRepo) List(_ context.Context) ([]domain.Wine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Wine, 0, len(r.data))
	for _, w := range r.data {
		out = append(out, w)
	}
	return out, nil
}

// VarietyRepo is an in-memory domain.VarietyRepository.
type VarietyRepo struct {
	mu   sync.RWMutex
	data map[domain.ID]domain.Variety
}

func NewVarietyRepo() *VarietyRepo {
	return &VarietyRepo{data: make(map[domain.ID]domain.Variety)}
}

func (r *VarietyRepo) Save(v domain.Variety) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[v.ID] = v
}

func (r *VarietyRepo) Get(_ context.Context, id domain.ID) (domain.Variety, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.data[id]
	if !ok {
		return domain.Variety{}, domain.ErrNotFound
	}
	return v, nil
}

func (r *VarietyRepo) List(_ context.Context) ([]domain.Variety, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Variety, 0, len(r.data))
	for _, v := range r.data {
		out = append(out, v)
	}
	return out, nil
}

// TastingRepo is an in-memory domain.TastingRepository.
type TastingRepo struct {
	mu   sync.RWMutex
	data map[domain.ID]domain.Tasting
}

func NewTastingRepo() *TastingRepo {
	return &TastingRepo{data: make(map[domain.ID]domain.Tasting)}
}

func (r *TastingRepo) Add(_ context.Context, t domain.Tasting) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[t.ID] = t
	return nil
}

func (r *TastingRepo) ListByDrinker(_ context.Context, drinkerID domain.ID) ([]domain.Tasting, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Tasting, 0)
	for _, t := range r.data {
		if t.DrinkerID == drinkerID {
			out = append(out, t)
		}
	}
	return out, nil
}

// CompanionRepo is an in-memory domain.CompanionRepository.
type CompanionRepo struct {
	mu   sync.RWMutex
	data map[domain.ID]domain.Companion
}

func NewCompanionRepo() *CompanionRepo {
	return &CompanionRepo{data: make(map[domain.ID]domain.Companion)}
}

func (r *CompanionRepo) Add(_ context.Context, c domain.Companion) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[c.ID] = c
	return nil
}

func (r *CompanionRepo) Get(_ context.Context, id domain.ID) (domain.Companion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.data[id]
	if !ok {
		return domain.Companion{}, domain.ErrNotFound
	}
	return c, nil
}

func (r *CompanionRepo) ListByDrinker(_ context.Context, drinkerID domain.ID) ([]domain.Companion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Companion, 0)
	for _, c := range r.data {
		if c.DrinkerID == drinkerID {
			out = append(out, c)
		}
	}
	return out, nil
}
