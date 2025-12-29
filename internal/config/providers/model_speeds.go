package providers

// ModelSpeed represents tokens per second output speed for a model
type ModelSpeed struct {
	TokensPerSecond float64
	SpeedTier       string // ultra-fast, fast, medium-fast, medium, slow
}

// ModelSpeedRegistry contains benchmarked speed data for AI models
// Sources: Cerebras official, ArtificialAnalysis.ai, DataStudios (Sept 2025)
var ModelSpeedRegistry = map[string]ModelSpeed{
	// Cerebras - Ultra-Fast (Specialized hardware)
	"llama-4-scout-17b-16e-instruct": {TokensPerSecond: 2600, SpeedTier: "ultra-fast"},
	"llama-3.3-70b":                  {TokensPerSecond: 2100, SpeedTier: "ultra-fast"},
	"qwen-3-32b":                     {TokensPerSecond: 1700, SpeedTier: "ultra-fast"}, // estimated
	"gpt-oss-120b":                   {TokensPerSecond: 1500, SpeedTier: "ultra-fast"}, // estimated

	// xAI Grok - Ultra-Fast/Fast
	"grok-4-fast":             {TokensPerSecond: 183.6, SpeedTier: "ultra-fast"},
	"grok-4-1-fast":           {TokensPerSecond: 180, SpeedTier: "ultra-fast"}, // estimated
	"grok-code-fast":          {TokensPerSecond: 170, SpeedTier: "ultra-fast"}, // estimated
	"grok-4-1-fast-reasoning": {TokensPerSecond: 150, SpeedTier: "ultra-fast"}, // estimated (reasoning)

	// Kimi - Fast
	"moonshotai/kimi-k2-thinking": {TokensPerSecond: 173, SpeedTier: "fast"},
	"moonshotai/kimi-k2-instruct": {TokensPerSecond: 120, SpeedTier: "fast"}, // estimated
	"kimi-k1.5":                   {TokensPerSecond: 100, SpeedTier: "fast"}, // estimated

	// OpenAI - Fast
	"gpt-4o-mini":  {TokensPerSecond: 55, SpeedTier: "fast"},
	"gpt-5-2":      {TokensPerSecond: 55, SpeedTier: "fast"},
	"gpt-5":        {TokensPerSecond: 48, SpeedTier: "fast"},
	"gpt-4.1":      {TokensPerSecond: 45, SpeedTier: "fast"}, // estimated
	"gpt-4o":       {TokensPerSecond: 45, SpeedTier: "fast"},
	"gpt-4o-audio": {TokensPerSecond: 40, SpeedTier: "fast"}, // estimated

	// Gemini - Fast/Medium-Fast
	"gemini-2.5-flash":      {TokensPerSecond: 50, SpeedTier: "fast"}, // estimated
	"gemini-2.0-flash":      {TokensPerSecond: 45, SpeedTier: "fast"}, // estimated
	"gemini-2.5-flash-lite": {TokensPerSecond: 55, SpeedTier: "fast"}, // estimated (lite)
	"gemini-2.5-pro":        {TokensPerSecond: 38, SpeedTier: "medium-fast"},
	"gemini-2.0-pro-exp":    {TokensPerSecond: 35, SpeedTier: "medium-fast"}, // estimated
	"gemini-3-pro-preview":  {TokensPerSecond: 40, SpeedTier: "fast"},        // estimated

	// xAI Grok - Medium/Fast
	"grok-4":      {TokensPerSecond: 35, SpeedTier: "medium-fast"},
	"grok-3":      {TokensPerSecond: 30, SpeedTier: "medium"},
	"grok-3-mini": {TokensPerSecond: 40, SpeedTier: "fast"}, // estimated (mini faster)

	// Anthropic Claude - Medium
	"claude-opus-4-5-20251101":   {TokensPerSecond: 32, SpeedTier: "medium"},
	"claude-sonnet-4-5-20250929": {TokensPerSecond: 33, SpeedTier: "medium"},
	"claude-haiku-4-5-20251001":  {TokensPerSecond: 45, SpeedTier: "fast"},        // estimated (Haiku traditionally faster)
	"claude-opus-4-1-20250805":   {TokensPerSecond: 30, SpeedTier: "medium"},      // estimated
	"claude-sonnet-3-5-20241022": {TokensPerSecond: 32, SpeedTier: "medium"},      // estimated
	"claude-sonnet-3-5-20240620": {TokensPerSecond: 30, SpeedTier: "medium"},      // estimated
	"claude-haiku-3-5-20241022":  {TokensPerSecond: 40, SpeedTier: "fast"},        // estimated
	"claude-haiku-3-0-20240307":  {TokensPerSecond: 38, SpeedTier: "medium-fast"}, // estimated

	// DeepSeek - Medium
	"deepseek-chat":     {TokensPerSecond: 35, SpeedTier: "medium-fast"},
	"deepseek-reasoner": {TokensPerSecond: 25, SpeedTier: "medium"}, // reasoning slower

	// Mistral - Medium/Fast
	"mistral-large-2512":  {TokensPerSecond: 35, SpeedTier: "medium-fast"}, // estimated
	"mistral-large-2411":  {TokensPerSecond: 33, SpeedTier: "medium"},      // estimated
	"pixtral-large-2411":  {TokensPerSecond: 30, SpeedTier: "medium"},      // vision slower
	"mistral-medium-2508": {TokensPerSecond: 40, SpeedTier: "fast"},        // estimated (medium tier)
	"mistral-small-2501":  {TokensPerSecond: 50, SpeedTier: "fast"},        // estimated (small faster)
	"ministral-8b-2512":   {TokensPerSecond: 60, SpeedTier: "fast"},        // estimated (tiny model)
	"ministral-3b-2512":   {TokensPerSecond: 70, SpeedTier: "fast"},        // estimated (very tiny)
	"codestral-latest":    {TokensPerSecond: 45, SpeedTier: "fast"},        // estimated
	"codestral-2501":      {TokensPerSecond: 45, SpeedTier: "fast"},        // estimated
	"devstral-2512":       {TokensPerSecond: 40, SpeedTier: "fast"},        // estimated
	"devstral-small-2512": {TokensPerSecond: 50, SpeedTier: "fast"},        // estimated

	// Z.AI GLM - Medium
	"glm-4.7":        {TokensPerSecond: 35, SpeedTier: "medium-fast"}, // estimated
	"glm-4.6":        {TokensPerSecond: 33, SpeedTier: "medium"},      // estimated
	"glm-4.5":        {TokensPerSecond: 32, SpeedTier: "medium"},      // estimated
	"glm-4.5-air":    {TokensPerSecond: 45, SpeedTier: "fast"},        // estimated (lite)
	"glm-4.6v":       {TokensPerSecond: 28, SpeedTier: "medium"},      // vision slower
	"glm-4.5v":       {TokensPerSecond: 28, SpeedTier: "medium"},      // vision slower
	"glm-4.5-flash":  {TokensPerSecond: 50, SpeedTier: "fast"},        // estimated (flash)
	"glm-4.6v-flash": {TokensPerSecond: 48, SpeedTier: "fast"},        // estimated (flash vision)

	// MiniMax - Medium
	"minimax-m2.1":       {TokensPerSecond: 35, SpeedTier: "medium-fast"}, // estimated
	"minimax-m2":         {TokensPerSecond: 33, SpeedTier: "medium"},      // estimated
	"text-01":            {TokensPerSecond: 40, SpeedTier: "fast"},        // estimated
	"vl-01":              {TokensPerSecond: 30, SpeedTier: "medium"},      // vision slower
	"abab7-chat-preview": {TokensPerSecond: 25, SpeedTier: "medium"},      // large model
	"abab6.5g-chat":      {TokensPerSecond: 30, SpeedTier: "medium"},      // estimated
	"abab6.5s-chat":      {TokensPerSecond: 45, SpeedTier: "fast"},        // speed variant

	// Kimi Legacy - Medium/Fast
	"moonshot-v1-8k":             {TokensPerSecond: 50, SpeedTier: "fast"},   // small context
	"moonshot-v1-32k":            {TokensPerSecond: 40, SpeedTier: "fast"},   // estimated
	"moonshot-v1-128k":           {TokensPerSecond: 30, SpeedTier: "medium"}, // large context
	"moonshot-v1-vision-preview": {TokensPerSecond: 28, SpeedTier: "medium"}, // vision

	// OpenAI Reasoning Models - Slow (prioritize accuracy over speed)
	"o1":      {TokensPerSecond: 20, SpeedTier: "slow"},
	"o1-mini": {TokensPerSecond: 25, SpeedTier: "slow"},
	"o3":      {TokensPerSecond: 15, SpeedTier: "slow"},
	"o3-mini": {TokensPerSecond: 20, SpeedTier: "slow"},
	"o4-mini": {TokensPerSecond: 22, SpeedTier: "slow"},
}

// GetModelSpeed returns the speed data for a model ID, with fallback defaults
func GetModelSpeed(modelID string) ModelSpeed {
	if speed, ok := ModelSpeedRegistry[modelID]; ok {
		return speed
	}

	// Default fallback based on model name patterns
	// This ensures we have reasonable defaults for models not in registry
	return ModelSpeed{
		TokensPerSecond: 30, // conservative default
		SpeedTier:       "medium",
	}
}

// GetFastestModel returns the fastest model from a list of model IDs
func GetFastestModel(modelIDs []string) (string, ModelSpeed) {
	if len(modelIDs) == 0 {
		return "", ModelSpeed{}
	}

	fastest := modelIDs[0]
	fastestSpeed := GetModelSpeed(fastest)

	for _, modelID := range modelIDs[1:] {
		speed := GetModelSpeed(modelID)
		if speed.TokensPerSecond > fastestSpeed.TokensPerSecond {
			fastest = modelID
			fastestSpeed = speed
		}
	}

	return fastest, fastestSpeed
}
