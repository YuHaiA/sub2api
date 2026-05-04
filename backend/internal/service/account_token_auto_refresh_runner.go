package service

import (
	"context"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"
)

const tokenAutoRefreshBatchPause = 2 * time.Second

type accountTokenAutoRefreshRunStats struct {
	Total   int
	Success int
	Failed  int
}

func (s *TokenRefreshService) SetSettingService(settingService *SettingService) {
	s.settingService = settingService
}

func (s *TokenRefreshService) RunConfiguredBatchRefresh(ctx context.Context, now time.Time) {
	if s == nil || s.settingService == nil || s.accountRepo == nil {
		return
	}
	cfg, err := s.settingService.GetAccountTokenAutoRefreshConfig(ctx)
	if err != nil || cfg == nil || !cfg.Enabled {
		return
	}
	if cfg.LastRunAt != nil {
		lastRunAt := time.Unix(*cfg.LastRunAt, 0)
		if now.Sub(lastRunAt) < accountTokenAutoRefreshDuration(cfg) {
			return
		}
	}

	stats, runErr := s.runConfiguredBatchRefresh(ctx, cfg)
	if runErr != nil {
		slog.Warn("token_refresh.auto_batch_failed", "error", runErr)
	}
	if err := s.settingService.MarkAccountTokenAutoRefreshRun(ctx, now, stats.Total, stats.Success, stats.Failed); err != nil {
		slog.Warn("token_refresh.auto_batch_mark_failed", "error", err)
	}
}

func (s *TokenRefreshService) runConfiguredBatchRefresh(
	ctx context.Context,
	cfg *AccountTokenAutoRefreshConfig,
) (accountTokenAutoRefreshRunStats, error) {
	eligible, err := s.listAutoRefreshEligibleAccounts(ctx)
	if err != nil {
		return accountTokenAutoRefreshRunStats{}, err
	}
	stats := accountTokenAutoRefreshRunStats{Total: len(eligible)}
	if len(eligible) == 0 {
		return stats, nil
	}

	batchSize := normalizeAccountTokenAutoRefreshConfig(cfg).BatchSize
	for start := 0; start < len(eligible); start += batchSize {
		end := start + batchSize
		if end > len(eligible) {
			end = len(eligible)
		}
		batchStats := s.runAutoRefreshBatch(ctx, eligible[start:end])
		stats.Success += batchStats.Success
		stats.Failed += batchStats.Failed

		if end < len(eligible) {
			select {
			case <-ctx.Done():
				return stats, ctx.Err()
			case <-time.After(tokenAutoRefreshBatchPause):
			}
		}
	}
	return stats, nil
}

func (s *TokenRefreshService) listAutoRefreshEligibleAccounts(ctx context.Context) ([]Account, error) {
	accounts, err := s.listActiveAccounts(ctx)
	if err != nil {
		return nil, err
	}
	filtered := make([]Account, 0, len(accounts))
	for i := range accounts {
		if _, _, ok := s.findRefreshExecutor(&accounts[i]); ok {
			filtered = append(filtered, accounts[i])
		}
	}
	return filtered, nil
}

func (s *TokenRefreshService) runAutoRefreshBatch(ctx context.Context, accounts []Account) accountTokenAutoRefreshRunStats {
	stats := accountTokenAutoRefreshRunStats{Total: len(accounts)}
	if len(accounts) == 0 {
		return stats
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(len(accounts))
	results := make(chan bool, len(accounts))

	for i := range accounts {
		account := accounts[i]
		g.Go(func() error {
			refresher, executor, ok := s.findRefreshExecutor(&account)
			if !ok {
				results <- false
				return nil
			}
			if err := s.refreshNowWithRetry(gctx, &account, refresher, executor); err != nil {
				slog.Warn("token_refresh.auto_batch_account_failed", "account_id", account.ID, "error", err)
				results <- false
				return nil
			}
			results <- true
			return nil
		})
	}

	_ = g.Wait()
	close(results)
	for ok := range results {
		if ok {
			stats.Success++
		} else {
			stats.Failed++
		}
	}
	return stats
}

func (s *TokenRefreshService) findRefreshExecutor(account *Account) (TokenRefresher, OAuthRefreshExecutor, bool) {
	for idx, refresher := range s.refreshers {
		if !refresher.CanRefresh(account) {
			continue
		}
		if idx < len(s.executors) {
			return refresher, s.executors[idx], true
		}
		return refresher, nil, true
	}
	return nil, nil, false
}

func (s *TokenRefreshService) refreshNowWithRetry(
	ctx context.Context,
	account *Account,
	refresher TokenRefresher,
	executor OAuthRefreshExecutor,
) error {
	var lastErr error

	for attempt := 1; attempt <= s.cfg.MaxRetries; attempt++ {
		var err error
		if s.refreshAPI != nil && executor != nil {
			result, refreshErr := s.refreshAPI.RefreshNow(ctx, account, executor)
			if refreshErr != nil {
				err = refreshErr
			} else if result.LockHeld {
				return s.refreshPolicy.handleLockHeld()
			} else {
				account = result.Account
			}
		} else {
			newCredentials, refreshErr := refresher.Refresh(ctx, account)
			if refreshErr != nil {
				err = refreshErr
			} else if newCredentials != nil {
				newCredentials["_token_version"] = time.Now().UnixMilli()
				if saveErr := persistAccountCredentials(ctx, s.accountRepo, account, newCredentials); saveErr != nil {
					return saveErr
				}
			}
		}

		if err == nil {
			s.postRefreshActions(ctx, account)
			return nil
		}
		if isNonRetryableRefreshError(err) {
			errorMsg := "Token refresh failed (non-retryable): " + err.Error()
			if setErr := s.accountRepo.SetError(ctx, account.ID, errorMsg); setErr != nil {
				slog.Warn("token_refresh.auto_batch_set_error_failed", "account_id", account.ID, "error", setErr)
			}
			return err
		}
		lastErr = err
		if attempt < s.cfg.MaxRetries {
			backoff := time.Duration(s.cfg.RetryBackoffSeconds) * time.Second * time.Duration(1<<(attempt-1))
			time.Sleep(backoff)
		}
	}
	until := time.Now().Add(tokenRefreshTempUnschedDuration)
	reason := "token refresh retry exhausted: " + lastErr.Error()
	if setErr := s.accountRepo.SetTempUnschedulable(ctx, account.ID, until, reason); setErr != nil {
		slog.Warn("token_refresh.auto_batch_set_temp_unsched_failed", "account_id", account.ID, "error", setErr)
	}
	return lastErr
}
