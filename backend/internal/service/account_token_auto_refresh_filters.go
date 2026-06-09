package service

import (
	"strings"
	"time"
)

const (
	accountTokenHealthStatusUnchecked   = "unchecked"
	accountTokenHealthStatusHealthy     = "healthy"
	accountTokenHealthStatusConstrained = "constrained"
	accountTokenHealthStatusUnavailable = "unavailable"
)

func normalizeAccountTokenAutoRefreshHealthStatus(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case accountTokenHealthStatusHealthy,
		accountTokenHealthStatusConstrained,
		accountTokenHealthStatusUnavailable,
		accountTokenHealthStatusUnchecked:
		return strings.TrimSpace(strings.ToLower(raw))
	default:
		return ""
	}
}

func filterAccountsByTokenRefreshHealthStatus(accounts []Account, healthStatus string) []Account {
	if normalizeAccountTokenAutoRefreshHealthStatus(healthStatus) == "" {
		return accounts
	}
	filtered := make([]Account, 0, len(accounts))
	for i := range accounts {
		if accountTokenRefreshStoredHealthStatus(&accounts[i]) == healthStatus {
			filtered = append(filtered, accounts[i])
		}
	}
	return filtered
}

func accountTokenRefreshStoredHealthStatus(account *Account) string {
	if account == nil {
		return accountTokenHealthStatusUnchecked
	}
	if runtimeStatus := currentTokenRefreshRuntimeHealthStatus(account); runtimeStatus != "" {
		return runtimeStatus
	}
	if account.Extra == nil {
		return accountTokenHealthStatusUnchecked
	}
	raw, ok := account.Extra["health_check"]
	if !ok {
		return accountTokenHealthStatusUnchecked
	}
	data, ok := raw.(map[string]any)
	if !ok {
		return accountTokenHealthStatusUnchecked
	}
	status, _ := data["status"].(string)
	message, _ := data["message"].(string)
	return normalizeAccountTokenRefreshStoredHealthStatus(status, message)
}

func normalizeAccountTokenRefreshStoredHealthStatus(status string, message string) string {
	switch normalized := strings.TrimSpace(strings.ToLower(status)); normalized {
	case "", accountTokenHealthStatusUnchecked:
		return accountTokenHealthStatusUnchecked
	case accountTokenHealthStatusHealthy:
		return accountTokenHealthStatusHealthy
	case accountTokenHealthStatusConstrained:
		return accountTokenHealthStatusConstrained
	case accountTokenHealthStatusUnavailable:
		return accountTokenHealthStatusUnavailable
	case "rate_limited":
		return accountTokenHealthStatusConstrained
	case "banned_or_exhausted":
		return classifyAccountTokenRefreshHealthFromMessage(message)
	default:
		return classifyAccountTokenRefreshHealthFromMessage(message)
	}
}

func classifyAccountTokenRefreshHealthFromMessage(message string) string {
	lower := strings.ToLower(strings.TrimSpace(message))
	switch {
	case lower == "":
		return accountTokenHealthStatusUnchecked
	case strings.Contains(lower, "too many requests"),
		strings.Contains(lower, "rate limit"),
		strings.Contains(lower, "rate_limit"),
		strings.Contains(lower, "rate-limited"),
		strings.Contains(lower, "retry after"),
		strings.Contains(lower, "api returned 429"),
		strings.Contains(lower, "(429)"),
		strings.Contains(lower, "quota exhausted"),
		strings.Contains(lower, "quota_exhausted"),
		strings.Contains(lower, "insufficient quota"),
		strings.Contains(lower, "insufficient balance"),
		strings.Contains(lower, "insufficient credit"),
		strings.Contains(lower, "credits exhausted"),
		strings.Contains(lower, "credit exhausted"),
		strings.Contains(lower, "resource_exhausted"),
		strings.Contains(lower, "payment required"),
		strings.Contains(lower, "api returned 402"),
		strings.Contains(lower, "(402)"):
		return accountTokenHealthStatusConstrained
	default:
		return accountTokenHealthStatusUnavailable
	}
}

func currentTokenRefreshRuntimeHealthStatus(account *Account) string {
	if account == nil {
		return ""
	}
	now := time.Now()
	if account.RateLimitResetAt != nil && now.Before(*account.RateLimitResetAt) {
		return accountTokenHealthStatusConstrained
	}
	if account.OverloadUntil != nil && now.Before(*account.OverloadUntil) {
		return accountTokenHealthStatusConstrained
	}
	if account.TempUnschedulableUntil != nil && now.Before(*account.TempUnschedulableUntil) {
		return accountTokenHealthStatusConstrained
	}
	return ""
}
