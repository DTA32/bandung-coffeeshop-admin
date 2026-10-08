// Package location holds the location route handlers.
package location

import (
	"net/http"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

func saveLocation(h *handlers.Handler, w http.ResponseWriter, r *http.Request, isNew bool) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	l := ParseLocationForm(r.PostForm)
	var err error
	if isNew {
		err = h.Ctl.Location.Create(r.Context(), &l)
	} else {
		err = h.Ctl.Location.Update(r.Context(), r.PathValue("id"), &l)
	}
	if err != nil {
		if controllers.IsNotFound(err) {
			h.NotFound(w, r, "Location")
			return
		}
		msg := handlers.UserError(err)
		if msg == "" {
			// Invalid GeoJSON is rejected by PostGIS; show it rather than a 500.
			if strings.Contains(err.Error(), "GeoJSON") || strings.Contains(err.Error(), "geometry") {
				msg = err.Error()
			} else {
				h.ServerError(w, r, err)
				return
			}
		}
		if !isNew {
			l.ID = r.PathValue("id")
		}
		h.Render(w, r, http.StatusUnprocessableEntity, templates.LocationFormPage(&l, isNew, msg, ""))
		return
	}
	utils.RedirectOK(w, r, "/locations", "Saved "+l.Name+".")
}
