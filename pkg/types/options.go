package types

// Config struct with nested FindTarget structure.
type Config struct {
	FindTarget struct {
		BugCrowd  *BugCrowdConfig  `yaml:"bugcrowd"`
		HackerOne *HackerOneConfig `yaml:"hackerone"`
	} `yaml:"findtarget"`
	Proxy    string `yaml:"proxy"`
	Template string // No yaml tag needed for command line flags
}

// SetDefaults assigns default values for the entire Config struct.
func (c *Config) SetDefaults() {
	if c.FindTarget.BugCrowd != nil {
		c.FindTarget.BugCrowd.SetDefaults()
	}
}
