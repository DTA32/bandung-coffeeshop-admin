// Package source holds the source route handlers.
package source

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

func renderSources(h *handlers.Handler, w http.ResponseWriter, r *http.Request, status int, ok, errMsg string) {
	q := r.URL.Query()
	f := templates.SourcesFilter{Source: q.Get("source"), Failing: q.Get("failing") != "", CafeID: utils.QueryInt(r, "cafe")}
	lp := utils.ParseListParams(r, "")
	pg, err := h.Ctl.Source.ListTargetsPage(r.Context(), models.TargetFilter{Source: f.Source, FailingOnly: f.Failing, CafeID: f.CafeID}, lp.Page)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	next := lp.Next("/sources", sourcesParams(f), pg.HasMore)
	if lp.Fragment {
		h.Render(w, r, http.StatusOK, templates.TargetRows(pg.Items, next, false))
		return
	}
	runs, err := h.Ctl.Source.RecentRuns(r.Context(), 12)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	cafes, err := h.Ctl.Cafe.Options(r.Context())
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, status, templates.SourcesPage(pg.Items, next, runs, cafes, f, ok, errMsg))
}

// sourcesParams are the filters carried through next-page requests.
func sourcesParams(f templates.SourcesFilter) url.Values {
	v := url.Values{"source": {f.Source}}
	if f.Failing {
		v.Set("failing", "1")
	}
	if f.CafeID != 0 {
		v.Set("cafe", strconv.Itoa(f.CafeID))
	}
	return v
}

func targetFromPath(h *handlers.Handler, w http.ResponseWriter, r *http.Request) (int, string, bool) {
	cafeID, ok := utils.PathInt(r, "cafeID")
	source := r.PathValue("source")
	if !ok || !utils.Contains(models.Sources, source) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return 0, "", false
	}
	return cafeID, source, true
}

func renderTarget(h *handlers.Handler, w http.ResponseWriter, r *http.Request, cafeID int, source, errMsg string) {
	targets, err := h.Ctl.Source.ListTargets(r.Context(), models.TargetFilter{Source: source, CafeID: cafeID})
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	if len(targets) == 0 {
		http.Error(w, "target not found", http.StatusNotFound)
		return
	}
	h.Render(w, r, http.StatusOK, templates.TargetRow(targets[0], errMsg))
}
