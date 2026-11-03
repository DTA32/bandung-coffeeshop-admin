package controllers

import (
	"context"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

// ParsePostBatch validates POST /posts/batch: the action must be publish or
// ignore; ids follow the ParseBatch rules (positive, deduplicated, 1..MaxBatch).
func ParsePostBatch(action string, rawIDs []string) (string, []int64, error) {
	if action != BatchPublish && action != BatchIgnore {
		return "", nil, invalid("unknown batch action %q", action)
	}
	ids, err := parseBatchIDs(rawIDs, "post")
	if err != nil {
		return "", nil, err
	}
	return action, ids, nil
}

// postBatchSteps are the single-post operations a batch runs; the real ones
// are the controller's own Get / CreateFeed (with DefaultFeedInput) / Ignore.
type postBatchSteps struct {
	get     func(ctx context.Context, id int64) (*models.Post, error)
	publish func(ctx context.Context, post models.Post) error
	ignore  func(ctx context.Context, id int64) error
}

// ApplyBatch publishes or ignores each post through the single-post path.
// Publish creates the feed entry exactly as the "Create feed" form would be
// prefilled (DefaultFeedInput) — one transaction per post. One failing post
// never aborts the rest. The caller validates with ParsePostBatch.
func (c *PostController) ApplyBatch(ctx context.Context, action string, ids []int64, d FeedDefaults) models.BatchResult {
	return runPostBatch(ctx, action, ids, postBatchSteps{
		get: c.Get,
		publish: func(ctx context.Context, post models.Post) error {
			_, err := c.CreateFeed(ctx, post, DefaultFeedInput(post, d), d.ImageBase, d.TZ)
			return err
		},
		ignore: c.Ignore,
	})
}

func runPostBatch(ctx context.Context, action string, ids []int64, s postBatchSteps) models.BatchResult {
	res := models.BatchResult{Action: action}
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			res.Failed = append(res.Failed, models.BatchFailure{ID: id, Reason: "request cancelled"})
			continue
		}
		post, err := s.get(ctx, id)
		if err == nil && post.Status != models.StatusPending {
			err = ErrProposalNotPending
		}
		if err == nil {
			switch {
			case action == BatchPublish && strings.TrimSpace(post.Value.Caption) == "":
				err = invalid("no caption (use Create feed to write one)")
			case action == BatchPublish:
				err = s.publish(ctx, *post)
			default:
				err = s.ignore(ctx, id)
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
