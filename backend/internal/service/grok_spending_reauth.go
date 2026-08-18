package service

import (
	"context"
	"strings"
	"time"
)

// Spending-limit is recoverable at the end of the observed billing period.
// When no billing snapshot is available, use a short probe rather than
// fabricating a 24h boundary from the error arrival time.
const grokSpendingLimitProbeCooldown = 10 * time.Minute

func grokSpendingLimitResetAt(account *Account, now time.Time) time.Time {
	if resetAt, ok := grokOfficialUsageResetAt(account, now); ok {
		return resetAt
	}
	return now.Add(grokSpendingLimitProbeCooldown)
}

func grokOfficialUsageResetAt(account *Account, now time.Time) (time.Time, bool) {
	if account == nil {
		return time.Time{}, false
	}
	billing, err := grokBillingSnapshotFromExtra(account.Extra)
	if err != nil || billing == nil {
		return time.Time{}, false
	}

	var candidates []string
	switch strings.ToLower(strings.TrimSpace(billing.PeriodType)) {
	case "weekly":
		candidates = []string{billing.PeriodEnd}
	case "monthly":
		candidates = []string{billing.BillingPeriodEnd, billing.PeriodEnd}
	default:
		candidates = []string{billing.PeriodEnd, billing.BillingPeriodEnd}
	}
	for _, raw := range candidates {
		if resetAt, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(raw)); parseErr == nil && resetAt.After(now) {
			return resetAt, true
		}
	}
	return time.Time{}, false
}

// clearGrokNeedsReauthExtra drops the soft reauth flag after successful refresh
// or reauth. Best-effort; never fails the request path.
func clearGrokNeedsReauthExtra(ctx context.Context, repo AccountRepository, accountID int64) {
	if repo == nil || accountID <= 0 {
		return
	}
	stateCtx, cancel := openAIAccountStateContext(ctx)
	defer cancel()
	_ = repo.UpdateExtra(stateCtx, accountID, map[string]any{
		"grok_needs_reauth":        false,
		"grok_needs_reauth_reason": "",
		"grok_needs_reauth_at":     "",
	})
}

func accountGrokNeedsReauth(account *Account) bool {
	if account == nil {
		return false
	}
	if account.Status == StatusError {
		msg := strings.ToLower(account.ErrorMessage)
		if strings.Contains(msg, "spending limit") || strings.Contains(msg, "reauthorize") {
			return true
		}
	}
	if v, ok := account.Extra["grok_needs_reauth"].(bool); ok && v {
		return true
	}
	if s, ok := account.Extra["grok_needs_reauth"].(string); ok {
		return strings.EqualFold(s, "true") || s == "1"
	}
	return false
}
