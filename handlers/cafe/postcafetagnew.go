package cafe

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

type PostCafeTagNewHandler struct{ *handlers.Handler }

func NewPostCafeTagNewHandler(h *handlers.Handler) *PostCafeTagNewHandler {
	return &PostCafeTagNewHandler{h}
}

// createCafeTag is the cafe form's inline "Add tag": it creates the tag and
// returns its row (assigned + visible) to append to the form's tag list.
func (h *PostCafeTagNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	t := models.Tag{Name: r.PostForm.Get("new_tag_name"), Slug: handlers.OptString(r.PostForm, "new_tag_slug")}
	err := handlers.FriendlyUnique(h.Ctl.Tag.Create(r.Context(), &t), "slug")
	if err != nil {
		msg := handlers.UserError(err)
		if msg == "" {
			h.ServerError(w, r, err)
			return
		}
		w.Header().Set("HX-Retarget", "#new-tag-error")
		w.Header().Set("HX-Reswap", "innerHTML")
		h.Render(w, r, http.StatusOK, templates.NewTagError(msg))
		return
	}
	h.Render(w, r, http.StatusOK, templates.CafeTagRow(t, true, true, "", nil, false))
}
