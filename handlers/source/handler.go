// Package source holds the source route handlers.
package source

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

func renderSources(h *handlers.Handler, w http.ResponseWriter, r *http.Request, status int, ok, errMsg string) {
	q := r.URL.Query()
	f := templates.SourcesFilter{Source: q.Get("source"), Failing: q.Get("failing") != "", CafeID: utils.QueryInt(r, "cafe")}
	targets, err := h.Ctl.Source.ListTargets(r.Context(), models.TargetFilter{Source: f.Source, FailingOnly: f.Failing, CafeID: f.CafeID})
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	runs, err := h.Ctl.Source.RecentRuns(r.Context(), 12)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	cafes, err := h.Ctl.Cafe.List(r.Context(), "")
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, status, templates.SourcesPage(targets, runs, cafes, f, ok, errMsg))
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
