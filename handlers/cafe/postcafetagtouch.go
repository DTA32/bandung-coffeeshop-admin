package cafe

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type PostCafeTagTouchHandler struct{ *handlers.Handler }

func NewPostCafeTagTouchHandler(h *handlers.Handler) *PostCafeTagTouchHandler {
	return &PostCafeTagTouchHandler{h}
}

// ServeHTTP bumps cafe_tag.updated_at for one saved tag (cafe.updated_at is not
// touched) and returns the row's touch fragment, which HTMX swaps in place.
func (h *PostCafeTagTouchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cafeID := r.PathValue("id")
	tagID, ok := utils.PathInt(r, "tagId")
	if !ok {
		h.NotFound(w, r, "Tag")
		return
	}
	at, err := h.Ctl.Cafe.TouchTag(r.Context(), cafeID, tagID)
	if controllers.IsNotFound(err) {
		http.Error(w, "tag is not saved for this cafe", http.StatusNotFound)
		return
	}
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	// The touched tag is now the newest, so it is the current remark if visible.
	h.Render(w, r, http.StatusOK, templates.CafeTagTouch(cafeID, tagID, at, true, true))
}
