package adapter

type GeminiAdapter struct{}

func (GeminiAdapter) Name() string { return "gemini" }
func (GeminiAdapter) SupportsSkills() bool {
	return false
}
func (GeminiAdapter) SkillDir() string {
	return ""
}
