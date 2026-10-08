package models

type Dashboard struct {
	Cafes          int
	PendingChanges int
	PendingPosts   int
	FailingTargets int
	LastRuns       []ScrapeRun
}
