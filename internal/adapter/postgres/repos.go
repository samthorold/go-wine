package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"go-wine/internal/domain"
)

// DrinkerRepo implements domain.DrinkerRepository.
type DrinkerRepo struct{ pool *pgxpool.Pool }

func NewDrinkerRepo(p *pgxpool.Pool) *DrinkerRepo { return &DrinkerRepo{pool: p} }

func (r *DrinkerRepo) Get(ctx context.Context, id domain.ID) (domain.Drinker, error) {
	var idStr, name string
	err := r.pool.QueryRow(ctx, `SELECT id, name FROM drinkers WHERE id=$1`, id.String()).Scan(&idStr, &name)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Drinker{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Drinker{}, err
	}
	return domain.Drinker{ID: domain.ID(idStr), Name: name}, nil
}

func (r *DrinkerRepo) List(ctx context.Context) ([]domain.Drinker, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name FROM drinkers ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Drinker
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		out = append(out, domain.Drinker{ID: domain.ID(id), Name: name})
	}
	return out, rows.Err()
}

// WineRepo implements domain.WineRepository.
type WineRepo struct{ pool *pgxpool.Pool }

func NewWineRepo(p *pgxpool.Pool) *WineRepo { return &WineRepo{pool: p} }

func (r *WineRepo) Get(ctx context.Context, id domain.ID) (domain.Wine, error) {
	var idStr, producer, name, style string
	err := r.pool.QueryRow(ctx, `SELECT id, producer, name, style FROM wines WHERE id=$1`, id.String()).Scan(&idStr, &producer, &name, &style)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Wine{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Wine{}, err
	}
	return domain.Wine{ID: domain.ID(idStr), Producer: producer, Name: name, Style: style}, nil
}

func (r *WineRepo) List(ctx context.Context) ([]domain.Wine, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, producer, name, style FROM wines ORDER BY producer, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Wine
	for rows.Next() {
		var id, producer, name, style string
		if err := rows.Scan(&id, &producer, &name, &style); err != nil {
			return nil, err
		}
		out = append(out, domain.Wine{ID: domain.ID(id), Producer: producer, Name: name, Style: style})
	}
	return out, rows.Err()
}

// TastingRepo implements domain.TastingRepository.
type TastingRepo struct{ pool *pgxpool.Pool }

func NewTastingRepo(p *pgxpool.Pool) *TastingRepo { return &TastingRepo{pool: p} }

func (r *TastingRepo) Add(ctx context.Context, t domain.Tasting) error {
	var vintage sql.NullInt64
	if t.Vintage != nil {
		vintage = sql.NullInt64{Int64: int64(*t.Vintage), Valid: true}
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO tastings (id, drinker_id, wine_id, vintage, rating, note, drunk_on)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		t.ID.String(), t.DrinkerID.String(), t.WineID.String(), vintage, t.Rating.Int(), t.Note, t.DrunkOn,
	)
	return err
}

func (r *TastingRepo) ListByDrinker(ctx context.Context, drinkerID domain.ID) ([]domain.Tasting, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, drinker_id, wine_id, vintage, rating, note, drunk_on
		 FROM tastings WHERE drinker_id=$1 ORDER BY drunk_on DESC`, drinkerID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Tasting
	for rows.Next() {
		var (
			id, did, wid, note string
			rating             int
			vintage            sql.NullInt64
			drunkOn            time.Time
		)
		if err := rows.Scan(&id, &did, &wid, &vintage, &rating, &note, &drunkOn); err != nil {
			return nil, err
		}
		var vp *int
		if vintage.Valid {
			v := int(vintage.Int64)
			vp = &v
		}
		ratingVO, err := domain.NewRating(rating)
		if err != nil {
			return nil, err
		}
		out = append(out, domain.Tasting{
			ID:        domain.ID(id),
			DrinkerID: domain.ID(did),
			WineID:    domain.ID(wid),
			Vintage:   vp,
			Rating:    ratingVO,
			Note:      note,
			DrunkOn:   drunkOn,
		})
	}
	return out, rows.Err()
}
