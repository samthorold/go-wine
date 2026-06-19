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

// Save upserts a Drinker keyed by ID: a fresh Drinker inserts; a rename updates
// the existing row's name in place.
func (r *DrinkerRepo) Save(ctx context.Context, d domain.Drinker) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO drinkers (id, name) VALUES ($1, $2)
		 ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name`,
		d.ID.String(), d.Name,
	)
	return err
}

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
	parts, err := r.compositionParts(ctx, domain.ID(idStr))
	if err != nil {
		return domain.Wine{}, err
	}
	return domain.Wine{ID: domain.ID(idStr), Producer: producer, Name: name, Style: style, Composition: domain.Composition{Parts: parts}}, nil
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Attach each Wine's Composition from the link table.
	for i := range out {
		parts, err := r.compositionParts(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Composition = domain.Composition{Parts: parts}
	}
	return out, nil
}

// compositionParts reads a Wine's Composition rows, ordered by descending share
// so the dominant Variety leads.
func (r *WineRepo) compositionParts(ctx context.Context, wineID domain.ID) ([]domain.CompositionPart, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT variety_id, proportion FROM wine_varieties WHERE wine_id=$1 ORDER BY proportion DESC`, wineID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.CompositionPart
	for rows.Next() {
		var vid string
		var prop int
		if err := rows.Scan(&vid, &prop); err != nil {
			return nil, err
		}
		out = append(out, domain.CompositionPart{VarietyID: domain.ID(vid), Proportion: prop})
	}
	return out, rows.Err()
}

// SetComposition replaces a Wine's Composition atomically: it clears the
// existing wine_varieties rows and inserts the new ones in a single
// transaction, so the Wine aggregate is never left with a half-written
// Composition. The caller has already validated the Composition through the
// domain.
func (r *WineRepo) SetComposition(ctx context.Context, wineID domain.ID, c domain.Composition) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM wine_varieties WHERE wine_id=$1`, wineID.String()); err != nil {
		return err
	}
	for _, p := range c.Parts {
		if _, err := tx.Exec(ctx,
			`INSERT INTO wine_varieties (wine_id, variety_id, proportion) VALUES ($1, $2, $3)`,
			wineID.String(), p.VarietyID.String(), p.Proportion,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// VarietyRepo implements domain.VarietyRepository.
type VarietyRepo struct{ pool *pgxpool.Pool }

func NewVarietyRepo(p *pgxpool.Pool) *VarietyRepo { return &VarietyRepo{pool: p} }

func (r *VarietyRepo) Get(ctx context.Context, id domain.ID) (domain.Variety, error) {
	var idStr, name string
	err := r.pool.QueryRow(ctx, `SELECT id, name FROM varieties WHERE id=$1`, id.String()).Scan(&idStr, &name)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Variety{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Variety{}, err
	}
	return domain.Variety{ID: domain.ID(idStr), Name: name}, nil
}

func (r *VarietyRepo) List(ctx context.Context) ([]domain.Variety, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name FROM varieties ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Variety
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		out = append(out, domain.Variety{ID: domain.ID(id), Name: name})
	}
	return out, rows.Err()
}

// GetCharacteristics returns a Variety's intrinsic Characteristics, or the zero
// bundle (IsZero) when none has been seeded yet — the absent-row case. The
// flavour-note tags are read from the child table.
func (r *VarietyRepo) GetCharacteristics(ctx context.Context, id domain.ID) (domain.Characteristics, error) {
	var body, tannin, acidity, sweetness, alcohol int
	var prov string
	err := r.pool.QueryRow(ctx,
		`SELECT body, tannin, acidity, sweetness, alcohol, provenance
		 FROM variety_characteristics WHERE variety_id=$1`, id.String()).
		Scan(&body, &tannin, &acidity, &sweetness, &alcohol, &prov)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Characteristics{}, nil
	}
	if err != nil {
		return domain.Characteristics{}, err
	}
	notes, err := r.notes(ctx, id)
	if err != nil {
		return domain.Characteristics{}, err
	}
	return domain.NewCharacteristics(body, tannin, acidity, sweetness, alcohol, notes, provenanceFromText(prov))
}

func (r *VarietyRepo) notes(ctx context.Context, id domain.ID) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT note FROM variety_notes WHERE variety_id=$1 ORDER BY note`, id.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// SetCharacteristics replaces a Variety's Characteristics atomically: it upserts
// the 1:1 axes/provenance row and rewrites the flavour-note rows in a single
// transaction, so the aggregate is never half-written. The caller has already
// run the value through the domain seed-merge, so the no-clobber rule is honoured
// before we reach here. An unknown Variety is ErrNotFound.
func (r *VarietyRepo) SetCharacteristics(ctx context.Context, id domain.ID, c domain.Characteristics) error {
	var exists bool
	if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM varieties WHERE id=$1)`, id.String()).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return domain.ErrNotFound
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx,
		`INSERT INTO variety_characteristics (variety_id, body, tannin, acidity, sweetness, alcohol, provenance)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (variety_id) DO UPDATE SET
		   body=EXCLUDED.body, tannin=EXCLUDED.tannin, acidity=EXCLUDED.acidity,
		   sweetness=EXCLUDED.sweetness, alcohol=EXCLUDED.alcohol, provenance=EXCLUDED.provenance`,
		id.String(), c.Body.Int(), c.Tannin.Int(), c.Acidity.Int(), c.Sweetness.Int(), c.Alcohol.Int(), provenanceText(c.Provenance),
	); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM variety_notes WHERE variety_id=$1`, id.String()); err != nil {
		return err
	}
	for _, n := range c.Notes {
		if _, err := tx.Exec(ctx, `INSERT INTO variety_notes (variety_id, note) VALUES ($1, $2)`, id.String(), n); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// provenanceText / provenanceFromText map the binary Provenance to and from its
// stored 'default'/'confirmed' text form.
func provenanceText(p domain.Provenance) string {
	if p == domain.ProvenanceConfirmed {
		return "confirmed"
	}
	return "default"
}

func provenanceFromText(s string) domain.Provenance {
	if s == "confirmed" {
		return domain.ProvenanceConfirmed
	}
	return domain.ProvenanceDefault
}

// TastingRepo implements domain.TastingRepository.
type TastingRepo struct{ pool *pgxpool.Pool }

func NewTastingRepo(p *pgxpool.Pool) *TastingRepo { return &TastingRepo{pool: p} }

// Add persists a Tasting and its Companion links atomically: the base row plus
// one tasting_companions row per Companion, in a single transaction, so a
// Tasting is never stored without the company it was logged with (or vice
// versa).
func (r *TastingRepo) Add(ctx context.Context, t domain.Tasting) error {
	var vintage sql.NullInt64
	if t.Vintage != nil {
		vintage = sql.NullInt64{Int64: int64(*t.Vintage), Valid: true}
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx,
		`INSERT INTO tastings (id, drinker_id, wine_id, vintage, rating, note, drunk_on)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		t.ID.String(), t.DrinkerID.String(), t.WineID.String(), vintage, t.Rating.Int(), t.Note, t.DrunkOn,
	); err != nil {
		return err
	}

	for _, cid := range t.Companions {
		if _, err := tx.Exec(ctx,
			`INSERT INTO tasting_companions (tasting_id, companion_id) VALUES ($1, $2)`,
			t.ID.String(), cid.String(),
		); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Attach each Tasting's Companion IDs from the link table.
	for i := range out {
		cids, err := r.companionIDs(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Companions = cids
	}
	return out, nil
}

func (r *TastingRepo) companionIDs(ctx context.Context, tastingID domain.ID) ([]domain.ID, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT companion_id FROM tasting_companions WHERE tasting_id=$1`, tastingID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ID
	for rows.Next() {
		var cid string
		if err := rows.Scan(&cid); err != nil {
			return nil, err
		}
		out = append(out, domain.ID(cid))
	}
	return out, rows.Err()
}

// CompanionRepo implements domain.CompanionRepository.
type CompanionRepo struct{ pool *pgxpool.Pool }

func NewCompanionRepo(p *pgxpool.Pool) *CompanionRepo { return &CompanionRepo{pool: p} }

func (r *CompanionRepo) Add(ctx context.Context, c domain.Companion) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO companions (id, drinker_id, name) VALUES ($1, $2, $3)`,
		c.ID.String(), c.DrinkerID.String(), c.Name,
	)
	return err
}

func (r *CompanionRepo) Get(ctx context.Context, id domain.ID) (domain.Companion, error) {
	var idStr, did, name string
	err := r.pool.QueryRow(ctx,
		`SELECT id, drinker_id, name FROM companions WHERE id=$1`, id.String()).Scan(&idStr, &did, &name)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Companion{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Companion{}, err
	}
	return domain.Companion{ID: domain.ID(idStr), DrinkerID: domain.ID(did), Name: name}, nil
}

func (r *CompanionRepo) ListByDrinker(ctx context.Context, drinkerID domain.ID) ([]domain.Companion, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, drinker_id, name FROM companions WHERE drinker_id=$1 ORDER BY name`, drinkerID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Companion
	for rows.Next() {
		var id, did, name string
		if err := rows.Scan(&id, &did, &name); err != nil {
			return nil, err
		}
		out = append(out, domain.Companion{ID: domain.ID(id), DrinkerID: domain.ID(did), Name: name})
	}
	return out, rows.Err()
}
