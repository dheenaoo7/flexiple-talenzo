// Package domain holds the shared types passed between search, llm, store,
// session and api — the vocabulary of a talent search session.
package domain

import "time"

// Filters is the deterministic, editable narrowing criteria for a search.
type Filters struct {
	TitleKeywords      []string `json:"title_keywords"`
	MinYearsExperience *int     `json:"min_years_experience"`
	MaxYearsExperience *int     `json:"max_years_experience"`
	Locations          []string `json:"locations"`
	CompanyTypes       []string `json:"company_types"`
	RequiredSkills     []string `json:"required_skills"`
	NiceToHaveSkills   []string `json:"nice_to_have_skills"`
	ExcludedSkills     []string `json:"excluded_skills"`
	Keywords           string   `json:"keywords"`
}

// RubricCriterion is a single weighted thing the LLM should look for.
type RubricCriterion struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Weight      int    `json:"weight"` // 1-5
}

type Rubric struct {
	Criteria []RubricCriterion `json:"criteria"`
}

// MatchedPoint ties a rubric criterion to real evidence in a specific profile.
type MatchedPoint struct {
	CriterionID    string `json:"criterion_id"`
	CriterionLabel string `json:"criterion_label"`
	EvidenceText   string `json:"evidence_text"`
}

type ScoredResult struct {
	ProfileID     string         `json:"profile_id"`
	Rank          int            `json:"rank"`
	Score         float64        `json:"score"`
	SummaryLine   string         `json:"summary_line"`
	MatchedPoints []MatchedPoint `json:"matched_points"`
}

type EmptyResultDiagnostic struct {
	MostRestrictiveFilter string `json:"most_restrictive_filter"`
	Explanation           string `json:"explanation"`
}

type TriggerType string

const (
	TriggerInitial    TriggerType = "initial"
	TriggerManualEdit TriggerType = "manual_edit"
	TriggerChat       TriggerType = "chat"
)

type Version struct {
	ID              string                 `json:"id"`
	SessionID       string                 `json:"session_id"`
	VersionNumber   int                    `json:"version_number"`
	TriggerType     TriggerType            `json:"trigger_type"`
	TriggerText     string                 `json:"trigger_text,omitempty"`
	Filters         Filters                `json:"filters"`
	Rubric          Rubric                 `json:"rubric"`
	ChangeSummary   string                 `json:"change_summary,omitempty"`
	Results         []ScoredResult         `json:"results"`
	EmptyDiagnostic *EmptyResultDiagnostic `json:"empty_diagnostic,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
}

type SessionStatus string

const (
	StatusActive SessionStatus = "active"
	StatusFrozen SessionStatus = "frozen"
)

type Session struct {
	ID                  string        `json:"id"`
	OriginalQuery       string        `json:"original_query"`
	CreatedAt           time.Time     `json:"created_at"`
	Status              SessionStatus `json:"status"`
	FrozenVersionNumber *int          `json:"frozen_version_number,omitempty"`
}

type ChatRole string

const (
	RoleUser      ChatRole = "user"
	RoleAssistant ChatRole = "assistant"
)

type ChatMessage struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	VersionID string    `json:"version_id,omitempty"`
	Role      ChatRole  `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Reaction is a per-profile fit/not-fit tag the recruiter can send instead of
// (or alongside) free chat text.
type Reaction struct {
	ProfileID string `json:"profile_id"`
	Verdict   string `json:"verdict"` // "fit" | "not_fit"
	Note      string `json:"note,omitempty"`
}

// LikedSignal is a soft signal: the recruiter liked a specific matched point
// on a specific profile. It feeds future LLM context/weighting — it never
// becomes a hard filter.
type LikedSignal struct {
	ID             string    `json:"id"`
	SessionID      string    `json:"session_id"`
	VersionID      string    `json:"version_id"`
	ProfileID      string    `json:"profile_id"`
	CriterionID    string    `json:"criterion_id"`
	CriterionLabel string    `json:"criterion_label"`
	EvidenceText   string    `json:"evidence_text"`
	CreatedAt      time.Time `json:"created_at"`
}

// SessionDetail is the full hierarchy returned by GET /api/sessions/:id.
type SessionDetail struct {
	Session      Session       `json:"session"`
	Versions     []Version     `json:"versions"`
	ChatMessages []ChatMessage `json:"chat_messages"`
	LikedSignals []LikedSignal `json:"liked_signals"`
}

type SessionSummary struct {
	ID            string        `json:"id"`
	OriginalQuery string        `json:"original_query"`
	CreatedAt     time.Time     `json:"created_at"`
	Status        SessionStatus `json:"status"`
}
