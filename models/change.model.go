package models

import (
	"encoding/json"
	"time"
)

// Proposal fields written by bandung-coffeeshop-scraper. See
// bandung-coffeeshop-be/docs/freshness-pipeline-plan.md "Proposal fields".
const (
	FieldHours       = "hours"
	FieldPhone       = "phone"
	FieldAddress     = "address"
	FieldStatus      = "status"
	FieldTempClosed  = "temp_closed"
	FieldMenu        = "menu"
	FieldMenuContent = "menu_content"
	FieldIGBio       = "ig_bio"
	FieldIGPost      = "ig_post"
)

// IsInformational reports whether approving the field has nothing to write:
// the admin can only dismiss it or jump to the cafe form.
func IsInformational(field string) bool {
	switch field {
	case FieldTempClosed, FieldMenuContent, FieldIGBio:
		return true
	}
	return false
}

const (
	StatusPending    = "pending"
	StatusApproved   = "approved"
	StatusRejected   = "rejected"
	StatusSuperseded = "superseded"
)

// Proposal is a row of change_proposal. The same table backs field changes
// (this entity) and Instagram posts (see post.model.go).
type Proposal struct {
	ID          int64
	CafeID      int
	LocationID  string
	CafeName    string
	Source      string
	Field       string
	ItemKey     string
	OldValue    json.RawMessage
	NewValue    json.RawMessage
	EvidenceURL *string
	SnapshotID  *int64
	Status      string
	CreatedAt   time.Time
	// Current is the live value of the field right now (nil for
	// informational fields and posts).
	Current json.RawMessage
}

// CafeProposals groups pending field proposals of one cafe.
type CafeProposals struct {
	CafeID     int
	LocationID string
	CafeName   string
	Items      []Proposal
}

// FieldUpdate is a validated value to write for an applied field proposal.
type FieldUpdate struct {
	Text  *string        // phone, menu, address, status
	Hours *HoursProposal // hours: window + rows (new_value shape in cafe_hour.model.go)
}

// BatchFailure is one item of a batch that could not be processed.
type BatchFailure struct {
	ID     int64
	Reason string
}

// BatchResult is the outcome of a batch approve / reject. Done items were
// resolved (their rows leave the queue); Skipped items are informational
// proposals left untouched by an approve; Failed items stay pending.
type BatchResult struct {
	Action  string
	Done    []int64
	Skipped []int64
	Failed  []BatchFailure
}
