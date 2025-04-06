package types

import (
	"time"
)

type HackerOneResponse struct {
	Data []struct {
		Attributes struct {
			Name   string `json:"name"`
			Handle string `json:"handle"`
			Policy string `json:"policy"`
		} `json:"attributes"`
	} `json:"data"`
	Links HackerOneLink `json:"links"`
}

type HackerOneLink struct {
	Next string `json:"next"`
}

type HackerOneConfig struct {
	Reward      string `yaml:"reward"`
	Category    string `yaml:"category"`
	Scope       string `yaml:"scope"`
	MaxPrograms int8   `yaml:"maxPrograms"`
	Include     []string `yaml:"include"`
	H1Username  string   `yaml:"h1Username"`
	H1Token  string   `yaml:"h1Token"`
}

type H1ProgramStruct struct {
	Data  []H1ScopeData `json:"data"`
	Links H1Links       `json:"links"`
}

type H1ScopeData struct {
	ID         string       `json:"id"`
	Type       string       `json:"type"`
	Attributes H1Attributes `json:"attributes"`
}

type H1Attributes struct {
	AssetType                  string    `json:"asset_type"`
	AssetIdentifier            string    `json:"asset_identifier"`
	EligibleForBounty          bool      `json:"eligible_for_bounty"`
	EligibleForSubmission      bool      `json:"eligible_for_submission"`
	Instruction                string    `json:"instruction"`
	MaxSeverity                string    `json:"max_severity"`
	CreatedAt                  time.Time `json:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at"`
	ConfidentialityRequirement *string   `json:"confidentiality_requirement,omitempty"`
	IntegrityRequirement       *string   `json:"integrity_requirement,omitempty"`
	AvailabilityRequirement    *string   `json:"availability_requirement,omitempty"`
}

type H1Links struct {
	Self string `json:"self"`
	Next string `json:"next"`
	Last string `json:"last"`
}

func (b *HackerOneConfig) SetDefaults() {
	if b.Scope == "" { // Ensuring WideScope is explicitly set to false
		b.Scope = "all"
	}
}
