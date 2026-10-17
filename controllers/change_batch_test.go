package controllers

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func TestParseBatch(t *testing.T) {
	a, ids, err := ParseBatch("approve", []string{"3", " 1 ", "3", "2"})
	if err != nil || a != "approve" || !reflect.DeepEqual(ids, []int64{3, 1, 2}) {
		t.Fatalf("ok case: %v %v %v", a, ids, err)
	}
	for name, c := range map[string]struct {
		action string
		ids    []string
	}{
		"bad action":   {"delete", []string{"1"}},
		"empty action": {"", []string{"1"}},
		"no ids":       {"reject", nil},
		"non-numeric":  {"reject", []string{"1", "x"}},
		"zero":         {"reject", []string{"0"}},
		"negative":     {"reject", []string{"-4"}},
	} {
		if _, _, err := ParseBatch(c.action, c.ids); !IsValidation(err) {
			t.Errorf("%s: err = %v, want validation error", name, err)
		}
	}
	many := make([]string, MaxBatch+1)
	for i := range many {
		many[i] = fmt.Sprint(i + 1)
	}
	if _, _, err := ParseBatch("approve", many); !IsValidation(err) {
		t.Errorf("cap: %v", err)
	}
	if _, ids, err := ParseBatch("approve", many[:MaxBatch]); err != nil || len(ids) != MaxBatch {
		t.Errorf("exactly the cap must pass: %d %v", len(ids), err)
	}
	// duplicates do not count towards the cap
	dup := append(append([]string{}, many[:MaxBatch]...), "1", "2")
	if _, ids, err := ParseBatch("approve", dup); err != nil || len(ids) != MaxBatch {
		t.Errorf("dedup before cap: %d %v", len(ids), err)
	}
}

// fakeQueue is an in-memory stand-in for the single-row controller path.
type fakeQueue struct {
	props    map[int64]*models.Proposal
	approved []int64
	rejected []int64
}

func (f *fakeQueue) steps() batchSteps {
	return batchSteps{
		get: func(_ context.Context, id int64) (*models.Proposal, error) {
			if p, ok := f.props[id]; ok {
				return p, nil
			}
			return nil, ErrNotFound
		},
		approve: func(_ context.Context, id int64, edit *Edit) (*models.Proposal, error) {
			p := f.props[id]
			if _, _, err := FieldUpdateFor(*p, edit); err != nil { // the real validation
				return p, err
			}
			f.approved = append(f.approved, id)
			return p, nil
		},
		reject: func(_ context.Context, id int64) error {
			if id == 99 {
				return errors.New("db exploded")
			}
			f.rejected = append(f.rejected, id)
			return nil
		},
	}
}

func prop(id int64, field, newValue, status string) *models.Proposal {
	return &models.Proposal{ID: id, Field: field, NewValue: []byte(newValue), Status: status}
}

func TestRunBatchApprovePartialFailure(t *testing.T) {
	f := &fakeQueue{props: map[int64]*models.Proposal{
		1: prop(1, models.FieldPhone, `"+62811"`, models.StatusPending),
		2: prop(2, models.FieldMenu, `"not a url"`, models.StatusPending), // validation error
		3: prop(3, "ig_bio", `"bio"`, models.StatusPending),               // informational
		4: prop(4, models.FieldPhone, `"+62822"`, models.StatusApproved),  // stale
		5: prop(5, models.FieldIGPost, `{}`, models.StatusPending),        // not a field change
		6: prop(6, models.FieldMenu, `"https://x.test/m"`, models.StatusPending),
	}}
	res := runBatch(context.Background(), BatchApprove, []int64{1, 2, 3, 4, 5, 6, 7}, f.steps())

	if !reflect.DeepEqual(res.Done, []int64{1, 6}) || !reflect.DeepEqual(f.approved, []int64{1, 6}) {
		t.Fatalf("done = %v approved = %v", res.Done, f.approved)
	}
	if !reflect.DeepEqual(res.Skipped, []int64{3}) {
		t.Fatalf("skipped = %v", res.Skipped)
	}
	reasons := map[int64]string{}
	for _, fl := range res.Failed {
		reasons[fl.ID] = fl.Reason
	}
	if len(res.Failed) != 4 || reasons[2] != "menu must be an http(s) URL" || reasons[4] != "already resolved" ||
		reasons[7] != "not found" || !strings.Contains(reasons[5], "not a field change") {
		t.Fatalf("failed = %+v", res.Failed)
	}
	sum := BatchSummary(res)
	for _, want := range []string{"Approved 2", "skipped 1 informational", "failed 4 ("} {
		if !strings.Contains(sum, want) {
			t.Errorf("summary %q lacks %q", sum, want)
		}
	}
}

func TestRunBatchRejectDismissesInformationalAndContinuesAfterError(t *testing.T) {
	f := &fakeQueue{props: map[int64]*models.Proposal{
		1:  prop(1, "ig_bio", `"x"`, models.StatusPending),
		99: prop(99, models.FieldPhone, `"+1"`, models.StatusPending), // reject errors
		3:  prop(3, models.FieldPhone, `"+3"`, models.StatusPending),
	}}
	res := runBatch(context.Background(), BatchReject, []int64{1, 99, 3}, f.steps())
	if !reflect.DeepEqual(res.Done, []int64{1, 3}) || len(res.Skipped) != 0 {
		t.Fatalf("done = %v skipped = %v", res.Done, res.Skipped)
	}
	if len(res.Failed) != 1 || res.Failed[0].ID != 99 || res.Failed[0].Reason != "unexpected error" {
		t.Fatalf("failed = %+v", res.Failed)
	}
	if sum := BatchSummary(res); sum != "Rejected 2, failed 1 (unexpected error)" {
		t.Fatalf("summary = %q", sum)
	}
}

func TestBatchSummaryTruncatesReasons(t *testing.T) {
	res := models.BatchResult{Action: BatchApprove}
	for i, r := range []string{"a", "b", "c", "d", "a"} {
		res.Failed = append(res.Failed, models.BatchFailure{ID: int64(i + 1), Reason: r})
	}
	if got := BatchSummary(res); got != "Approved 0, failed 5 (a; b; c; …)" {
		t.Fatalf("got %q", got)
	}
}
