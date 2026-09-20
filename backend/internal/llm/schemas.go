package llm

import "google.golang.org/genai"

func stringArraySchema(desc string) *genai.Schema {
	return &genai.Schema{
		Type:        genai.TypeArray,
		Description: desc,
		Items:       &genai.Schema{Type: genai.TypeString},
	}
}

func filtersSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"title_keywords":       stringArraySchema("Job title keywords that would appear in a matching candidate's current title, e.g. \"Backend Engineer\". Empty array if titles shouldn't be restricted."),
			"min_years_experience": {Type: genai.TypeInteger, Nullable: boolPtr(true), Description: "Minimum years of experience required, or null if no minimum."},
			"max_years_experience": {Type: genai.TypeInteger, Nullable: boolPtr(true), Description: "Maximum years of experience allowed, or null if no maximum."},
			"locations":            stringArraySchema("Acceptable candidate locations. Empty array if location doesn't matter."),
			"company_types":        stringArraySchema("Acceptable current_company_type values: startup, scaleup, enterprise, or agency. Empty array if it doesn't matter."),
			"required_skills":      stringArraySchema("Skills the candidate MUST have (hard filter). Keep this short and only include truly non-negotiable skills."),
			"nice_to_have_skills":  stringArraySchema("Skills that are a bonus but not required."),
			"excluded_skills":      stringArraySchema("Skills that should disqualify a candidate if present."),
			"keywords":             {Type: genai.TypeString, Description: "A short free-text phrase capturing extra nuance from the query (e.g. domain like \"fintech\") that isn't a hard filter."},
		},
		Required:         []string{"title_keywords", "locations", "company_types", "required_skills", "nice_to_have_skills", "excluded_skills", "keywords"},
		PropertyOrdering: []string{"title_keywords", "min_years_experience", "max_years_experience", "locations", "company_types", "required_skills", "nice_to_have_skills", "excluded_skills", "keywords"},
	}
}

func rubricSchema() *genai.Schema {
	return &genai.Schema{
		Type:        genai.TypeObject,
		Description: "A short rubric (3-6 criteria) the recruiter can read at a glance.",
		Properties: map[string]*genai.Schema{
			"criteria": {
				Type:     genai.TypeArray,
				Items:    rubricCriterionSchema(),
				MinItems: int64Ptr(3),
				MaxItems: int64Ptr(6),
			},
		},
		Required:         []string{"criteria"},
		PropertyOrdering: []string{"criteria"},
	}
}

func rubricCriterionSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"id":          {Type: genai.TypeString, Description: "A short stable slug id, e.g. \"aws_rds_depth\"."},
			"label":       {Type: genai.TypeString, Description: "Short human-readable label, e.g. \"Direct AWS RDS production experience\"."},
			"description": {Type: genai.TypeString, Description: "One sentence explaining what to look for."},
			"weight":      {Type: genai.TypeInteger, Description: "Importance from 1 (nice-to-have) to 5 (critical)."},
		},
		Required:         []string{"id", "label", "description", "weight"},
		PropertyOrdering: []string{"id", "label", "description", "weight"},
	}
}

func draftResponseSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"filters": filtersSchema(),
			"rubric":  rubricSchema(),
		},
		Required:         []string{"filters", "rubric"},
		PropertyOrdering: []string{"filters", "rubric"},
	}
}

func matchedPointSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"criterion_id":    {Type: genai.TypeString},
			"criterion_label": {Type: genai.TypeString},
			"evidence_text":   {Type: genai.TypeString, Description: "A specific, concrete detail quoted or closely paraphrased from THIS profile's actual skills/summary/companies that supports the criterion. Never invent details."},
		},
		Required:         []string{"criterion_id", "criterion_label", "evidence_text"},
		PropertyOrdering: []string{"criterion_id", "criterion_label", "evidence_text"},
	}
}

func scoreResponseSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"results": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"profile_id":     {Type: genai.TypeString},
						"score":          {Type: genai.TypeNumber, Description: "Fit score from 0-100 against the rubric."},
						"summary_line":   {Type: genai.TypeString, Description: "One specific sentence summarizing the fit, referencing real profile details."},
						"matched_points": {Type: genai.TypeArray, Items: matchedPointSchema()},
					},
					Required:         []string{"profile_id", "score", "summary_line", "matched_points"},
					PropertyOrdering: []string{"profile_id", "score", "summary_line", "matched_points"},
				},
			},
		},
		Required:         []string{"results"},
		PropertyOrdering: []string{"results"},
	}
}

func evolveResponseSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"filters":        filtersSchema(),
			"rubric":         rubricSchema(),
			"change_summary": {Type: genai.TypeString, Description: "Plain-English bullet-style summary of what changed vs the previous version and why, written to the recruiter (e.g. \"Raised minimum experience to 5 years because you flagged profile 1 as too junior.\")."},
		},
		Required:         []string{"filters", "rubric", "change_summary"},
		PropertyOrdering: []string{"filters", "rubric", "change_summary"},
	}
}

func boolPtr(b bool) *bool    { return &b }
func int64Ptr(i int64) *int64 { return &i }
