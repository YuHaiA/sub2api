package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveHealthCheckModelIDUsesPlatformDefaultsForMismatchedModels(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		account  *Account
		modelID  string
		expected string
	}{
		{name: "empty model stays empty", account: &Account{Platform: PlatformOpenAI}, modelID: "", expected: ""},
		{name: "openai model on openai account", account: &Account{Platform: PlatformOpenAI}, modelID: "gpt-5.5", expected: "gpt-5.5"},
		{name: "openai model on anthropic account falls back", account: &Account{Platform: PlatformAnthropic}, modelID: "gpt-5.5", expected: ""},
		{name: "claude model on anthropic account", account: &Account{Platform: PlatformAnthropic}, modelID: "claude-sonnet-4-5", expected: "claude-sonnet-4-5"},
		{name: "claude model on gemini account falls back", account: &Account{Platform: PlatformGemini}, modelID: "claude-sonnet-4-5", expected: ""},
		{name: "gemini model on gemini account", account: &Account{Platform: PlatformGemini}, modelID: "gemini-2.5-pro", expected: "gemini-2.5-pro"},
		{name: "gemini model on antigravity account", account: &Account{Platform: PlatformAntigravity}, modelID: "gemini-2.5-pro", expected: "gemini-2.5-pro"},
		{name: "claude model on antigravity account", account: &Account{Platform: PlatformAntigravity}, modelID: "claude-sonnet-4-5", expected: "claude-sonnet-4-5"},
		{name: "grok model on grok account", account: &Account{Platform: PlatformGrok}, modelID: "grok-4.3", expected: "grok-4.3"},
		{name: "openai model on grok account falls back", account: &Account{Platform: PlatformGrok}, modelID: "gpt-5.5", expected: ""},
		{name: "custom model only applied to openai", account: &Account{Platform: PlatformOpenAI}, modelID: "my-custom-proxy-model", expected: "my-custom-proxy-model"},
		{name: "custom model not forced onto anthropic", account: &Account{Platform: PlatformAnthropic}, modelID: "my-custom-proxy-model", expected: ""},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, ResolveHealthCheckModelID(tc.account, tc.modelID))
		})
	}
}
