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
	"strings"
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
	companions    domain.CompanionRepository
}

func NewServer(d domain.DrinkerRepository, w domain.WineRepository, c domain.CompanionRepository, logH *app.LogTastingHandler, listH *app.ListTastingsHandler, listV *app.ListVarietiesHandler) *Server {
	s := &Server{
		mux:           http.NewServeMux(),
		logTasting:    logH,
		listTastings:  listH,
		listVarieties: listV,
		drinkers:      d,
		wines:         w,
		companions:    c,
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
	copts, err := s.companionOptions(ctx, active.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tastings, err := s.listTastings.Handle(ctx, active.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = views.TastingsPage(dopts, views.LogFormModel{Wines: wopts, Companions: copts}, tastings).Render(ctx, w)
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

	companionIDs, err := s.resolveCompanions(ctx, active.ID, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := time.Now()
	_, err = s.logTasting.Handle(ctx, app.LogTastingCommand{
		DrinkerID:  active.ID,
		WineID:     wineID,
		Vintage:    vintage,
		Rating:     rating,
		Note:       note,
		Companions: companionIDs,
		DrunkOn:    now,
	})
	if err != nil {
		// Validation failure: re-render the form (422, which htmx swaps) with the
		// entered values preserved and the error against the offending field. The
		// tastings list is untouched, so no OOB fragment.
		model := s.logFormModel(ctx, active.ID, r)
		model.Errors = logTastingErrors(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = views.LogForm(model).Render(ctx, w)
		return
	}

	// Success: the primary target (#log-form) swaps to a fresh empty form, and
	// the #tastings list updates out-of-band in the same response.
	wopts, _ := s.wineOptions(ctx)
	copts, _ := s.companionOptions(ctx, active.ID)
	tastings, _ := s.listTastings.Handle(ctx, active.ID)
	_ = views.LogForm(views.LogFormModel{Wines: wopts, Companions: copts}).Render(ctx, w)
	_ = views.TastingListOOB(tastings).Render(ctx, w)
}

// resolveCompanions turns the submitted form into a set of Companion IDs to
// attach: the existing Companions ticked (validated to belong to the active
// Drinker's personal zone) plus any new names typed, each persisted as a fresh
// Companion scoped to the active Drinker.
func (s *Server) resolveCompanions(ctx context.Context, drinkerID domain.ID, r *http.Request) ([]domain.ID, error) {
	owned := make(map[domain.ID]bool)
	existing, err := s.companions.ListByDrinker(ctx, drinkerID)
	if err != nil {
		return nil, err
	}
	for _, c := range existing {
		owned[c.ID] = true
	}

	var ids []domain.ID
	for _, raw := range r.Form["companion_id"] {
		id := domain.ID(raw)
		if owned[id] { // never attach a Companion from another Drinker's zone
			ids = append(ids, id)
		}
	}

	for _, name := range parseNewCompanions(r.FormValue("new_companions")) {
		c, err := domain.NewCompanion(drinkerID, name)
		if err != nil {
			continue // skip blanks rather than fail the whole tasting
		}
		if err := s.companions.Add(ctx, c); err != nil {
			return nil, err
		}
		ids = append(ids, c.ID)
	}
	return ids, nil
}

// parseNewCompanions splits the free-text "add new" field into trimmed names,
// separated by commas or newlines, dropping blanks.
func parseNewCompanions(s string) []string {
	fields := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == '\n' || r == '\r' })
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if name := strings.TrimSpace(f); name != "" {
			out = append(out, name)
		}
	}
	return out
}

// logFormModel rebuilds the form's view model from the submitted request,
// preserving the entered values so a failed submit re-renders what the Drinker
// typed. Errors are filled in by the caller.
func (s *Server) logFormModel(ctx context.Context, drinkerID domain.ID, r *http.Request) views.LogFormModel {
	wopts, _ := s.wineOptions(ctx)
	copts, _ := s.companionOptions(ctx, drinkerID)
	return views.LogFormModel{
		Wines:         wopts,
		Companions:    copts,
		WineID:        r.FormValue("wine_id"),
		Vintage:       r.FormValue("vintage"),
		Rating:        r.FormValue("rating"),
		Note:          r.FormValue("note"),
		CompanionIDs:  r.Form["companion_id"],
		NewCompanions: r.FormValue("new_companions"),
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

func (s *Server) companionOptions(ctx context.Context, drinkerID domain.ID) ([]views.CompanionOption, error) {
	cs, err := s.companions.ListByDrinker(ctx, drinkerID)
	if err != nil {
		return nil, err
	}
	sort.Slice(cs, func(i, j int) bool { return cs[i].Name < cs[j].Name })
	opts := make([]views.CompanionOption, 0, len(cs))
	for _, c := range cs {
		opts = append(opts, views.CompanionOption{ID: c.ID.String(), Name: c.Name})
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
