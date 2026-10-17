package controllers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

// Batch actions and limits for POST /changes/batch.
const (
	BatchApprove = "approve"
	BatchReject  = "reject"
	MaxBatch     = 200
)

// ParseBatch validates the action and the repeated id values: the action must
// be approve or reject, ids must be positive integers (duplicates are dropped,
// order is kept), at least one and at most MaxBatch.
func ParseBatch(action string, rawIDs []string) (string, []int64, error) {
	if action != BatchApprove && action != BatchReject {
		return "", nil, invalid("unknown batch action %q", action)
	}
	seen := make(map[int64]bool, len(rawIDs))
	ids := make([]int64, 0, len(rawIDs))
	for _, s := range rawIDs {
		id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
		if err != nil || id <= 0 {
			return "", nil, invalid("invalid change id %q", s)
		}
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return "", nil, invalid("select at least one change")
	}
	if len(ids) > MaxBatch {
		return "", nil, invalid("select at most %d changes at a time (got %d)", MaxBatch, len(ids))
	}
	return action, ids, nil
}

// batchSteps are the single-row operations a batch runs; the real ones are
// the controller's own Get / Approve / Reject (same validation, one
// transaction per item, same cafe.updated_at bump).
type batchSteps struct {
	get     func(ctx context.Context, id int64) (*models.Proposal, error)
	approve func(ctx context.Context, id int64, edit *Edit) (*models.Proposal, error)
	reject  func(ctx context.Context, id int64) error
}

// ApplyBatch approves (as proposed) or rejects each id through the existing
// single-row path. Informational proposals are skipped on approve (they can
// only be dismissed); on reject they are dismissed like the single-row button.
// One failing item never aborts the rest. The caller validates with ParseBatch.
func (c *ChangeController) ApplyBatch(ctx context.Context, action string, ids []int64) models.BatchResult {
	return runBatch(ctx, action, ids, batchSteps{get: c.Get, approve: c.Approve, reject: c.Reject})
}

func runBatch(ctx context.Context, action string, ids []int64, s batchSteps) models.BatchResult {
	res := models.BatchResult{Action: action}
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			res.Failed = append(res.Failed, models.BatchFailure{ID: id, Reason: "request cancelled"})
			continue
		}
		p, err := s.get(ctx, id)
		if err == nil && p.Field == models.FieldIGPost {
			err = invalid("not a field change (use the Instagram posts page)")
		}
		if err == nil && p.Status != models.StatusPending {
			err = ErrProposalNotPending
		}
		if err == nil {
			switch {
			case action == BatchApprove && models.IsInformational(p.Field):
				res.Skipped = append(res.Skipped, id)
				continue
			case action == BatchApprove:
				_, err = s.approve(ctx, id, nil)
			default:
				err = s.reject(ctx, id)
			}
		}
		if err != nil {
			res.Failed = append(res.Failed, models.BatchFailure{ID: id, Reason: batchReason(id, err)})
			continue
		}
		res.Done = append(res.Done, id)
	}
	return res
}

func batchReason(id int64, err error) string {
	switch {
	case errors.Is(err, ErrProposalNotPending):
		return "already resolved"
	case IsNotFound(err):
		return "not found"
	case IsValidation(err):
		return err.Error()
	}
	slog.Error("batch change failed", "id", id, "err", err)
	return "unexpected error"
}

// BatchSummary renders the flash text, e.g.
// "Approved 12, skipped 2 informational, failed 1 (menu must be an http(s) URL)".
func BatchSummary(res models.BatchResult) string {
	verb := "Approved"
	if res.Action == BatchReject {
		verb = "Rejected"
	}
	parts := []string{fmt.Sprintf("%s %d", verb, len(res.Done))}
	if n := len(res.Skipped); n > 0 {
		parts = append(parts, fmt.Sprintf("skipped %d informational (dismiss them instead)", n))
	}
	if n := len(res.Failed); n > 0 {
		var reasons []string
		seen := map[string]bool{}
		for _, f := range res.Failed {
			if !seen[f.Reason] {
				seen[f.Reason] = true
				reasons = append(reasons, f.Reason)
			}
		}
		if len(reasons) > 3 {
			reasons = append(reasons[:3], "…")
		}
		parts = append(parts, fmt.Sprintf("failed %d (%s)", n, strings.Join(reasons, "; ")))
	}
	return strings.Join(parts, ", ")
}
