package routes

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/image"
)

// ImageRoutes registers the R2 image upload used by the cafe / location forms.
func ImageRoutes(mux *http.ServeMux, h *handlers.Handler) {
	mux.Handle("POST /images", image.NewPostImageUploadHandler(h))
}
