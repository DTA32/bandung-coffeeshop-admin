// Package change holds the change route handlers.
package change

import (
	"errors"
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

func renderProposalResult(h *handlers.Handler, w http.ResponseWriter, r *http.Request, p *models.Proposal, err error, status string) {
	if p == nil && err != nil {
		if controllers.IsNotFound(err) {
			http.Error(w, "proposal not found", http.StatusNotFound)
			return
		}
		h.ServerError(w, r, err)
		return
	}
	if err != nil {
		if errors.Is(err, controllers.ErrProposalNotPending) {
			h.Render(w, r, http.StatusOK, templates.ProposalResolved(*p, "already resolved", "refresh to see the current queue"))
			return
		}
		if msg := handlers.UserError(err); msg != "" {
			h.Render(w, r, http.StatusOK, templates.ProposalRow(*p, msg))
			return
		}
		h.ServerError(w, r, err)
		return
	}
	detail := ""
	if status == models.StatusRejected && models.IsInformational(p.Field) {
		status = "dismissed"
	}
	h.Render(w, r, http.StatusOK, templates.ProposalResolved(*p, status, detail))
}
