package tag

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type PostTagNewHandler struct{ *handlers.Handler }

func NewPostTagNewHandler(h *handlers.Handler) *PostTagNewHandler { return &PostTagNewHandler{h} }

func (h *PostTagNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	t := ParseTagForm(r.PostForm)
	err := handlers.FriendlyUnique(h.Ctl.Tag.Create(r.Context(), &t), "slug")
	if err != nil {
		if msg := handlers.UserError(err); msg != "" {
			renderTags(h.Handler, w, r, http.StatusUnprocessableEntity, msg)
			return
		}
		h.ServerError(w, r, err)
		return
	}
	http.Redirect(w, r, "/tags", http.StatusSeeOther)
}
