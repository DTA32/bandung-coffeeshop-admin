package review

import (
	"net/http"
	"strconv"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type GetReviewsHandler struct{ *handlers.Handler }

func NewGetReviewsHandler(h *handlers.Handler) *GetReviewsHandler { return &GetReviewsHandler{h} }

// ServeHTTP renders the queue, or only the next cards for the infinite-scroll
// sentinel. Paging is by the after= cursor (last review id shown), like /posts;
// page still increments so fragment detection works.
func (h *GetReviewsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lp := utils.ParseListParams(r, "")
	q := r.URL.Query()
	p := controllers.NormalizeReviewListParams(models.ReviewListParams{
		Status:          q.Get("status"),
		CafeID:          utils.QueryInt(r, "cafe"),
		PartnershipOnly: q.Get("partnership") == "1",
		AfterID:         utils.QueryInt(r, "after"),
	})
	pg, err := h.Ctl.Review.List(r.Context(), p)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	f := templates.ReviewFilter{Status: p.Status, CafeID: p.CafeID, PartnershipOnly: p.PartnershipOnly}
	params := templates.ReviewFilterValues(f)
	if n := len(pg.Items); n > 0 {
		params.Set("after", strconv.Itoa(pg.Items[n-1].ID))
	}
	next := lp.Next("/reviews", params, pg.HasMore)
	if lp.Fragment {
		h.Render(w, r, http.StatusOK, templates.ReviewCards(pg.Items, p.Status, next))
		return
	}
	cafes, err := h.Ctl.Cafe.Options(r.Context())
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, http.StatusOK, templates.ReviewsPage(pg.Items, f, cafes, next, q.Get("ok"), q.Get("err")))
}
