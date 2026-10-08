package ratingtype

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type GetRatingTypesHandler struct{ *handlers.Handler }

func NewGetRatingTypesHandler(h *handlers.Handler) *GetRatingTypesHandler {
	return &GetRatingTypesHandler{h}
}

func (h *GetRatingTypesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	renderRatingTypes(h.Handler, w, r, http.StatusOK, "")
}
