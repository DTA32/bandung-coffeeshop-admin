package change

import (
	"net/http"
	"net/url"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

// PostChangeBatchHandler handles POST /changes/batch: action=approve|reject
// plus repeated id= values. Each id goes through the single-row controller
// path; failures do not abort the rest. The HTMX response removes the
// processed rows and fills the #batch-flash summary (out-of-band swaps); a
// plain form post redirects back with the summary.
type PostChangeBatchHandler struct{ *handlers.Handler }

func NewPostChangeBatchHandler(h *handlers.Handler) *PostChangeBatchHandler {
	return &PostChangeBatchHandler{h}
}

func (h *PostChangeBatchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if r.ParseForm() != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	action, ids, err := controllers.ParseBatch(r.PostForm.Get("action"), r.PostForm["id"])
	if err != nil {
		h.reply(w, r, models.BatchResult{}, err.Error(), true)
		return
	}
	res := h.Ctl.Change.ApplyBatch(r.Context(), action, ids)
	h.reply(w, r, res, controllers.BatchSummary(res), len(res.Failed) > 0)
}

func (h *PostChangeBatchHandler) reply(w http.ResponseWriter, r *http.Request, res models.BatchResult, msg string, isErr bool) {
	if !utils.IsHTMX(r) {
		key := "ok"
		if isErr {
			key = "err"
		}
		http.Redirect(w, r, "/changes?"+key+"="+url.QueryEscape(msg), http.StatusSeeOther)
		return
	}
	h.Render(w, r, http.StatusOK, templates.BatchResult(res, msg, isErr))
}
