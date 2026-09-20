// Package session orchestrates the pipeline: deterministic search + Gemini
// scoring/evolution + SQLite persistence. It is the only package that ties
// llm, search and store together, so api handlers stay thin.
package session

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"talent-search-rubric-arm/backend/internal/data"
	"talent-search-rubric-arm/backend/internal/domain"
	"talent-search-rubric-arm/backend/internal/llm"
	"talent-search-rubric-arm/backend/internal/search"
	"talent-search-rubric-arm/backend/internal/store"
)

const maxShownResults = 5

// Stage reports pipeline progress so the API layer can push it to the client
// over SSE — this is what powers the "thinking" state.
type Stage func(stage, message string)

func noopStage(string, string) {}

type Service struct {
	store    *store.Store
	llm      *llm.Client
	profiles *data.Store
}

func New(st *store.Store, llmClient *llm.Client, profiles *data.Store) *Service {
	return &Service{store: st, llm: llmClient, profiles: profiles}
}

// Draft turns a free-text query into a Filters+Rubric pair. Nothing is
// persisted until Run is called with the (possibly recruiter-edited) draft.
func (s *Service) Draft(ctx context.Context, query string, onStage Stage) (domain.Filters, domain.Rubric, error) {
	if onStage == nil {
		onStage = noopStage
	}
	onStage("reading_query", "Reading your query…")
	onStage("drafting", "Drafting filters & rubric…")

	return s.llm.Draft(ctx, query)
}

// StartSession persists a new session for the given original query. Called
// once Draft has succeeded, right before Run.
func (s *Service) StartSession(ctx context.Context, query string) (domain.Session, error) {
	return s.store.CreateSession(ctx, query)
}

// Run executes the deterministic filter + LLM scoring pipeline against the
// given filters/rubric and persists the result as the next version. It is
// used both for the very first search (version 1) and for manual filter or
// rubric edits (version N).
func (s *Service) Run(ctx context.Context, sessionID string, filters domain.Filters, rubric domain.Rubric, onStage Stage) (domain.Version, error) {
	if onStage == nil {
		onStage = noopStage
	}

	sess, err := s.store.GetSession(ctx, sessionID)
	if err != nil {
		return domain.Version{}, fmt.Errorf("loading session: %w", err)
	}

	nextVN, err := s.store.NextVersionNumber(ctx, sessionID)
	if err != nil {
		return domain.Version{}, fmt.Errorf("computing next version number: %w", err)
	}
	trigger := domain.TriggerManualEdit
	if nextVN == 1 {
		trigger = domain.TriggerInitial
	}

	return s.searchAndScore(ctx, sess, filters, rubric, trigger, "", "", onStage)
}

// Evolve interprets recruiter feedback (chat text and/or per-profile
// reactions), updates filters/rubric via the LLM, and re-runs the pipeline.
// The recruiter's message is persisted even if the LLM call subsequently
// fails, so nothing typed is lost.
func (s *Service) Evolve(ctx context.Context, sessionID, message string, reactions []domain.Reaction, onStage Stage) (domain.Version, error) {
	if onStage == nil {
		onStage = noopStage
	}

	sess, err := s.store.GetSession(ctx, sessionID)
	if err != nil {
		return domain.Version{}, fmt.Errorf("loading session: %w", err)
	}

	latest, err := s.store.LatestVersion(ctx, sessionID)
	if err != nil {
		return domain.Version{}, fmt.Errorf("loading latest version: %w", err)
	}

	userChatContent := describeFeedback(message, reactions)
	if _, err := s.store.AddChatMessage(ctx, domain.ChatMessage{
		SessionID: sessionID,
		Role:      domain.RoleUser,
		Content:   userChatContent,
	}); err != nil {
		return domain.Version{}, fmt.Errorf("saving chat message: %w", err)
	}

	onStage("understanding", "Understanding your feedback…")

	signals, err := s.store.ListLikedSignals(ctx, sessionID)
	if err != nil {
		return domain.Version{}, fmt.Errorf("loading liked signals: %w", err)
	}

	resultContexts, err := s.buildEvolveContext(latest)
	if err != nil {
		return domain.Version{}, err
	}

	newFilters, newRubric, changeSummary, err := s.llm.Evolve(
		ctx, sess.OriginalQuery, latest.Filters, latest.Rubric, resultContexts, message, reactions, signals,
	)
	if err != nil {
		return domain.Version{}, err
	}

	onStage("searching", "Searching profiles…")
	version, err := s.searchAndScore(ctx, sess, newFilters, newRubric, domain.TriggerChat, userChatContent, changeSummary, onStage)
	if err != nil {
		return domain.Version{}, err
	}

	if _, err := s.store.AddChatMessage(ctx, domain.ChatMessage{
		SessionID: sessionID,
		VersionID: version.ID,
		Role:      domain.RoleAssistant,
		Content:   changeSummary,
	}); err != nil {
		return domain.Version{}, fmt.Errorf("saving assistant chat message: %w", err)
	}

	return version, nil
}

func (s *Service) searchAndScore(ctx context.Context, sess domain.Session, filters domain.Filters, rubric domain.Rubric, trigger domain.TriggerType, triggerText, changeSummary string, onStage Stage) (domain.Version, error) {
	onStage("searching", "Searching profiles…")
	filterResult := search.Run(s.profiles.Profiles, filters)

	nextVN, err := s.store.NextVersionNumber(ctx, sess.ID)
	if err != nil {
		return domain.Version{}, fmt.Errorf("computing next version number: %w", err)
	}

	version := domain.Version{
		SessionID:       sess.ID,
		VersionNumber:   nextVN,
		TriggerType:     trigger,
		TriggerText:     triggerText,
		Filters:         filters,
		Rubric:          rubric,
		ChangeSummary:   changeSummary,
		EmptyDiagnostic: filterResult.EmptyDiagnostic,
	}

	if len(filterResult.Candidates) == 0 {
		return s.store.CreateVersion(ctx, version)
	}

	onStage("scoring", "Scoring & explaining matches…")

	signals, err := s.store.ListLikedSignals(ctx, sess.ID)
	if err != nil {
		return domain.Version{}, fmt.Errorf("loading liked signals: %w", err)
	}

	results, err := s.llm.Score(ctx, sess.OriginalQuery, rubric, filterResult.Candidates, signals)
	if err != nil {
		return domain.Version{}, err
	}

	version.Results = topResults(results)
	return s.store.CreateVersion(ctx, version)
}

func topResults(results []domain.ScoredResult) []domain.ScoredResult {
	sort.SliceStable(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	n := maxShownResults
	if len(results) < n {
		n = len(results)
	}
	top := make([]domain.ScoredResult, n)
	copy(top, results[:n])
	for i := range top {
		top[i].Rank = i + 1
	}
	return top
}

func (s *Service) buildEvolveContext(v domain.Version) ([]llm.EvolveResultContext, error) {
	out := make([]llm.EvolveResultContext, 0, len(v.Results))
	for _, r := range v.Results {
		profile, ok := s.profiles.Get(r.ProfileID)
		if !ok {
			continue
		}
		out = append(out, llm.EvolveResultContext{Rank: r.Rank, Profile: profile, Result: r})
	}
	return out, nil
}

func describeFeedback(message string, reactions []domain.Reaction) string {
	message = strings.TrimSpace(message)
	if len(reactions) == 0 {
		return message
	}
	var parts []string
	for _, r := range reactions {
		verdict := "good fit"
		if r.Verdict == "not_fit" {
			verdict = "not a fit"
		}
		part := fmt.Sprintf("profile %s: %s", r.ProfileID, verdict)
		if r.Note != "" {
			part += fmt.Sprintf(" (%s)", r.Note)
		}
		parts = append(parts, part)
	}
	reactionText := strings.Join(parts, "; ")
	if message == "" {
		return reactionText
	}
	return fmt.Sprintf("%s — %s", message, reactionText)
}

func (s *Service) Freeze(ctx context.Context, sessionID string) (domain.Session, error) {
	latest, err := s.store.LatestVersion(ctx, sessionID)
	if err != nil {
		return domain.Session{}, fmt.Errorf("loading latest version: %w", err)
	}
	if err := s.store.FreezeSession(ctx, sessionID, latest.VersionNumber); err != nil {
		return domain.Session{}, err
	}
	return s.store.GetSession(ctx, sessionID)
}

func (s *Service) LikeSignal(ctx context.Context, sig domain.LikedSignal) (domain.LikedSignal, error) {
	return s.store.AddLikedSignal(ctx, sig)
}

func (s *Service) UnlikeSignal(ctx context.Context, sessionID, signalID string) error {
	return s.store.RemoveLikedSignal(ctx, sessionID, signalID)
}

func (s *Service) GetSessionDetail(ctx context.Context, sessionID string) (domain.SessionDetail, error) {
	return s.store.GetSessionDetail(ctx, sessionID)
}

func (s *Service) ListSessions(ctx context.Context) ([]domain.SessionSummary, error) {
	return s.store.ListSessions(ctx)
}

func (s *Service) GetProfile(id string) (data.Profile, bool) {
	return s.profiles.Get(id)
}

// ListSkills returns every distinct skill across the dataset, sorted, so the
// frontend can offer a searchable multi-select instead of free text.
func (s *Service) ListSkills() []string {
	seen := make(map[string]bool)
	out := []string{}
	for _, p := range s.profiles.Profiles {
		for _, sk := range p.Skills {
			if !seen[sk] {
				seen[sk] = true
				out = append(out, sk)
			}
		}
	}
	sort.Strings(out)
	return out
}

// ListLocations returns every distinct location across the dataset, sorted,
// so the frontend can offer a searchable multi-select instead of free text.
func (s *Service) ListLocations() []string {
	seen := make(map[string]bool)
	out := []string{}
	for _, p := range s.profiles.Profiles {
		if !seen[p.Location] {
			seen[p.Location] = true
			out = append(out, p.Location)
		}
	}
	sort.Strings(out)
	return out
}

// ListTitles returns every distinct current_title across the dataset, sorted,
// so the frontend can offer a searchable multi-select for title keywords.
func (s *Service) ListTitles() []string {
	seen := make(map[string]bool)
	out := []string{}
	for _, p := range s.profiles.Profiles {
		if !seen[p.CurrentTitle] {
			seen[p.CurrentTitle] = true
			out = append(out, p.CurrentTitle)
		}
	}
	sort.Strings(out)
	return out
}

// ListCompanyTypes returns every distinct current_company_type across the
// dataset, sorted, so the frontend can offer a searchable multi-select.
func (s *Service) ListCompanyTypes() []string {
	seen := make(map[string]bool)
	out := []string{}
	for _, p := range s.profiles.Profiles {
		if !seen[p.CurrentCompanyType] {
			seen[p.CurrentCompanyType] = true
			out = append(out, p.CurrentCompanyType)
		}
	}
	sort.Strings(out)
	return out
}
