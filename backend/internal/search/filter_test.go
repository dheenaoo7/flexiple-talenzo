package search

import (
	"testing"

	"talent-search-rubric-arm/backend/internal/data"
	"talent-search-rubric-arm/backend/internal/domain"
)

func intPtr(i int) *int { return &i }

func testProfiles() []data.Profile {
	return []data.Profile{
		{ID: "p1", CurrentTitle: "Senior Backend Engineer", YearsExperience: 6, Location: "Bangalore", CurrentCompanyType: "startup", Skills: []string{"AWS RDS", "PostgreSQL", "Node.js"}},
		{ID: "p2", CurrentTitle: "Backend Engineer", YearsExperience: 2, Location: "Mumbai", CurrentCompanyType: "enterprise", Skills: []string{"Python", "Django"}},
		{ID: "p3", CurrentTitle: "Frontend Engineer", YearsExperience: 4, Location: "Bangalore", CurrentCompanyType: "scaleup", Skills: []string{"React", "TypeScript"}},
		{ID: "p4", CurrentTitle: "Senior Backend Engineer", YearsExperience: 8, Location: "Remote - India", CurrentCompanyType: "startup", Skills: []string{"AWS RDS", "PostgreSQL", "Redis"}},
	}
}

func TestRun_CombinedFilters(t *testing.T) {
	profiles := testProfiles()
	f := domain.Filters{
		RequiredSkills:     []string{"AWS RDS", "PostgreSQL"},
		MinYearsExperience: intPtr(5),
		TitleKeywords:      []string{"Backend Engineer"},
		CompanyTypes:       []string{"startup"},
	}

	res := Run(profiles, f)

	if res.TotalMatched != 2 {
		t.Fatalf("expected 2 matches (p1, p4), got %d", res.TotalMatched)
	}
	ids := map[string]bool{}
	for _, c := range res.Candidates {
		ids[c.ID] = true
	}
	if !ids["p1"] || !ids["p4"] {
		t.Fatalf("expected p1 and p4 in candidates, got %+v", res.Candidates)
	}
	if res.EmptyDiagnostic != nil {
		t.Fatalf("expected no empty diagnostic, got %+v", res.EmptyDiagnostic)
	}
}

func TestRun_ExcludedSkills(t *testing.T) {
	profiles := testProfiles()
	f := domain.Filters{ExcludedSkills: []string{"Django"}}
	res := Run(profiles, f)
	for _, c := range res.Candidates {
		if c.ID == "p2" {
			t.Fatalf("p2 has excluded skill Django and should not match")
		}
	}
	if res.TotalMatched != 3 {
		t.Fatalf("expected 3 matches, got %d", res.TotalMatched)
	}
}

func TestRun_EmptyResultDiagnosesMostRestrictiveFilter(t *testing.T) {
	profiles := testProfiles()
	// min years experience of 100 excludes all 4; required skill "AWS RDS" excludes 2 (p2, p3).
	f := domain.Filters{
		MinYearsExperience: intPtr(100),
		RequiredSkills:     []string{"AWS RDS"},
	}
	res := Run(profiles, f)

	if res.TotalMatched != 0 {
		t.Fatalf("expected 0 matches, got %d", res.TotalMatched)
	}
	if res.EmptyDiagnostic == nil {
		t.Fatal("expected an empty result diagnostic")
	}
	if res.EmptyDiagnostic.MostRestrictiveFilter != "minimum years experience: 100" {
		t.Fatalf("expected years-experience to be the most restrictive filter, got %q", res.EmptyDiagnostic.MostRestrictiveFilter)
	}
}

func TestRun_CapsCandidatesAtTwentyAndRanksBySkillOverlap(t *testing.T) {
	var profiles []data.Profile
	for i := 0; i < 30; i++ {
		skills := []string{"Python"}
		if i < 25 {
			skills = append(skills, "AWS RDS")
		}
		profiles = append(profiles, data.Profile{
			ID:              "p" + string(rune('a'+i)),
			CurrentTitle:    "Backend Engineer",
			YearsExperience: 3,
			Skills:          skills,
		})
	}

	f := domain.Filters{RequiredSkills: []string{"Python"}, NiceToHaveSkills: []string{"AWS RDS"}}
	res := Run(profiles, f)

	if res.TotalMatched != 30 {
		t.Fatalf("expected 30 matches before capping, got %d", res.TotalMatched)
	}
	if !res.Capped {
		t.Fatal("expected result to be capped")
	}
	if len(res.Candidates) != maxCandidatesForLLM {
		t.Fatalf("expected %d candidates after cap, got %d", maxCandidatesForLLM, len(res.Candidates))
	}
	for _, c := range res.Candidates {
		found := false
		for _, s := range c.Skills {
			if s == "AWS RDS" {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected capped candidates to prefer AWS RDS overlap, got profile without it: %+v", c)
		}
	}
}
