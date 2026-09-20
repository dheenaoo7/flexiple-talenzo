// Package search applies the deterministic, editable Filters to the in-memory
// profile dataset. Ranking/explanation of the resulting candidates is the
// LLM's job (see internal/llm); this package only narrows the pool.
package search

import (
	"fmt"
	"sort"
	"strings"

	"talent-search-rubric-arm/backend/internal/data"
	"talent-search-rubric-arm/backend/internal/domain"
)

const maxCandidatesForLLM = 20

type Result struct {
	Candidates      []data.Profile
	TotalMatched    int
	Capped          bool
	EmptyDiagnostic *domain.EmptyResultDiagnostic
}

func Run(profiles []data.Profile, f domain.Filters) Result {
	candidates := make([]data.Profile, 0, len(profiles))
	for _, p := range profiles {
		if matches(p, f) {
			candidates = append(candidates, p)
		}
	}

	total := len(candidates)
	capped := false
	if total > maxCandidatesForLLM {
		candidates = rankBySkillOverlap(candidates, f)[:maxCandidatesForLLM]
		capped = true
	}

	var diag *domain.EmptyResultDiagnostic
	if total == 0 {
		diag = diagnoseEmptyResult(profiles, f)
	}

	return Result{Candidates: candidates, TotalMatched: total, Capped: capped, EmptyDiagnostic: diag}
}

func matches(p data.Profile, f domain.Filters) bool {
	if len(f.RequiredSkills) > 0 && !hasAllSkills(p.Skills, f.RequiredSkills) {
		return false
	}
	if len(f.ExcludedSkills) > 0 && hasAnySkill(p.Skills, f.ExcludedSkills) {
		return false
	}
	if f.MinYearsExperience != nil && p.YearsExperience < *f.MinYearsExperience {
		return false
	}
	if f.MaxYearsExperience != nil && p.YearsExperience > *f.MaxYearsExperience {
		return false
	}
	if len(f.Locations) > 0 && !containsFold(f.Locations, p.Location) {
		return false
	}
	if len(f.CompanyTypes) > 0 && !containsFold(f.CompanyTypes, p.CurrentCompanyType) {
		return false
	}
	if len(f.TitleKeywords) > 0 && !anySubstringFold(f.TitleKeywords, p.CurrentTitle) {
		return false
	}
	return true
}

// rankBySkillOverlap heuristically orders already-filtered candidates so that
// capping to maxCandidatesForLLM keeps the strongest-looking profiles rather
// than an arbitrary prefix. Real scoring/explanation still happens via the LLM.
func rankBySkillOverlap(candidates []data.Profile, f domain.Filters) []data.Profile {
	wanted := append(append([]string{}, f.RequiredSkills...), f.NiceToHaveSkills...)
	scored := make([]data.Profile, len(candidates))
	copy(scored, candidates)

	overlap := func(p data.Profile) int {
		count := 0
		for _, w := range wanted {
			if containsFold(p.Skills, w) {
				count++
			}
		}
		return count
	}

	sort.SliceStable(scored, func(i, j int) bool {
		oi, oj := overlap(scored[i]), overlap(scored[j])
		if oi != oj {
			return oi > oj
		}
		return scored[i].YearsExperience > scored[j].YearsExperience
	})
	return scored
}

// diagnoseEmptyResult reports which single filter, applied in isolation over
// the full dataset, eliminates the most profiles — the most actionable thing
// to relax first.
func diagnoseEmptyResult(profiles []data.Profile, f domain.Filters) *domain.EmptyResultDiagnostic {
	type check struct {
		label   string
		exclude func(data.Profile) bool
	}

	checks := []check{}
	if len(f.RequiredSkills) > 0 {
		checks = append(checks, check{
			label:   fmt.Sprintf("required skills: %s", strings.Join(f.RequiredSkills, ", ")),
			exclude: func(p data.Profile) bool { return !hasAllSkills(p.Skills, f.RequiredSkills) },
		})
	}
	if len(f.ExcludedSkills) > 0 {
		checks = append(checks, check{
			label:   fmt.Sprintf("excluded skills: %s", strings.Join(f.ExcludedSkills, ", ")),
			exclude: func(p data.Profile) bool { return hasAnySkill(p.Skills, f.ExcludedSkills) },
		})
	}
	if f.MinYearsExperience != nil {
		min := *f.MinYearsExperience
		checks = append(checks, check{
			label:   fmt.Sprintf("minimum years experience: %d", min),
			exclude: func(p data.Profile) bool { return p.YearsExperience < min },
		})
	}
	if f.MaxYearsExperience != nil {
		max := *f.MaxYearsExperience
		checks = append(checks, check{
			label:   fmt.Sprintf("maximum years experience: %d", max),
			exclude: func(p data.Profile) bool { return p.YearsExperience > max },
		})
	}
	if len(f.Locations) > 0 {
		checks = append(checks, check{
			label:   fmt.Sprintf("locations: %s", strings.Join(f.Locations, ", ")),
			exclude: func(p data.Profile) bool { return !containsFold(f.Locations, p.Location) },
		})
	}
	if len(f.CompanyTypes) > 0 {
		checks = append(checks, check{
			label:   fmt.Sprintf("company types: %s", strings.Join(f.CompanyTypes, ", ")),
			exclude: func(p data.Profile) bool { return !containsFold(f.CompanyTypes, p.CurrentCompanyType) },
		})
	}
	if len(f.TitleKeywords) > 0 {
		checks = append(checks, check{
			label:   fmt.Sprintf("title keywords: %s", strings.Join(f.TitleKeywords, ", ")),
			exclude: func(p data.Profile) bool { return !anySubstringFold(f.TitleKeywords, p.CurrentTitle) },
		})
	}

	if len(checks) == 0 {
		return &domain.EmptyResultDiagnostic{
			MostRestrictiveFilter: "none",
			Explanation:           "No filters matched any of the 48 profiles for an unexpected reason.",
		}
	}

	worstLabel := ""
	worstCount := -1
	for _, c := range checks {
		count := 0
		for _, p := range profiles {
			if c.exclude(p) {
				count++
			}
		}
		if count > worstCount {
			worstCount = count
			worstLabel = c.label
		}
	}

	return &domain.EmptyResultDiagnostic{
		MostRestrictiveFilter: worstLabel,
		Explanation: fmt.Sprintf(
			"%s excluded %d of %d profiles on its own — try relaxing it first.",
			worstLabel, worstCount, len(profiles),
		),
	}
}

func hasAllSkills(profileSkills, required []string) bool {
	for _, r := range required {
		if !containsFold(profileSkills, r) {
			return false
		}
	}
	return true
}

func hasAnySkill(profileSkills, anyOf []string) bool {
	for _, s := range anyOf {
		if containsFold(profileSkills, s) {
			return true
		}
	}
	return false
}

func containsFold(list []string, target string) bool {
	for _, item := range list {
		if strings.EqualFold(item, target) {
			return true
		}
	}
	return false
}

func anySubstringFold(needles []string, haystack string) bool {
	lower := strings.ToLower(haystack)
	for _, n := range needles {
		if strings.Contains(lower, strings.ToLower(n)) {
			return true
		}
	}
	return false
}
