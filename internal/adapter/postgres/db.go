// Package postgres is the production implementation of the domain repository
// ports, backed by Postgres via pgx. It is exercised by integration tests; the
// in-memory adapter backs fast unit tests.
package postgres

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"go-wine/internal/app"
	"go-wine/internal/domain"
	seedpkg "go-wine/internal/seed"
)

// Connect opens a pool and waits for the database to accept connections — useful
// under docker-compose where the app can start before Postgres is ready.
func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	var lastErr error
	for i := 0; i < 30; i++ {
		if lastErr = pool.Ping(ctx); lastErr == nil {
			return pool, nil
		}
		time.Sleep(time.Second)
	}
	pool.Close()
	return nil, fmt.Errorf("database not ready: %w", lastErr)
}

// Migrate applies any embedded migrations not yet recorded, in filename order.
func Migrate(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS) error {
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY)`); err != nil {
		return err
	}
	entries, err := fs.ReadDir(fsys, "migrations")
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		var applied bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, name).Scan(&applied); err != nil {
			return err
		}
		if applied {
			continue
		}
		body, err := fs.ReadFile(fsys, "migrations/"+name)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(body)); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
			return err
		}
	}
	return nil
}

// Seed inserts a couple of Drinkers and Wines on first run so the app is usable
// immediately. It is idempotent: it does nothing once the tables are populated.
func Seed(ctx context.Context, pool *pgxpool.Pool) error {
	var drinkerCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM drinkers`).Scan(&drinkerCount); err != nil {
		return err
	}
	if drinkerCount == 0 {
		for _, name := range []string{"Sam", "Partner"} {
			if _, err := pool.Exec(ctx, `INSERT INTO drinkers (id, name) VALUES ($1, $2)`, domain.NewID().String(), name); err != nil {
				return err
			}
		}
	}

	var wineCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM wines`).Scan(&wineCount); err != nil {
		return err
	}
	if wineCount == 0 {
		seed := []struct{ producer, name, style string }{
			{"Penfolds", "Bin 28 Shiraz", "Shiraz"},
			{"Cloudy Bay", "Sauvignon Blanc", "Sauvignon Blanc"},
			{"Guigal", "Côtes du Rhône Rouge", "GSM"},
		}
		for _, w := range seed {
			if _, err := pool.Exec(ctx, `INSERT INTO wines (id, producer, name, style) VALUES ($1, $2, $3, $4)`, domain.NewID().String(), w.producer, w.name, w.style); err != nil {
				return err
			}
		}
	}

	// Seed each Variety's intrinsic Characteristics through the domain seed-merge.
	// The grape rows themselves are created by migration 0002; here we merge the
	// rubric over whatever is stored, so a re-seed never clobbers a confirmed
	// value. Runs every startup (it is idempotent and non-clobbering), so a newly
	// added grape or a corrected rubric flows through to unconfirmed grapes.
	if err := app.SeedCharacteristics(ctx, NewVarietyRepo(pool), seedpkg.Characteristics()); err != nil {
		return err
	}
	return nil
}
