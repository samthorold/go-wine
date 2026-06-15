// Command web is the go-wine HTTP server. With DATABASE_URL set it runs on
// Postgres (the docker-compose path); without it, it runs on an in-memory store
// seeded in process — handy for a quick local look without a database.
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	gowine "go-wine"
	"go-wine/internal/adapter/memory"
	"go-wine/internal/adapter/postgres"
	"go-wine/internal/adapter/web"
	"go-wine/internal/app"
	"go-wine/internal/domain"
)

func main() {
	ctx := context.Background()

	var (
		drinkers  domain.DrinkerRepository
		wines     domain.WineRepository
		tastings  domain.TastingRepository
		varieties domain.VarietyRepository
	)

	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		pool, err := postgres.Connect(ctx, dsn)
		if err != nil {
			log.Fatalf("connect: %v", err)
		}
		defer pool.Close()
		if err := postgres.Migrate(ctx, pool, gowine.Migrations); err != nil {
			log.Fatalf("migrate: %v", err)
		}
		if err := postgres.Seed(ctx, pool); err != nil {
			log.Fatalf("seed: %v", err)
		}
		drinkers = postgres.NewDrinkerRepo(pool)
		wines = postgres.NewWineRepo(pool)
		tastings = postgres.NewTastingRepo(pool)
		varieties = postgres.NewVarietyRepo(pool)
		log.Println("store: postgres")
	} else {
		md, mw, mt, mv := memory.NewDrinkerRepo(), memory.NewWineRepo(), memory.NewTastingRepo(), memory.NewVarietyRepo()
		seedMemory(md, mw, mv)
		drinkers, wines, tastings, varieties = md, mw, mt, mv
		log.Println("store: in-memory (set DATABASE_URL for Postgres)")
	}

	logH := app.NewLogTastingHandler(drinkers, wines, tastings)
	listH := app.NewListTastingsHandler(wines, tastings)
	listV := app.NewListVarietiesHandler(varieties)
	srv := web.NewServer(drinkers, wines, logH, listH, listV)

	addr := ":" + envOr("PORT", "8080")
	log.Printf("go-wine listening on %s", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatal(err)
	}
}

func seedMemory(drinkers *memory.DrinkerRepo, wines *memory.WineRepo, varieties *memory.VarietyRepo) {
	for _, name := range []string{"Sam", "Partner"} {
		if d, err := domain.NewDrinker(name); err == nil {
			drinkers.Save(d)
		}
	}
	seed := []struct{ producer, name, style string }{
		{"Penfolds", "Bin 28 Shiraz", "Shiraz"},
		{"Cloudy Bay", "Sauvignon Blanc", "Sauvignon Blanc"},
		{"Guigal", "Côtes du Rhône Rouge", "GSM"},
	}
	for _, w := range seed {
		if wine, err := domain.NewWine(w.producer, w.name, w.style); err == nil {
			wines.Save(wine)
		}
	}
	// A coherent starter set of common grapes, mirroring the 0002 migration seed
	// so the /varieties page is populated on the in-memory store too.
	for _, name := range []string{
		"Cabernet Sauvignon", "Merlot", "Pinot Noir", "Syrah", "Grenache",
		"Mourvèdre", "Tempranillo", "Sangiovese", "Nebbiolo", "Malbec",
		"Chardonnay", "Sauvignon Blanc", "Riesling", "Pinot Grigio", "Chenin Blanc",
	} {
		if v, err := domain.NewVariety(name); err == nil {
			varieties.Save(v)
		}
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
