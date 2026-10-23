package controllers

import (
	"context"
	"math"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/repository"
)

// ReviewController moderates cafe reviews (BCF-31). On a user review the admin
// can only change the status (approve / reject), the pin and the ratings; the
// author's text, scores and updated_at stay as written (the public BE puts an
// edited review back to pending itself). Partnership reviews (user_id NULL)
// are fully edited here and are created approved.
type ReviewController struct {
	repo        *repository.ReviewRepository
	ratingTypes *repository.RatingTypeRepository
}

func NewReviewController(repo *repository.ReviewRepository, ratingTypes *repository.RatingTypeRepository) *ReviewController {
	return &ReviewController{repo: repo, ratingTypes: ratingTypes}
}

// Allowed status moves: target status -> statuses it can be reached from.
var reviewTransitions = map[string][]string{
	models.ReviewStatusApproved: {models.ReviewStatusPending, models.ReviewStatusRejected},
	models.ReviewStatusRejected: {models.ReviewStatusPending, models.ReviewStatusApproved},
	models.ReviewStatusDeleted:  {models.ReviewStatusPending, models.ReviewStatusApproved, models.ReviewStatusRejected},
}

// NormalizeReviewListParams maps an unknown status tab to pending and a
// negative cafe / cursor to none.
func NormalizeReviewListParams(p models.ReviewListParams) models.ReviewListParams {
	if !slices.Contains(models.ReviewStatuses, p.Status) {
		p.Status = models.ReviewStatusPending
	}
	p.CafeID = max(p.CafeID, 0)
	p.AfterID = max(p.AfterID, 0)
	return p
}

func (c *ReviewController) List(ctx context.Context, p models.ReviewListParams) (models.Page[models.Review], error) {
	return c.repo.List(ctx, NormalizeReviewListParams(p))
}

func (c *ReviewController) Get(ctx context.Context, id int) (*models.Review, error) {
	return c.repo.Get(ctx, id)
}

// Approve publishes a pending or rejected review.
func (c *ReviewController) Approve(ctx context.Context, id int) error {
	return c.move(ctx, id, models.ReviewStatusApproved, false)
}

// Reject hides a pending or approved review.
func (c *ReviewController) Reject(ctx context.Context, id int) error {
	return c.move(ctx, id, models.ReviewStatusRejected, false)
}

// Remove soft-deletes a partnership review (user reviews are deleted by their
// author on the public site, never here).
func (c *ReviewController) Remove(ctx context.Context, id int) error {
	return c.move(ctx, id, models.ReviewStatusDeleted, true)
}

func (c *ReviewController) move(ctx context.Context, id int, to string, partnershipOnly bool) error {
	return c.repo.SetStatus(ctx, id, to, reviewTransitions[to], partnershipOnly)
}

// SetPinned pins / unpins any non-deleted review.
func (c *ReviewController) SetPinned(ctx context.Context, id int, pinned bool) error {
	return c.repo.SetPinned(ctx, id, pinned)
}

// SaveRatings validates ratings against each type's bucket bounds and replaces
// the review's ratings (status and updated_at untouched).
func (c *ReviewController) SaveRatings(ctx context.Context, id int, ratings []models.Rating) error {
	if err := c.validateRatings(ctx, ratings); err != nil {
		return err
	}
	return c.repo.ReplaceRatings(ctx, id, ratings)
}

// CreatePartnership validates and inserts an approved partnership review;
// returns its id.
func (c *ReviewController) CreatePartnership(ctx context.Context, rv *models.Review) (int, error) {
	if err := NormalizePartnershipReview(rv); err != nil {
		return 0, err
	}
	if err := c.validateRatings(ctx, rv.Ratings); err != nil {
		return 0, err
	}
	return c.repo.CreatePartnership(ctx, *rv)
}

// UpdatePartnership validates and rewrites a partnership review (bumps
// updated_at, keeps the status).
func (c *ReviewController) UpdatePartnership(ctx context.Context, id int, rv *models.Review) error {
	if err := NormalizePartnershipReview(rv); err != nil {
		return err
	}
	if err := c.validateRatings(ctx, rv.Ratings); err != nil {
		return err
	}
	return c.repo.UpdatePartnership(ctx, id, *rv)
}

func (c *ReviewController) validateRatings(ctx context.Context, ratings []models.Rating) error {
	seen := make(map[string]struct{}, len(ratings))
	for _, rt := range ratings {
		if _, ok := seen[rt.CategoryType]; ok {
			return invalid("duplicate rating type %q", rt.CategoryType)
		}
		seen[rt.CategoryType] = struct{}{}
	}
	if len(ratings) == 0 {
		return nil
	}
	bounds, err := c.ratingTypes.Bounds(ctx)
	if err != nil {
		return err
	}
	return ValidateRatings(ratings, bounds)
}

// NormalizePartnershipReview trims and validates a partnership review in
// place: cafe, reviewer name and content are required; scores are optional,
// 1–5 in 0.1 steps; the visit month ("YYYY-MM" or a date) becomes the first
// of that month; the external link must be http(s) and a title needs a link.
func NormalizePartnershipReview(rv *models.Review) error {
	if rv.CafeID <= 0 {
		return invalid("cafe is required")
	}
	rv.ReviewerName = strings.TrimSpace(rv.ReviewerName)
	if rv.ReviewerName == "" {
		return invalid("reviewer name is required")
	}
	rv.Content = NilIfBlank(rv.Content)
	if rv.Content == nil {
		return invalid("content is required")
	}
	for _, sc := range []struct {
		name string
		v    *float64
	}{{"overall score", rv.OverallScore}, {"WFC score", rv.WFCScore}} {
		if sc.v == nil {
			continue
		}
		if *sc.v < 1 || *sc.v > 5 {
			return invalid("%s must be between 1 and 5", sc.name)
		}
		if math.Abs(*sc.v*10-math.Round(*sc.v*10)) > 1e-9 {
			return invalid("%s takes one decimal at most", sc.name)
		}
	}

	rv.VisitedAt = NilIfBlank(rv.VisitedAt)
	if rv.VisitedAt != nil {
		m, err := parseVisitMonth(*rv.VisitedAt)
		if err != nil {
			return err
		}
		rv.VisitedAt = &m
	}

	rv.ExternalURLTitle = NilIfBlank(rv.ExternalURLTitle)
	rv.ExternalURLLink = NilIfBlank(rv.ExternalURLLink)
	if l := rv.ExternalURLLink; l != nil {
		u, err := url.Parse(*l)
		if err != nil || !isHTTPURL(*l) || u.Host == "" {
			return invalid("external link must be an http(s) URL")
		}
	}
	if rv.ExternalURLTitle != nil && rv.ExternalURLLink == nil {
		return invalid("external link title needs a link")
	}
	return nil
}

// parseVisitMonth turns "YYYY-MM" (or "YYYY-MM-DD") into the first of that
// month, "YYYY-MM-01".
func parseVisitMonth(s string) (string, error) {
	for _, layout := range []string{"2006-01", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format("2006-01") + "-01", nil
		}
	}
	return "", invalid("invalid visit month %q: expected YYYY-MM", s)
}
