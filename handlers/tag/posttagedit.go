package tag

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type PostTagEditHandler struct{ *handlers.Handler }

func NewPostTagEditHandler(h *handlers.Handler) *PostTagEditHandler { return &PostTagEditHandler{h} }

func (h *PostTagEditHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, ok := utils.PathInt(r, "id")
	if !ok || r.ParseForm() != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	t := ParseTagForm(r.PostForm)
	t.ID = id
	err := handlers.FriendlyUnique(h.Ctl.Tag.Update(r.Context(), id, &t), "slug")
	if err != nil && handlers.UserError(err) == "" {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, http.StatusOK, templates.TagRow(t, handlers.SavedOr(err)))
}
