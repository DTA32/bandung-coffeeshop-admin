package review

import (
	"net/url"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
)

// ParseRatings reads the ratings editor (templates.RatingsEditor): one
// rating per ticked rating_enabled_<type>, with rating_score_<type> and the
// optional rating_override(_indo)_<type>. It always returns the best-effort
// ratings so the form can be re-rendered with the user's input.
func ParseRatings(v url.Values, types []models.RatingType) ([]models.Rating, error) {
	var (
		out      []models.Rating
		firstErr error
	)
	for _, rt := range types {
		if !handlers.Checked(v, "rating_enabled_"+rt.Type) {
			continue
		}
		score, err := handlers.OptFloat(v, "rating_score_"+rt.Type)
		if err == nil && score == nil {
			err = handlers.FormErrorf("%s is enabled but has no score", rt.Label)
		}
		if err != nil && firstErr == nil {
			firstErr = err
		}
		r := models.Rating{
			CategoryType:                 rt.Type,
			Label:                        rt.Label,
			ShortDescriptionOverride:     handlers.OptString(v, "rating_override_"+rt.Type),
			ShortDescriptionOverrideIndo: handlers.OptString(v, "rating_override_indo_"+rt.Type),
		}
		if score != nil {
			r.Score = *score
		}
		out = append(out, r)
	}
	return out, firstErr
}

// ParsePartnershipForm maps the partnership review form onto a review
// (user_id stays nil). It always returns the best-effort payload so the form
// can be re-rendered with the user's input; normalisation and the remaining
// validation happen in controllers.NormalizePartnershipReview.
func ParsePartnershipForm(v url.Values, types []models.RatingType) (models.Review, error) {
	rv := models.Review{
		ReviewerName:     v.Get("reviewer_name"),
		Content:          handlers.OptString(v, "content"),
		VisitedAt:        handlers.OptString(v, "visited_month"),
		ExternalURLTitle: handlers.OptString(v, "external_url_title"),
		ExternalURLLink:  handlers.OptString(v, "external_url_link"),
		IsPinned:         handlers.Checked(v, "is_pinned"),
	}
	var firstErr error
	keep := func(err error) {
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	cafeID, err := handlers.OptInt(v, "cafe_id")
	keep(err)
	if cafeID != nil {
		rv.CafeID = *cafeID
	}
	rv.OverallScore, err = handlers.OptFloat(v, "overall_score")
	keep(err)
	rv.WFCScore, err = handlers.OptFloat(v, "wfc_score")
	keep(err)
	rv.Ratings, err = ParseRatings(v, types)
	keep(err)
	return rv, firstErr
}
