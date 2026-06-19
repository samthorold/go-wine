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
	mux                 *http.ServeMux
	logTasting          *app.LogTastingHandler
	listTastings        *app.ListTastingsHandler
	listVarieties       *app.ListVarietiesHandler
	getVariety          *app.GetVarietyHandler
	editCharacteristics *app.EditCharacteristicsHandler
	listWines           *app.ListWinesHandler
	getWine             *app.GetWineHandler
	editComposition     *app.EditCompositionHandler
	styleComposition    *app.ResolveStyleCompositionHandler
	createDrinker       *app.CreateDrinkerHandler
	renameDrinker       *app.RenameDrinkerHandler
	drinkers            domain.DrinkerRepository
	wines               domain.WineRepository
	varieties           domain.VarietyRepository
	companions          domain.CompanionRepository
}

func NewServer(d domain.DrinkerRepository, w domain.WineRepository, v domain.VarietyRepository, c domain.CompanionRepository, logH *app.LogTastingHandler, listH *app.ListTastingsHandler, listV *app.ListVarietiesHandler, getV *app.GetVarietyHandler, editVC *app.EditCharacteristicsHandler, listW *app.ListWinesHandler, getW *app.GetWineHandler, editC *app.EditCompositionHandler, styleC *app.ResolveStyleCompositionHandler, createD *app.CreateDrinkerHandler, renameD *app.RenameDrinkerHandler) *Server {
	s := &Server{
		mux:                 http.NewServeMux(),
		logTasting:          logH,
		listTastings:        listH,
		listVarieties:       listV,
		getVariety:          getV,
		editCharacteristics: editVC,
		listWines:           listW,
		getWine:             getW,
		editComposition:     editC,
		styleComposition:    styleC,
		createDrinker:       createD,
		renameDrinker:       renameD,
		drinkers:            d,
		wines:               w,
		varieties:           v,
		companions:          c,
	}
	s.mux.HandleFunc("GET /{$}", s.handleRoot)
	s.mux.HandleFunc("GET /tastings", s.handleTastings)
	s.mux.HandleFunc("POST /tastings", s.handleLogTasting)
	s.mux.HandleFunc("GET /varieties", s.handleVarieties)
	s.mux.HandleFunc("GET /varieties/{id}", s.handleVariety)
	s.mux.HandleFunc("GET /varieties/{id}/edit", s.handleVarietyEdit)
	s.mux.HandleFunc("PUT /varieties/{id}", s.handleEditCharacteristics)
	s.mux.HandleFunc("GET /wines", s.handleWines)
	s.mux.HandleFunc("GET /wines/{id}", s.handleWine)
	s.mux.HandleFunc("GET /wines/{id}/style-default", s.handleWineStyleDefault)
	s.mux.HandleFunc("PUT /wines/{id}", s.handleEditComposition)
	s.mux.HandleFunc("POST /switch", s.handleSwitch)
	s.mux.HandleFunc("POST /drinkers", s.handleCreateDrinker)
	s.mux.HandleFunc("PUT /drinkers/{id}", s.handleRenameDrinker)
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

// handleVariety renders a single Variety's detail page: its name, intrinsic
// Characteristics, and the edit form. An unknown Variety is a 404.
func (s *Server) handleVariety(w http.ResponseWriter, r *http.Request) {
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
	id := domain.ID(r.PathValue("id"))
	variety, err := s.getVariety.Handle(ctx, id)
	if errors.Is(err, domain.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = views.VarietyDetailPage(dopts, variety, characteristicsForm(variety)).Render(ctx, w)
}

// handleVarietyEdit returns the bare edit-characteristics form fragment, prefilled
// from the stored values. An unknown Variety is a 404.
func (s *Server) handleVarietyEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := domain.ID(r.PathValue("id"))
	variety, err := s.getVariety.Handle(ctx, id)
	if errors.Is(err, domain.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = views.VarietyCharacteristicsForm(characteristicsForm(variety)).Render(ctx, w)
}

// handleEditCharacteristics persists a hand-edit of a Variety's Characteristics,
// marking them confirmed so a future re-seed cannot clobber them. Success swaps a
// fresh form into #characteristics-form plus the updated #characteristics view
// out-of-band; a validation failure re-renders the form (422) preserving the
// entered values; an unknown Variety is a 404.
func (s *Server) handleEditCharacteristics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := domain.ID(r.PathValue("id"))

	cmd := app.EditCharacteristicsCommand{
		VarietyID: id,
		Body:      atoiOrZero(r.FormValue("body")),
		Tannin:    atoiOrZero(r.FormValue("tannin")),
		Acidity:   atoiOrZero(r.FormValue("acidity")),
		Sweetness: atoiOrZero(r.FormValue("sweetness")),
		Alcohol:   atoiOrZero(r.FormValue("alcohol")),
		Notes:     parseNotes(r.FormValue("notes")),
	}
	err := s.editCharacteristics.Handle(ctx, cmd)
	if err == nil {
		variety, _ := s.getVariety.Handle(ctx, id)
		_ = views.VarietyCharacteristicsForm(characteristicsForm(variety)).Render(ctx, w)
		_ = views.CharacteristicsViewOOB(variety).Render(ctx, w)
		return
	}
	if errors.Is(err, domain.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if errors.Is(err, domain.ErrInvalidCharacteristics) {
		model := characteristicsFormFromRequest(id, r)
		model.Errors = map[string]string{"": "each characteristic must be a whole number between 1 and 5"}
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = views.VarietyCharacteristicsForm(model).Render(ctx, w)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// characteristicsForm builds the edit form prefilled from a Variety's stored
// Characteristics. Empty for an unseeded Variety.
func characteristicsForm(v app.VarietyDetailView) views.VarietyCharacteristicsFormModel {
	m := views.VarietyCharacteristicsFormModel{VarietyID: v.ID.String(), VarietyName: v.Name}
	if v.HasCharacteristics {
		m.Body = strconv.Itoa(v.Body)
		m.Tannin = strconv.Itoa(v.Tannin)
		m.Acidity = strconv.Itoa(v.Acidity)
		m.Sweetness = strconv.Itoa(v.Sweetness)
		m.Alcohol = strconv.Itoa(v.Alcohol)
		m.Notes = strings.Join(v.Notes, ", ")
	}
	return m
}

// characteristicsFormFromRequest rebuilds the form from a failed submit,
// preserving exactly what the Drinker typed.
func characteristicsFormFromRequest(id domain.ID, r *http.Request) views.VarietyCharacteristicsFormModel {
	return views.VarietyCharacteristicsFormModel{
		VarietyID: id.String(),
		Body:      r.FormValue("body"),
		Tannin:    r.FormValue("tannin"),
		Acidity:   r.FormValue("acidity"),
		Sweetness: r.FormValue("sweetness"),
		Alcohol:   r.FormValue("alcohol"),
		Notes:     r.FormValue("notes"),
	}
}

// atoiOrZero parses a form integer, yielding 0 (which the domain rejects) for a
// blank or non-numeric value.
func atoiOrZero(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// parseNotes splits the comma-separated flavour-note field into trimmed tags,
// dropping blanks.
func parseNotes(s string) []string {
	fields := strings.Split(s, ",")
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if n := strings.TrimSpace(f); n != "" {
			out = append(out, n)
		}
	}
	return out
}

func (s *Server) handleWines(w http.ResponseWriter, r *http.Request) {
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
	wines, err := s.listWines.Handle(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = views.WinesPage(dopts, wines).Render(ctx, w)
}

func (s *Server) handleWine(w http.ResponseWriter, r *http.Request) {
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
	id := domain.ID(r.PathValue("id"))
	wine, err := s.getWine.Handle(ctx, id)
	if errors.Is(err, domain.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	vopts, err := s.varietyOptions(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = views.WineDetailPage(dopts, wine, s.compositionForm(wine, vopts)).Render(ctx, w)
}

// handleEditComposition sets a Wine's Composition. Success swaps a fresh empty
// form into #composition-form plus the updated #composition view out-of-band; a
// validation failure re-renders the form (422) with the entered rows preserved
// and an inline error. An unknown Wine is a 404.
func (s *Server) handleEditComposition(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := domain.ID(r.PathValue("id"))

	parts := parseCompositionParts(r)
	err := s.editComposition.Handle(ctx, app.EditCompositionCommand{WineID: id, Parts: parts})
	if err == nil {
		// Success: re-render a fresh form (primary target) and update the
		// #composition view out-of-band.
		wine, _ := s.getWine.Handle(ctx, id)
		vopts, _ := s.varietyOptions(ctx)
		_ = views.CompositionForm(s.compositionForm(wine, vopts)).Render(ctx, w)
		_ = views.CompositionViewOOB(wine).Render(ctx, w)
		return
	}

	// An unknown Wine (the resource itself is missing) is a 404; everything
	// else — an empty/sum-off Composition, or an unknown Variety picked in the
	// form — is a validation failure re-rendered against the form (422).
	if errors.Is(err, domain.ErrNotFound) && !wineExists(ctx, s, id) {
		http.NotFound(w, r)
		return
	}
	if errors.Is(err, domain.ErrInvalidComposition) || errors.Is(err, domain.ErrNotFound) {
		vopts, _ := s.varietyOptions(ctx)
		model := s.compositionFormFromRequest(id, r, vopts)
		model.Errors = map[string]string{"": compositionError(err)}
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = views.CompositionForm(model).Render(ctx, w)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// handleWineStyleDefault returns the edit-Composition form prefilled with the
// Wine's Style → default Composition — the conventional grapes a label's Style
// implies — so a Drinker who doesn't know the grapes can fill them from the
// Style and then override before saving. It is a fragment GET (a sub-resource of
// the Wine, named for what fills the target): it returns the bare CompositionForm
// the caller swaps into #composition-form. Saving the form (PUT) confirms the
// grapes. A Style with no conventional default falls back to the Wine's current
// rows with a gentle note; an unknown Wine is a 404.
func (s *Server) handleWineStyleDefault(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := domain.ID(r.PathValue("id"))
	wine, err := s.getWine.Handle(ctx, id)
	if errors.Is(err, domain.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	vopts, err := s.varietyOptions(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c, err := s.styleComposition.Handle(ctx, wine.Style)
	if errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrInvalidComposition) {
		// No conventional default for this Style (or none its grapes can form):
		// re-render the form as-is with a gentle banner, rather than erroring.
		model := s.compositionForm(wine, vopts)
		model.Errors = map[string]string{"": "no conventional grapes are known for this style — please enter them"}
		_ = views.CompositionForm(model).Render(ctx, w)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Prefill the form rows from the Style default; the Drinker can override
	// before saving (which confirms the grapes).
	rows := make([]views.CompositionRow, 0, len(c.Parts)+1)
	for _, p := range c.Parts {
		rows = append(rows, views.CompositionRow{VarietyID: p.VarietyID.String(), Proportion: strconv.Itoa(p.Proportion)})
	}
	rows = append(rows, views.CompositionRow{})
	model := views.CompositionFormModel{
		WineID:    wine.ID.String(),
		WineLabel: wine.Label,
		Style:     wine.Style,
		Varieties: vopts,
		Rows:      rows,
	}
	_ = views.CompositionForm(model).Render(ctx, w)
}

// wineExists reports whether the Wine itself exists, used to distinguish an
// unknown-Variety ErrNotFound (a 422 against the form) from an unknown-Wine
// ErrNotFound (a 404).
func wineExists(ctx context.Context, s *Server, id domain.ID) bool {
	_, err := s.wines.Get(ctx, id)
	return err == nil
}

// parseCompositionParts pairs up the parallel variety_id / proportion form
// fields into command inputs, dropping rows where no Variety was chosen. Blank
// or non-numeric proportions become 0, which the domain rejects.
func parseCompositionParts(r *http.Request) []app.CompositionPartInput {
	vids := r.Form["variety_id"]
	props := r.Form["proportion"]
	var out []app.CompositionPartInput
	for i, vid := range vids {
		if vid == "" {
			continue
		}
		prop := 0
		if i < len(props) {
			prop, _ = strconv.Atoi(props[i])
		}
		out = append(out, app.CompositionPartInput{VarietyID: domain.ID(vid), Proportion: prop})
	}
	return out
}

// compositionForm builds the edit form for a Wine, prefilling one row per
// existing Composition part plus a couple of blank rows to add more grapes.
func (s *Server) compositionForm(wine app.WineDetailView, vopts []views.VarietyOption) views.CompositionFormModel {
	rows := make([]views.CompositionRow, 0, len(wine.Composition)+2)
	for _, p := range wine.Composition {
		rows = append(rows, views.CompositionRow{VarietyID: p.VarietyID.String(), Proportion: strconv.Itoa(p.Proportion)})
	}
	rows = append(rows, views.CompositionRow{}, views.CompositionRow{})
	return views.CompositionFormModel{
		WineID:    wine.ID.String(),
		WineLabel: wine.Label,
		Style:     wine.Style,
		Varieties: vopts,
		Rows:      rows,
	}
}

// compositionFormFromRequest rebuilds the form from a failed submit, preserving
// the rows the Drinker entered.
func (s *Server) compositionFormFromRequest(wineID domain.ID, r *http.Request, vopts []views.VarietyOption) views.CompositionFormModel {
	vids := r.Form["variety_id"]
	props := r.Form["proportion"]
	rows := make([]views.CompositionRow, 0, len(vids)+1)
	for i, vid := range vids {
		prop := ""
		if i < len(props) {
			prop = props[i]
		}
		rows = append(rows, views.CompositionRow{VarietyID: vid, Proportion: prop})
	}
	rows = append(rows, views.CompositionRow{})
	style := ""
	if wine, err := s.wines.Get(r.Context(), wineID); err == nil {
		style = wine.Style
	}
	return views.CompositionFormModel{
		WineID:    wineID.String(),
		Style:     style,
		Varieties: vopts,
		Rows:      rows,
	}
}

// compositionError maps an edit failure to a form-level message.
func compositionError(err error) string {
	if errors.Is(err, domain.ErrNotFound) {
		return "please choose known grapes"
	}
	return "grapes must add up to about 100%, with at least one grape"
}

func (s *Server) varietyOptions(ctx context.Context) ([]views.VarietyOption, error) {
	vs, err := s.varieties.List(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(vs, func(i, j int) bool { return vs[i].Name < vs[j].Name })
	opts := make([]views.VarietyOption, 0, len(vs))
	for _, v := range vs {
		opts = append(opts, views.VarietyOption{ID: v.ID.String(), Name: v.Name})
	}
	return opts, nil
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

// handleCreateDrinker adds a Drinker and switches to them. Success is a
// navigational mutation: 303 to /tastings (and HX-Redirect so a boosted/htmx
// submit reloads the whole page, surfacing the new active Drinker everywhere).
// An empty name fails in the domain and re-renders the switcher region (422)
// with the entered value and an inline error.
func (s *Server) handleCreateDrinker(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))

	id, err := s.createDrinker.Handle(ctx, app.CreateDrinkerCommand{Name: name})
	if err != nil {
		active, _ := s.activeDrinker(r)
		s.renderSwitcher422(w, r, active.ID, name, err)
		return
	}

	// The new Drinker becomes the active one via the switcher cookie.
	http.SetCookie(w, &http.Cookie{Name: drinkerCookie, Value: id.String(), Path: "/", HttpOnly: true})
	s.redirectToTastings(w, r)
}

// handleRenameDrinker renames an existing Drinker. Success redirects to
// /tastings so the renamed Drinker's new name shows in the switcher; an empty
// name re-renders the switcher region (422) with an inline error.
func (s *Server) handleRenameDrinker(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := domain.ID(r.PathValue("id"))
	name := strings.TrimSpace(r.FormValue("name"))

	err := s.renameDrinker.Handle(ctx, app.RenameDrinkerCommand{ID: id, Name: name})
	switch {
	case err == nil:
		s.redirectToTastings(w, r)
	case errors.Is(err, domain.ErrValidation):
		s.renderSwitcher422(w, r, id, name, err)
	case errors.Is(err, domain.ErrNotFound):
		http.NotFound(w, r)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// redirectToTastings lands a navigational mutation on the Tastings page: a 303
// for a plain/boosted submit, plus HX-Redirect so an explicit-htmx submit
// (the add/rename forms target #drinker-switcher) navigates the whole page
// rather than swapping the redirect target into that region.
func (s *Server) redirectToTastings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("HX-Redirect", "/tastings")
	http.Redirect(w, r, "/tastings", http.StatusSeeOther)
}

// renderSwitcher422 re-renders the switcher region with the entered name and an
// inline domain error, on a 422 that htmx swaps into #drinker-switcher.
func (s *Server) renderSwitcher422(w http.ResponseWriter, r *http.Request, activeID domain.ID, name string, err error) {
	ctx := r.Context()
	dopts, derr := s.drinkerOptions(ctx, activeID)
	if derr != nil {
		http.Error(w, derr.Error(), http.StatusInternalServerError)
		return
	}
	model := views.DrinkerSwitcherModel{
		Drinkers: dopts,
		Name:     name,
		Errors:   map[string]string{"name": drinkerNameError(err)},
	}
	w.WriteHeader(http.StatusUnprocessableEntity)
	_ = views.DrinkerSwitcher(model).Render(ctx, w)
}

// drinkerNameError maps a create/rename failure to a field message. A blank
// name is the only attributable case; anything else surfaces generically.
func drinkerNameError(err error) string {
	if errors.Is(err, domain.ErrValidation) {
		return "please enter a name"
	}
	return "could not save that drinker, please try again"
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
