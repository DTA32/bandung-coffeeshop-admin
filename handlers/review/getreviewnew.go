package review

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type GetReviewNewHandler struct{ *handlers.Handler }

func NewGetReviewNewHandler(h *handlers.Handler) *GetReviewNewHandler { return &GetReviewNewHandler{h} }

// ServeHTTP renders an empty partnership review form (?cafe= preselects one).
func (h *GetReviewNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	opts, err := formOptions(h.Handler, r)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	rv := models.Review{CafeID: utils.QueryInt(r, "cafe")}
	h.Render(w, r, http.StatusOK, templates.PartnershipFormPage(rv, opts, ""))
}
