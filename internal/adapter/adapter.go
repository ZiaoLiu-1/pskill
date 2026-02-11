package adapter

type Adapter interface {
	Name() string
	SkillDir() string
	SupportsSkills() bool
}

func All() map[string]Adapter {
	return map[string]Adapter{
		"cursor": CursorAdapter{},
		"claude": ClaudeAdapter{},
		"codex":  CodexAdapter{},
		"gemini": GeminiAdapter{},
	}
}
