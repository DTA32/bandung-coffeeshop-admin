package models

import (
	"encoding/json"
	"time"
)

var Sources = []string{"gmaps", "instagram", "linktree"}

type SourceTarget struct {
	CafeID              int
	LocationID          string
	CafeName            string
	Source              string
	ExternalID          *string
	Enabled             bool
	LastSuccessAt       *time.Time
	LastError           *string
	ConsecutiveFailures int
	LatestSnapshotID    *int64
	LatestSnapshotAt    *time.Time
}

// TargetFilter narrows the source target list (zero values = no filter).
type TargetFilter struct {
	Source      string
	FailingOnly bool
	CafeID      int
}

type Snapshot struct {
	ID         int64
	CafeID     int
	LocationID string
	CafeName   string
	Source     string
	FetchedAt  time.Time
	Raw        json.RawMessage
	Parsed     json.RawMessage
	Hash       *string
}

type ScrapeRun struct {
	ID         int64
	Source     string
	StartedAt  time.Time
	FinishedAt *time.Time
	OK         int
	Failed     int
	Proposals  int
	Error      *string
}
