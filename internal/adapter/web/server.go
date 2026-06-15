// Package web is the HTTP/HTMX adapter. It renders templ components and drives
// the application's command and query handlers. Like the Postgres adapter, it
// sits on the rim of the onion and the domain knows nothing about it.
package web

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"time"

	"go-wine/internal/adapter/web/views"
	"go-wine/internal/app"
	"go-wine/internal/domain"
)

const drinkerCookie = "drinker"

// Server holds the routes and the handlers/repos they need.
type Server struct {
	mux          *http.ServeMux
	logTasting   *app.LogTastingHandler
	listTastings *app.ListTastingsHandler
	drinkers     domain.DrinkerRepository
	wines        domain.WineRepository
}

func NewServer(d domain.DrinkerRepository, w domain.WineRepository, logH *app.LogTastingHandler, listH *app.ListTastingsHandler) *Server {
	s := &Server{
		mux:          http.NewServeMux(),
		logTasting:   logH,
		listTastings: listH,
		drinkers:     d,
		wines:        w,
	}
	s.mux.HandleFunc("GET /{$}", s.handleRoot)
	s.mux.HandleFunc("GET /tastings", s.handleTastings)
	s.mux.HandleFunc("POST /tastings", s.handleLogTasting)
	s.mux.HandleFunc("POST /switch", s.handleSwitch)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/tastings", http.StatusSeeOther)
}

func (s *Server) handleTastings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	active, err := s.activeDrinker(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dopts, err := s.drinkerOptions(ctx, active.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	wopts, err := s.wineOptions(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tastings, err := s.listTastings.Handle(ctx, active.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = views.TastingsPage(dopts, wopts, tastings).Render(ctx, w)
}

func (s *Server) handleLogTasting(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	active, err := s.activeDrinker(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	wineID := domain.ID(r.FormValue("wine_id"))
	rating, _ := strconv.Atoi(r.FormValue("rating"))
	note := r.FormValue("note")
	var vintage *int
	if v := r.FormValue("vintage"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			vintage = &n
		}
	}

	now := time.Now()
	if _, err := s.logTasting.Handle(ctx, app.LogTastingCommand{
		DrinkerID: active.ID,
		WineID:    wineID,
		Vintage:   vintage,
		Rating:    rating,
		Note:      note,
		DrunkOn:   now,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	label := "(unknown wine)"
	if wine, err := s.wines.Get(ctx, wineID); err == nil {
		label = wine.Label()
	}
	_ = views.TastingRow(app.TastingView{
		WineLabel: label,
		Vintage:   vintage,
		Rating:    rating,
		Note:      note,
		DrunkOn:   now,
	}).Render(ctx, w)
}

func (s *Server) handleSwitch(w http.ResponseWriter, r *http.Request) {
	if id := r.FormValue("drinker"); id != "" {
		http.SetCookie(w, &http.Cookie{Name: drinkerCookie, Value: id, Path: "/", HttpOnly: true})
	}
	http.Redirect(w, r, "/tastings", http.StatusSeeOther)
}

// activeDrinker resolves the current Drinker from the switcher cookie, falling
// back to the first Drinker. There is no authentication; this is selection, not
// sign-in.
func (s *Server) activeDrinker(r *http.Request) (domain.Drinker, error) {
	ctx := r.Context()
	if c, err := r.Cookie(drinkerCookie); err == nil && c.Value != "" {
		if d, err := s.drinkers.Get(ctx, domain.ID(c.Value)); err == nil {
			return d, nil
		}
	}
	ds, err := s.drinkers.List(ctx)
	if err != nil {
		return domain.Drinker{}, err
	}
	if len(ds) == 0 {
		return domain.Drinker{}, domain.ErrNotFound
	}
	sort.Slice(ds, func(i, j int) bool { return ds[i].Name < ds[j].Name })
	return ds[0], nil
}

func (s *Server) drinkerOptions(ctx context.Context, activeID domain.ID) ([]views.DrinkerOption, error) {
	ds, err := s.drinkers.List(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(ds, func(i, j int) bool { return ds[i].Name < ds[j].Name })
	opts := make([]views.DrinkerOption, 0, len(ds))
	for _, d := range ds {
		opts = append(opts, views.DrinkerOption{ID: d.ID.String(), Name: d.Name, Active: d.ID == activeID})
	}
	return opts, nil
}

func (s *Server) wineOptions(ctx context.Context) ([]views.WineOption, error) {
	ws, err := s.wines.List(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(ws, func(i, j int) bool { return ws[i].Label() < ws[j].Label() })
	opts := make([]views.WineOption, 0, len(ws))
	for _, w := range ws {
		opts = append(opts, views.WineOption{ID: w.ID.String(), Label: w.Label()})
	}
	return opts, nil
}
