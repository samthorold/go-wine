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
	seedpkg "go-wine/internal/seed"
)

func main() {
	ctx := context.Background()

	var (
		drinkers   domain.DrinkerRepository
		wines      domain.WineRepository
		tastings   domain.TastingRepository
		varieties  domain.VarietyRepository
		companions domain.CompanionRepository
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
		companions = postgres.NewCompanionRepo(pool)
		log.Println("store: postgres")
	} else {
		md, mw, mt := memory.NewDrinkerRepo(), memory.NewWineRepo(), memory.NewTastingRepo()
		mv, mc := memory.NewVarietyRepo(), memory.NewCompanionRepo()
		seedMemory(md, mw, mv, mc)
		drinkers, wines, tastings, varieties, companions = md, mw, mt, mv, mc
		log.Println("store: in-memory (set DATABASE_URL for Postgres)")
	}

	logH := app.NewLogTastingHandler(drinkers, wines, tastings)
	listH := app.NewListTastingsHandler(wines, tastings, companions)
	listV := app.NewListVarietiesHandler(varieties)
	getV := app.NewGetVarietyHandler(varieties)
	editVC := app.NewEditCharacteristicsHandler(varieties)
	listW := app.NewListWinesHandler(wines)
	getW := app.NewGetWineHandler(wines, varieties)
	editC := app.NewEditCompositionHandler(wines, varieties)
	styleC := app.NewResolveStyleCompositionHandler(varieties, seedpkg.StyleCompositions())
	createD := app.NewCreateDrinkerHandler(drinkers)
	renameD := app.NewRenameDrinkerHandler(drinkers)
	prefs := app.NewPreferencesHandler(wines, varieties, tastings)
	srv := web.NewServer(drinkers, wines, varieties, companions, logH, listH, listV, getV, editVC, listW, getW, editC, styleC, createD, renameD, prefs)

	addr := ":" + envOr("PORT", "8080")
	log.Printf("go-wine listening on %s", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatal(err)
	}
}

func seedMemory(drinkers *memory.DrinkerRepo, wines *memory.WineRepo, varieties *memory.VarietyRepo, companions *memory.CompanionRepo) {
	for _, name := range []string{"Sam", "Partner"} {
		d, err := domain.NewDrinker(name)
		if err != nil {
			continue
		}
		_ = drinkers.Save(context.Background(), d)
		// A couple of Companions in each Drinker's personal zone so the picker is
		// populated. Scoped to the Drinker — never linked across owners.
		for _, cn := range []string{"Alex", "Jo"} {
			if c, err := domain.NewCompanion(d.ID, cn); err == nil {
				_ = companions.Add(context.Background(), c)
			}
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
	// Seed each grape's intrinsic Characteristics through the domain seed-merge,
	// so confirmed values would survive a re-seed exactly as on Postgres.
	if err := app.SeedCharacteristics(context.Background(), varieties, seedpkg.Characteristics()); err != nil {
		log.Printf("seed characteristics (memory): %v", err)
	}
	// Fill each Wine's default Composition from the Style → Composition map through
	// the same domain seed-merge, so a confirmed Composition would survive a
	// re-seed exactly as on Postgres.
	if err := app.SeedStyleCompositions(context.Background(), wines, varieties, seedpkg.StyleCompositions()); err != nil {
		log.Printf("seed style compositions (memory): %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
