package review

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

type GetReviewEditHandler struct{ *handlers.Handler }

func NewGetReviewEditHandler(h *handlers.Handler) *GetReviewEditHandler {
	return &GetReviewEditHandler{h}
}

func (h *GetReviewEditHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rv, ok := loadPartnership(h.Handler, w, r)
	if !ok {
		return
	}
	opts, err := formOptions(h.Handler, r)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, http.StatusOK, templates.PartnershipFormPage(*rv, opts, ""))
}
