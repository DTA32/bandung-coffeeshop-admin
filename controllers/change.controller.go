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

// List returns pending field proposals grouped by cafe.
func (c *ChangeController) List(ctx context.Context, cafeID int) ([]models.CafeProposals, error) {
	items, err := c.repo.ListPending(ctx, cafeID)
	if err != nil {
		return nil, err
	}
	return GroupByCafe(items), nil
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
// Value is used for single-value fields, Open/Close for hours.
type Edit struct {
	Value string
	Open  string
	Close string
}

// FieldUpdateFor turns a proposal (and an optional admin edit) into the
// validated write plus the value recorded in resolved_value.
func FieldUpdateFor(p models.Proposal, edit *Edit) (models.FieldUpdate, any, error) {
	if models.IsInformational(p.Field) || p.Field == models.FieldIGPost {
		return models.FieldUpdate{}, nil, invalid("%s cannot be approved; dismiss it instead", p.Field)
	}

	if p.Field == models.FieldHours {
		var hv models.HoursValue
		if edit != nil {
			hv.Open, hv.Close = strPtr(edit.Open), strPtr(edit.Close)
		} else if err := json.Unmarshal(p.NewValue, &hv); err != nil {
			return models.FieldUpdate{}, nil, fmt.Errorf("decode hours: %w", err)
		}
		hv.Open, hv.Close = NilIfBlank(hv.Open), NilIfBlank(hv.Close)
		if hv.Open == nil || hv.Close == nil {
			return models.FieldUpdate{}, nil, invalid("both open and close hours are required")
		}
		if err := validateHour(hv.Open); err != nil {
			return models.FieldUpdate{}, nil, err
		}
		if err := validateHour(hv.Close); err != nil {
			return models.FieldUpdate{}, nil, err
		}
		return models.FieldUpdate{OpenHour: hv.Open, CloseHour: hv.Close},
			map[string]string{"open": *hv.Open, "close": *hv.Close}, nil
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
