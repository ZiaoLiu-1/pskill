package skill

type Skill struct {
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	Body        string   `json:"body" yaml:"body"`
	Path        string   `json:"path" yaml:"path"`
	SourceCLI   string   `json:"sourceCli" yaml:"sourceCli"`
	Tags        []string `json:"tags" yaml:"tags"`
	InstalledIn []string `json:"installedIn" yaml:"installedIn"`
	UsageCount  int64    `json:"usageCount" yaml:"usageCount"`
	LastUsedAt  string   `json:"lastUsedAt" yaml:"lastUsedAt"`
}

type Frontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	License     string `yaml:"license,omitempty"`
}
