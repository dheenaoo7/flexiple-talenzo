package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"talent-search-rubric-arm/backend/internal/data"
	"talent-search-rubric-arm/backend/internal/domain"
)

const rubricArmPrimer = `You are the reasoning engine behind Rubric-ARM, a recruiter talent-search tool.
Rubric-ARM works in two structured artifacts:
- Filters: hard, deterministic constraints applied in code to narrow a candidate pool (skills, years of experience, location, company type, title).
- Rubric: a short set of weighted, human-readable criteria used to score and explain how well a candidate fits, beyond the hard filters.
Always output strict JSON matching the given schema. Be concrete and specific — never generic praise. Base every claim about a candidate only on the profile data you are given; never invent details.`

type draftResponse struct {
	Filters domain.Filters `json:"filters"`
	Rubric  domain.Rubric  `json:"rubric"`
}

func (c *Client) Draft(ctx context.Context, query string) (domain.Filters, domain.Rubric, error) {
	prompt := fmt.Sprintf(`A recruiter typed this free-text search query:

%q

Turn it into Filters and a Rubric per your instructions. The Filters should only capture things you are confident are hard requirements; put everything else, including softer signals, into the Rubric as weighted criteria. Keep the rubric to 3-6 criteria that a recruiter could read in a few seconds.`, query)

	var out draftResponse
	if err := generateJSON(ctx, c, "drafting filters & rubric", rubricArmPrimer, prompt, draftResponseSchema(), &out); err != nil {
		return domain.Filters{}, domain.Rubric{}, err
	}
	return out.Filters, out.Rubric, nil
}

type scoreResponse struct {
	Results []domain.ScoredResult `json:"results"`
}

func (c *Client) Score(ctx context.Context, query string, rubric domain.Rubric, candidates []data.Profile, likedSignals []domain.LikedSignal) ([]domain.ScoredResult, error) {
	if len(candidates) == 0 {
		return nil, nil
	}

	candidatesJSON, err := json.MarshalIndent(candidates, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling candidates: %w", err)
	}
	rubricJSON, _ := json.MarshalIndent(rubric, "", "  ")

	signalsBlock := "None yet."
	if len(likedSignals) > 0 {
		var lines []string
		for _, s := range likedSignals {
			lines = append(lines, fmt.Sprintf("- On a previous profile, the recruiter liked the point: %q (criterion: %s)", s.EvidenceText, s.CriterionLabel))
		}
		signalsBlock = strings.Join(lines, "\n")
	}

	prompt := fmt.Sprintf(`Original recruiter query: %q

Rubric to score against:
%s

Signals the recruiter has already confirmed matter to them (weight these more heavily, but they are not hard filters):
%s

Candidate profiles (already passed the hard filters):
%s

For EACH candidate, produce a score 0-100, a one-sentence summary_line, and 2-4 matched_points. Every matched_point's evidence_text MUST reference a real, specific detail actually present in that candidate's data (a skill, a past company, years of experience, education, or a phrase from their summary) — never generic praise like "strong engineer" and never invent facts.`, query, rubricJSON, signalsBlock, candidatesJSON)

	var out scoreResponse
	if err := generateJSON(ctx, c, "scoring & explaining matches", rubricArmPrimer, prompt, scoreResponseSchema(), &out); err != nil {
		return nil, err
	}

	byID := make(map[string]data.Profile, len(candidates))
	for _, p := range candidates {
		byID[p.ID] = p
	}

	results := make([]domain.ScoredResult, 0, len(out.Results))
	for _, r := range out.Results {
		profile, ok := byID[r.ProfileID]
		if !ok {
			continue // model referenced a profile_id we didn't send; drop it rather than show a broken card
		}
		r.MatchedPoints = groundMatchedPoints(r.MatchedPoints, profile)
		results = append(results, r)
	}
	return results, nil
}

type evolveResponse struct {
	Filters       domain.Filters `json:"filters"`
	Rubric        domain.Rubric  `json:"rubric"`
	ChangeSummary string         `json:"change_summary"`
}

// EvolveResultContext is one currently-shown profile, given to the Evolve
// prompt so recruiter feedback like "1 is too junior" can be resolved to a
// real candidate.
type EvolveResultContext struct {
	Rank    int                 `json:"rank"`
	Profile data.Profile        `json:"profile"`
	Result  domain.ScoredResult `json:"result"`
}

func (c *Client) Evolve(
	ctx context.Context,
	query string,
	currentFilters domain.Filters,
	currentRubric domain.Rubric,
	currentResults []EvolveResultContext,
	message string,
	reactions []domain.Reaction,
	likedSignals []domain.LikedSignal,
) (domain.Filters, domain.Rubric, string, error) {
	filtersJSON, _ := json.MarshalIndent(currentFilters, "", "  ")
	rubricJSON, _ := json.MarshalIndent(currentRubric, "", "  ")
	resultsJSON, _ := json.MarshalIndent(currentResults, "", "  ")

	reactionsBlock := "None."
	if len(reactions) > 0 {
		var lines []string
		for _, r := range reactions {
			line := fmt.Sprintf("- profile_id %s: %s", r.ProfileID, r.Verdict)
			if r.Note != "" {
				line += fmt.Sprintf(" (%s)", r.Note)
			}
			lines = append(lines, line)
		}
		reactionsBlock = strings.Join(lines, "\n")
	}

	messageBlock := "(no free-text message)"
	if strings.TrimSpace(message) != "" {
		messageBlock = message
	}

	signalsBlock := "None yet."
	if len(likedSignals) > 0 {
		var lines []string
		for _, s := range likedSignals {
			lines = append(lines, fmt.Sprintf("- Liked: %q (criterion: %s)", s.EvidenceText, s.CriterionLabel))
		}
		signalsBlock = strings.Join(lines, "\n")
	}

	prompt := fmt.Sprintf(`Original recruiter query: %q

Current Filters:
%s

Current Rubric:
%s

Currently shown ranked results (rank, full profile data, and the current score/explanation):
%s

Recruiter's chat message:
%s

Recruiter's per-profile reactions:
%s

Confirmed signals from earlier likes:
%s

Update the Filters and/or Rubric to reflect this feedback. Resolve any reference to a rank number (e.g. "1", "profile 2") using the ranked results above. Make the smallest change that addresses the feedback — don't rewrite things the recruiter didn't ask about. Then write change_summary as 1-4 short bullet-style sentences a recruiter would read, stating exactly what changed and why (referencing their feedback).`, query, filtersJSON, rubricJSON, resultsJSON, messageBlock, reactionsBlock, signalsBlock)

	var out evolveResponse
	if err := generateJSON(ctx, c, "understanding feedback & updating filters/rubric", rubricArmPrimer, prompt, evolveResponseSchema(), &out); err != nil {
		return domain.Filters{}, domain.Rubric{}, "", err
	}
	return out.Filters, out.Rubric, out.ChangeSummary, nil
}

// groundMatchedPoints drops any matched point whose evidence doesn't share at
// least one meaningful token with the profile's real data, guarding against
// hallucinated "evidence" slipping into the UI.
func groundMatchedPoints(points []domain.MatchedPoint, p data.Profile) []domain.MatchedPoint {
	haystack := strings.ToLower(strings.Join(profileTextFields(p), " "))
	haystackTokens := tokenSet(haystack)

	grounded := make([]domain.MatchedPoint, 0, len(points))
	for _, pt := range points {
		for _, tok := range tokenize(strings.ToLower(pt.EvidenceText)) {
			if len(tok) < 4 {
				continue
			}
			if haystackTokens[tok] {
				grounded = append(grounded, pt)
				break
			}
		}
	}
	return grounded
}

func profileTextFields(p data.Profile) []string {
	fields := []string{p.CurrentTitle, p.CurrentCompany, p.CurrentCompanyType, p.Education, p.Summary}
	fields = append(fields, p.Skills...)
	for _, pc := range p.PastCompanies {
		fields = append(fields, pc.Company, pc.CompanyType, pc.Title)
	}
	return fields
}

func tokenize(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
	})
}

func tokenSet(s string) map[string]bool {
	out := map[string]bool{}
	for _, tok := range tokenize(s) {
		out[tok] = true
	}
	return out
}
