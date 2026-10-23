package models

import "time"

// Review is one cafe_review row (BE migrations 011 + 012, BCF-29). UserID nil
// = a partnership review entered in mission control (no account); otherwise a
// review written by a public-site user, which the admin can only moderate
// (status, pin) and re-score (ratings).
type Review struct {
	ID               int
	CafeID           int
	LocationID       string // cafe's location slug, for links
	CafeName         string
	UserID           *int
	UserEmail        *string // users.email (nil for partnership reviews)
	ReviewerName     string
	Status           string
	IsPinned         bool
	OverallScore     *float64
	WFCScore         *float64
	Content          *string
	VisitedAt        *string // YYYY-MM-DD, always the first of the month
	ExternalURLTitle *string
	ExternalURLLink  *string
	CreatedAt        time.Time
	UpdatedAt        time.Time // the author's last edit (public ordering); moderation never bumps it
	Ratings          []Rating
}

// IsPartnership reports whether the review was entered by the admin.
func (r Review) IsPartnership() bool { return r.UserID == nil }

// Rating is one cafe_rating row of a review: a score on one rating dimension.
// Label and Bucket are display-only (rating_type_label.label and the
// rating_category the score falls in) and are never written.
type Rating struct {
	CategoryType                 string
	Score                        float64
	ShortDescriptionOverride     *string
	ShortDescriptionOverrideIndo *string
	Label                        string
	Bucket                       string
}

// Review statuses (review_status_enum).
const (
	ReviewStatusPending  = "pending"
	ReviewStatusApproved = "approved"
	ReviewStatusRejected = "rejected"
	ReviewStatusDeleted  = "deleted"
)

// ReviewStatuses are the /reviews tabs, in display order.
var ReviewStatuses = []string{ReviewStatusPending, ReviewStatusApproved, ReviewStatusRejected, ReviewStatusDeleted}

// ReviewListParams is the /reviews query: status tab, optional cafe filter,
// partnership-only toggle and the after= cursor (last review id shown, 0 =
// from the start).
type ReviewListParams struct {
	Status          string
	CafeID          int
	PartnershipOnly bool
	AfterID         int
}

// ReviewFormOptions is the reference data the review forms need.
type ReviewFormOptions struct {
	Cafes            []CafeListItem
	RatingTypes      []RatingType
	RatingCategories []RatingCategory
}
