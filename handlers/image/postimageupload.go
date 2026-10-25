// Package image serves the image upload endpoint used by the cafe and location
// forms' image list (templates.ImagesEditor).
package image

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type PostImageUploadHandler struct{ *handlers.Handler }

func NewPostImageUploadHandler(h *handlers.Handler) *PostImageUploadHandler {
	return &PostImageUploadHandler{h}
}

// ServeHTTP takes a multipart form with folder (cafe | location), slug and
// file, uploads it to R2 and answers {"url": ...}, or {"error": ...} with 422
// for anything the admin can fix. forms.js puts the URL into the image row.
func (h *PostImageUploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, controllers.MaxImageBytes+1<<20)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{
				"error": fmt.Sprintf("images must be at most %d MB", controllers.MaxImageBytes>>20)})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad upload form"})
		return
	}
	defer r.MultipartForm.RemoveAll()
	f, fh, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no file"})
		return
	}
	defer f.Close()
	body, err := io.ReadAll(io.LimitReader(f, controllers.MaxImageBytes+1))
	if err != nil {
		h.ServerError(w, r, err)
		return
	}

	url, err := h.Ctl.Image.Upload(r.Context(), r.FormValue("folder"), r.FormValue("slug"), fh.Filename, body)
	if err != nil {
		if msg := handlers.UserError(err); msg != "" {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": msg})
			return
		}
		h.ServerError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
