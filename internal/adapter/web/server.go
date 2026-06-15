// Package web is the HTTP/HTMX adapter. It renders templ components and drives
// the application's command and query handlers. Like the Postgres adapter, it
// sits on the rim of the onion and the domain knows nothing about it.
package web

import (
	"context"
	"errors"
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
	mux           *http.ServeMux
	logTasting    *app.LogTastingHandler
	listTastings  *app.ListTastingsHandler
	listVarieties *app.ListVarietiesHandler
	drinkers      domain.DrinkerRepository
	wines         domain.WineRepository
}

func NewServer(d domain.DrinkerRepository, w domain.WineRepository, logH *app.LogTastingHandler, listH *app.ListTastingsHandler, listV *app.ListVarietiesHandler) *Server {
	s := &Server{
		mux:           http.NewServeMux(),
		logTasting:    logH,
		listTastings:  listH,
		listVarieties: listV,
		drinkers:      d,
		wines:         w,
	}
	s.mux.HandleFunc("GET /{$}", s.handleRoot)
	s.mux.HandleFunc("GET /tastings", s.handleTastings)
	s.mux.HandleFunc("POST /tastings", s.handleLogTasting)
	s.mux.HandleFunc("GET /varieties", s.handleVarieties)
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
	_ = views.TastingsPage(dopts, views.LogFormModel{Wines: wopts}, tastings).Render(ctx, w)
}

func (s *Server) handleVarieties(w http.ResponseWriter, r *http.Request) {
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
	varieties, err := s.listVarieties.Handle(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = views.VarietiesPage(dopts, varieties).Render(ctx, w)
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
	_, err = s.logTasting.Handle(ctx, app.LogTastingCommand{
		DrinkerID: active.ID,
		WineID:    wineID,
		Vintage:   vintage,
		Rating:    rating,
		Note:      note,
		DrunkOn:   now,
	})
	if err != nil {
		// Validation failure: re-render the form (422, which htmx swaps) with the
		// entered values preserved and the error against the offending field. The
		// tastings list is untouched, so no OOB fragment.
		model := s.logFormModel(ctx, r)
		model.Errors = logTastingErrors(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = views.LogForm(model).Render(ctx, w)
		return
	}

	// Success: the primary target (#log-form) swaps to a fresh empty form, and
	// the #tastings list updates out-of-band in the same response.
	wopts, _ := s.wineOptions(ctx)
	tastings, _ := s.listTastings.Handle(ctx, active.ID)
	_ = views.LogForm(views.LogFormModel{Wines: wopts}).Render(ctx, w)
	_ = views.TastingListOOB(tastings).Render(ctx, w)
}

// logFormModel rebuilds the form's view model from the submitted request,
// preserving the entered values so a failed submit re-renders what the Drinker
// typed. Errors are filled in by the caller.
func (s *Server) logFormModel(ctx context.Context, r *http.Request) views.LogFormModel {
	wopts, _ := s.wineOptions(ctx)
	return views.LogFormModel{
		Wines:   wopts,
		WineID:  r.FormValue("wine_id"),
		Vintage: r.FormValue("vintage"),
		Rating:  r.FormValue("rating"),
		Note:    r.FormValue("note"),
	}
}

// logTastingErrors maps a LogTasting failure to a field->message error map. A
// rating outside 1..5 is attributable to the rating field; an unknown Wine to
// the wine_id field; anything else surfaces as a form-level banner.
func logTastingErrors(err error) map[string]string {
	switch {
	case errors.Is(err, domain.ErrInvalidRating):
		return map[string]string{"rating": err.Error()}
	case errors.Is(err, domain.ErrNotFound):
		return map[string]string{"wine_id": "please choose a known wine"}
	default:
		return map[string]string{"": "could not log that tasting, please try again"}
	}
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
