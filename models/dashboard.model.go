package models

type Dashboard struct {
	Cafes          int
	PendingChanges int
	PendingPosts   int
	PendingReviews int
	FailingTargets int
	ActiveUsers    int
	DeletedUsers   int
	LastRuns       []ScrapeRun
}
