package post

import (
	"net/http"
	"net/url"
	"time"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

// PostPostBatchHandler handles POST /posts/batch: action=publish|ignore plus
// repeated id= values. Each id goes through the single-post controller path
// (publish = the "Create feed" form's defaults); failures do not abort the
// rest. The HTMX response removes the processed cards and fills the
// #batch-flash summary (out-of-band swaps); a plain form post redirects back.
type PostPostBatchHandler struct{ *handlers.Handler }

func NewPostPostBatchHandler(h *handlers.Handler) *PostPostBatchHandler {
	return &PostPostBatchHandler{h}
}

func (h *PostPostBatchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if r.ParseForm() != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	action, ids, err := controllers.ParsePostBatch(r.PostForm.Get("action"), r.PostForm["id"])
	if err != nil {
		h.reply(w, r, models.BatchResult{}, err.Error(), true)
		return
	}
	res := h.Ctl.Post.ApplyBatch(r.Context(), action, ids, controllers.FeedDefaults{
		ImageBase: h.Cfg.ImageBaseURL, TZ: h.Cfg.DisplayTZ, Now: time.Now(),
	})
	h.reply(w, r, res, controllers.BatchSummary(res), len(res.Failed) > 0)
}

func (h *PostPostBatchHandler) reply(w http.ResponseWriter, r *http.Request, res models.BatchResult, msg string, isErr bool) {
	if !utils.IsHTMX(r) {
		key := "ok"
		if isErr {
			key = "err"
		}
		http.Redirect(w, r, "/posts?"+key+"="+url.QueryEscape(msg), http.StatusSeeOther)
		return
	}
	h.Render(w, r, http.StatusOK, templates.BatchResult(res, "post", msg, isErr))
}
