package models

import "time"

// PostValue is new_value for FieldIGPost.
type PostValue struct {
	Caption   string      `json:"caption"`
	TakenAt   time.Time   `json:"taken_at"`
	Owner     string      `json:"owner"`
	IsCollab  bool        `json:"is_collab"`
	Permalink string      `json:"permalink"`
	Media     []PostMedia `json:"media"`
}

type PostMedia struct {
	Type  string `json:"type"` // image | video
	R2Key string `json:"r2_key"`
}

type Post struct {
	Proposal
	Value PostValue
}
