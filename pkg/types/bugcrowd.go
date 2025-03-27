package types

// BugCrowdConfig struct
type BugCrowdConfig struct {
	Reward      string `yaml:"reward"`
	Category    string `yaml:"category"`
	Scope       string `yaml:"scope"`
	MaxPrograms int16  `yaml:"maxPrograms"`
	Include     []string `yaml:"include"`
}

func (b *BugCrowdConfig) SetDefaults() {
	if b.Scope == "" { // Ensuring WideScope is explicitly set to false
		b.Scope = "all"
	}
}

// Define the structure of the response JSON
type PaginationMeta struct {
	TotalCount int `json:"totalCount"`
	Limit      int `json:"limit"`
}

type APIResponse struct {
	PaginationMeta struct {
		TotalCount int `json:"totalCount"`
		Limit      int `json:"limit"`
	} `json:"paginationMeta"`
	Engagements []Engagement `json:"engagements"`
}

// Engagement represents a single engagement program.
type Engagement struct {
	Name     string `json:"name"`
	BriefURL string `json:"briefUrl"`
}

type ScopeItem struct {
	Name    string   `json:"name"`
	Targets []Target `json:"targets"`
}

type Target struct {
	Name     string `json:"name"`
	URI      string `json:"uri"`
	Category string `json:"category"`
}
