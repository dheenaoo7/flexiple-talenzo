package data

import (
	"encoding/json"
	"fmt"
	"os"
)

type PastCompany struct {
	Company     string `json:"company"`
	CompanyType string `json:"company_type"`
	Title       string `json:"title"`
	Years       int    `json:"years"`
}

type Profile struct {
	ID                 string        `json:"id"`
	Name               string        `json:"name"`
	CurrentTitle       string        `json:"current_title"`
	YearsExperience    int           `json:"years_experience"`
	Location           string        `json:"location"`
	CurrentCompany     string        `json:"current_company"`
	CurrentCompanyType string        `json:"current_company_type"`
	Skills             []string      `json:"skills"`
	PastCompanies      []PastCompany `json:"past_companies"`
	Education          string        `json:"education"`
	Summary            string        `json:"summary"`
}

// Store holds the full profile dataset in memory, loaded once at startup.
type Store struct {
	Profiles []Profile
	byID     map[string]Profile
}

func LoadProfiles(path string) (*Store, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading profiles file: %w", err)
	}

	var profiles []Profile
	if err := json.Unmarshal(raw, &profiles); err != nil {
		return nil, fmt.Errorf("parsing profiles json: %w", err)
	}

	byID := make(map[string]Profile, len(profiles))
	for _, p := range profiles {
		byID[p.ID] = p
	}

	return &Store{Profiles: profiles, byID: byID}, nil
}

func (s *Store) Get(id string) (Profile, bool) {
	p, ok := s.byID[id]
	return p, ok
}
