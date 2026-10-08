// Package cafe holds the cafe route handlers.
package cafe

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

func cafeOptions(h *handlers.Handler, r *http.Request) (models.CafeFormOptions, error) {
	var opts models.CafeFormOptions
	var err error
	if opts.Tags, err = h.Ctl.Tag.List(r.Context()); err != nil {
		return opts, err
	}
	if opts.RatingTypes, err = h.Ctl.RatingType.List(r.Context()); err != nil {
		return opts, err
	}
	opts.RatingCategories, err = h.Ctl.RatingCategory.List(r.Context())
	return opts, err
}

func saveCafe(h *handlers.Handler, w http.ResponseWriter, r *http.Request, isNew bool) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	opts, err := cafeOptions(h, r)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	c, err := ParseCafeForm(r.PostForm, opts.RatingTypes)
	if err == nil {
		if isNew {
			err = h.Ctl.Cafe.Create(r.Context(), &c)
		} else {
			id := r.PathValue("id")
			if existing, gerr := h.Ctl.Cafe.Get(r.Context(), id); gerr == nil {
				c.CafeID = existing.CafeID
			}
			err = h.Ctl.Cafe.Update(r.Context(), id, &c)
		}
	}
	if err != nil {
		if controllers.IsNotFound(err) {
			h.NotFound(w, r, "Cafe")
			return
		}
		msg := handlers.UserError(err)
		if msg == "" {
			h.ServerError(w, r, err)
			return
		}
		if !isNew {
			c.ID = r.PathValue("id")
		}
		h.Render(w, r, http.StatusUnprocessableEntity, templates.CafeFormPage(&c, opts, isNew, msg, ""))
		return
	}
	utils.RedirectOK(w, r, "/cafes", "Saved "+c.Name+".")
}

// FromPath resolves the {id} location path value to the cafe id and name, writing
// the 404/500 response itself and returning false when it cannot.
func FromPath(h *handlers.Handler, w http.ResponseWriter, r *http.Request) (int, string, bool) {
	cafeID, name, err := h.Ctl.Cafe.IDByLocation(r.Context(), r.PathValue("id"))
	if controllers.IsNotFound(err) {
		h.NotFound(w, r, "Cafe")
		return 0, "", false
	}
	if err != nil {
		h.ServerError(w, r, err)
		return 0, "", false
	}
	return cafeID, name, true
}
