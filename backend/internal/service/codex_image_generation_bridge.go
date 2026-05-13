package service

import "strings"

const featureKeyCodexImageGenerationBridge = "codex_image_generation_bridge"

func boolOverridePtr(v bool) *bool {
	return &v
}

func boolOverrideFromMap(values map[string]any, keys ...string) *bool {
	if values == nil {
		return nil
	}
	for _, key := range keys {
		if v, ok := values[key].(bool); ok {
			return boolOverridePtr(v)
		}
	}
	return nil
}

func platformBoolOverride(values map[string]any, key string, platform string) *bool {
	if values == nil {
		return nil
	}
	if v, ok := values[key].(bool); ok {
		return boolOverridePtr(v)
	}
	raw, ok := values[key].(map[string]any)
	if !ok {
		return nil
	}
	platform = strings.TrimSpace(platform)
	if platform == "" {
		return nil
	}
	if v, ok := raw[platform].(bool); ok {
		return boolOverridePtr(v)
	}
	return nil
}

func (c *Channel) CodexImageGenerationBridgeOverride(platform string) *bool {
	if c == nil {
		return nil
	}
	return platformBoolOverride(c.FeaturesConfig, featureKeyCodexImageGenerationBridge, platform)
}

func (a *Account) CodexImageGenerationBridgeOverride() *bool {
	if a == nil || a.Platform != PlatformOpenAI || a.Extra == nil {
		return nil
	}
	if override := boolOverrideFromMap(a.Extra, featureKeyCodexImageGenerationBridge, "codex_image_generation_bridge_enabled"); override != nil {
		return override
	}
	openAIConfig, _ := a.Extra[PlatformOpenAI].(map[string]any)
	return boolOverrideFromMap(openAIConfig, featureKeyCodexImageGenerationBridge, "codex_image_generation_bridge_enabled")
}
