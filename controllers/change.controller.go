package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/repository"
)

type ChangeController struct {
	repo *repository.ChangeRepository
}

func NewChangeController(repo *repository.ChangeRepository) *ChangeController {
	return &ChangeController{repo: repo}
}

func (c *ChangeController) Get(ctx context.Context, id int64) (*models.Proposal, error) {
	return c.repo.Get(ctx, id)
}

// List returns one page of pending field proposals grouped by cafe (a page is
// PageSize cafes; groups are never split).
// List returns the groups of PageSize cafes after afterCafeID (0 = first page).
func (c *ChangeController) List(ctx context.Context, cafeID, afterCafeID int) (models.Page[models.CafeProposals], error) {
	items, more, err := c.repo.ListPendingPage(ctx, cafeID, afterCafeID)
	if err != nil {
		return models.Page[models.CafeProposals]{}, err
	}
	return models.Page[models.CafeProposals]{Items: GroupByCafe(items), HasMore: more}, nil
}

func GroupByCafe(items []models.Proposal) []models.CafeProposals {
	var out []models.CafeProposals
	for _, p := range items {
		if n := len(out); n == 0 || out[n-1].CafeID != p.CafeID {
			out = append(out, models.CafeProposals{CafeID: p.CafeID, LocationID: p.LocationID, CafeName: p.CafeName})
		}
		out[len(out)-1].Items = append(out[len(out)-1].Items, p)
	}
	return out
}

// Edit is the admin's replacement for a proposed value ("edit & approve").
// Value is used for single-value fields, Hours (the per-day editor's rows)
// for hours; the hours window always comes from the proposal.
type Edit struct {
	Value string
	Hours []models.CafeHour
}

// FieldUpdateFor turns a proposal (and an optional admin edit) into the
// validated write plus the value recorded in resolved_value.
func FieldUpdateFor(p models.Proposal, edit *Edit) (models.FieldUpdate, any, error) {
	if models.IsInformational(p.Field) || p.Field == models.FieldIGPost {
		return models.FieldUpdate{}, nil, invalid("%s cannot be approved; dismiss it instead", p.Field)
	}

	if p.Field == models.FieldHours {
		hp, err := hoursUpdate(p, edit)
		if err != nil {
			return models.FieldUpdate{}, nil, err
		}
		return models.FieldUpdate{Hours: &hp}, hp, nil
	}

	var v string
	if edit != nil {
		v = edit.Value
	} else if err := json.Unmarshal(p.NewValue, &v); err != nil {
		return models.FieldUpdate{}, nil, fmt.Errorf("decode %s: %w", p.Field, err)
	}
	v = strings.TrimSpace(v)
	if v == "" {
		return models.FieldUpdate{}, nil, invalid("%s must not be empty", p.Field)
	}
	switch p.Field {
	case models.FieldMenu:
		if !isHTTPURL(v) {
			return models.FieldUpdate{}, nil, invalid("menu must be an http(s) URL")
		}
	case models.FieldStatus:
		if v != "active" && v != "closed" {
			return models.FieldUpdate{}, nil, invalid("status must be active or closed")
		}
	case models.FieldPhone, models.FieldAddress:
	default:
		return models.FieldUpdate{}, nil, invalid("unknown field %q", p.Field)
	}
	return models.FieldUpdate{Text: &v}, v, nil
}

// hoursUpdate validates an hours proposal (or the admin's edit of it). The
// rows replace the cafe's undated rows and its dated rows overlapping the
// window, so a dated row must overlap the window too. The result (window +
// normalised rows, no raw days) is also the resolved_value.
func hoursUpdate(p models.Proposal, edit *Edit) (models.HoursProposal, error) {
	hp, ok := models.ParseHoursProposal(p.NewValue)
	if !ok {
		return models.HoursProposal{}, invalid("this hours proposal is in an old format; reject it")
	}
	rows := hp.Rows
	if edit != nil {
		rows = edit.Hours
	}
	rows, err := NormalizeHours(rows)
	if err != nil {
		return models.HoursProposal{}, err
	}
	for _, h := range rows {
		if h.Dated() && !models.ReplacedByWindow(h, hp.WindowFrom, hp.WindowUntil) {
			return models.HoursProposal{}, invalid("%s is outside this proposal's dates (%s – %s); add it on the cafe form",
				HourRowLabel(h), hp.WindowFrom, hp.WindowUntil)
		}
	}
	return models.HoursProposal{WindowFrom: hp.WindowFrom, WindowUntil: hp.WindowUntil, Rows: rows}, nil
}

func (c *ChangeController) Approve(ctx context.Context, id int64, edit *Edit) (*models.Proposal, error) {
	p, err := c.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.Status != models.StatusPending {
		return p, ErrProposalNotPending
	}
	u, resolved, err := FieldUpdateFor(*p, edit)
	if err != nil {
		return p, err
	}
	return p, c.repo.Apply(ctx, id, u, resolved)
}

func (c *ChangeController) Reject(ctx context.Context, id int64) error {
	return c.repo.Reject(ctx, id)
}

// ApproveAllForCafe applies every pending, non-informational proposal of a
// cafe as proposed. It stops at the first error.
func (c *ChangeController) ApproveAllForCafe(ctx context.Context, cafeID int) (int, error) {
	items, err := c.repo.ListPending(ctx, cafeID)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, p := range items {
		if models.IsInformational(p.Field) {
			continue
		}
		if _, err := c.Approve(ctx, p.ID, nil); err != nil {
			return n, fmt.Errorf("%s: %w", p.Field, err)
		}
		n++
	}
	return n, nil
}
