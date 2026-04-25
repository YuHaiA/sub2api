package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/robfig/cron/v3"
)

const scheduledTestDefaultMaxWorkers = 10

// ScheduledTestRunnerService periodically scans due test plans and executes them.
type ScheduledTestRunnerService struct {
	planRepo       ScheduledTestPlanRepository
	scheduledSvc   *ScheduledTestService
	accountTestSvc *AccountTestService
	rateLimitSvc   *RateLimitService
	settingService *SettingService
	accountRepo    AccountRepository
	cfg            *config.Config

	cron      *cron.Cron
	startOnce sync.Once
	stopOnce  sync.Once
}

// NewScheduledTestRunnerService creates a new runner.
func NewScheduledTestRunnerService(
	planRepo ScheduledTestPlanRepository,
	scheduledSvc *ScheduledTestService,
	accountTestSvc *AccountTestService,
	rateLimitSvc *RateLimitService,
	settingService *SettingService,
	accountRepo AccountRepository,
	cfg *config.Config,
) *ScheduledTestRunnerService {
	return &ScheduledTestRunnerService{
		planRepo:       planRepo,
		scheduledSvc:   scheduledSvc,
		accountTestSvc: accountTestSvc,
		rateLimitSvc:   rateLimitSvc,
		settingService: settingService,
		accountRepo:    accountRepo,
		cfg:            cfg,
	}
}

// Start begins the cron ticker (every minute).
func (s *ScheduledTestRunnerService) Start() {
	if s == nil {
		return
	}
	s.startOnce.Do(func() {
		loc := time.Local
		if s.cfg != nil {
			if parsed, err := time.LoadLocation(s.cfg.Timezone); err == nil && parsed != nil {
				loc = parsed
			}
		}

		c := cron.New(cron.WithParser(scheduledTestCronParser), cron.WithLocation(loc))
		_, err := c.AddFunc("* * * * *", func() { s.runScheduled() })
		if err != nil {
			logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] not started (invalid schedule): %v", err)
			return
		}
		s.cron = c
		s.cron.Start()
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] started (tick=every minute)")
	})
}

// Stop gracefully shuts down the cron scheduler.
func (s *ScheduledTestRunnerService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		if s.cron != nil {
			ctx := s.cron.Stop()
			select {
			case <-ctx.Done():
			case <-time.After(3 * time.Second):
				logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] cron stop timed out")
			}
		}
	})
}

func (s *ScheduledTestRunnerService) runScheduled() {
	// Delay 10s so execution lands at ~:10 of each minute instead of :00.
	time.Sleep(10 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	now := time.Now()
	s.runAutoAccountHealthCheck(ctx, now)
	plans, err := s.planRepo.ListDue(ctx, now)
	if err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] ListDue error: %v", err)
		return
	}
	if len(plans) == 0 {
		return
	}

	logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] found %d due plans", len(plans))

	sem := make(chan struct{}, scheduledTestDefaultMaxWorkers)
	var wg sync.WaitGroup

	for _, plan := range plans {
		sem <- struct{}{}
		wg.Add(1)
		go func(p *ScheduledTestPlan) {
			defer wg.Done()
			defer func() { <-sem }()
			s.runOnePlan(ctx, p)
		}(plan)
	}

	wg.Wait()
}

func (s *ScheduledTestRunnerService) runAutoAccountHealthCheck(ctx context.Context, now time.Time) {
	if s.settingService == nil || s.accountRepo == nil || s.accountTestSvc == nil {
		return
	}
	cfg, err := s.settingService.GetAccountHealthAutoCheckConfig(ctx)
	if err != nil || cfg == nil || !cfg.Enabled {
		return
	}
	interval := time.Duration(cfg.IntervalMinutes) * time.Minute
	if interval <= 0 {
		interval = time.Duration(DefaultAccountHealthIntervalMinutes) * time.Minute
	}
	if cfg.LastRunAt != nil {
		lastRunAt := time.Unix(*cfg.LastRunAt, 0)
		if now.Sub(lastRunAt) < interval {
			return
		}
	}

	accounts, err := s.listAllAccountsForAutoHealthCheck(ctx)
	if err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] auto health list accounts error: %v", err)
		return
	}
	if len(accounts) == 0 {
		_ = s.settingService.MarkAccountHealthAutoCheckRun(ctx, now)
		return
	}

	logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] auto health check started (accounts=%d interval=%dmin)", len(accounts), cfg.IntervalMinutes)
	sem := make(chan struct{}, scheduledTestDefaultMaxWorkers)
	var wg sync.WaitGroup
	for i := range accounts {
		account := accounts[i]
		sem <- struct{}{}
		wg.Add(1)
		go func(acc *Account) {
			defer wg.Done()
			defer func() { <-sem }()
			s.runOneAutoHealthCheck(ctx, acc, strings.TrimSpace(cfg.ModelID), now)
		}(&account)
	}
	wg.Wait()
	if err := s.settingService.MarkAccountHealthAutoCheckRun(ctx, now); err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] auto health mark last run error: %v", err)
	}
}

func (s *ScheduledTestRunnerService) listAllAccountsForAutoHealthCheck(ctx context.Context) ([]Account, error) {
	page := 1
	pageSize := 200
	out := make([]Account, 0)
	for {
		items, result, err := s.accountRepo.ListWithFilters(ctx, pagination.PaginationParams{
			Page:      page,
			PageSize:  pageSize,
			SortBy:    "name",
			SortOrder: "asc",
		}, "", "", "", "", 0, "")
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
		if result == nil || len(out) >= int(result.Total) || len(items) == 0 {
			break
		}
		page++
	}
	return out, nil
}

func (s *ScheduledTestRunnerService) runOneAutoHealthCheck(ctx context.Context, account *Account, modelID string, checkedAt time.Time) {
	if account == nil {
		return
	}
	result, err := s.accountTestSvc.RunTestBackground(ctx, account.ID, modelID)
	if err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] auto health account=%d test error: %v", account.ID, err)
	}
	snapshot := map[string]any{
		"status":          autoHealthStatusFromResult(result),
		"result_status":   "",
		"message":         "",
		"latency_ms":      int64(0),
		"last_checked_at": checkedAt.UTC().Format(time.RFC3339),
	}
	if result != nil {
		snapshot["result_status"] = result.Status
		snapshot["message"] = result.ErrorMessage
		snapshot["latency_ms"] = result.LatencyMs
	}
	extra := cloneAutoHealthExtra(account.Extra)
	extra["health_check"] = snapshot
	if err := s.accountRepo.UpdateExtra(ctx, account.ID, extra); err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] auto health account=%d persist error: %v", account.ID, err)
	}
	if result != nil && strings.EqualFold(result.Status, "success") && s.rateLimitSvc != nil {
		if _, recoverErr := s.rateLimitSvc.RecoverAccountAfterSuccessfulTest(ctx, account.ID); recoverErr != nil {
			logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] auto health account=%d recover error: %v", account.ID, recoverErr)
		}
	}
}

func cloneAutoHealthExtra(src map[string]any) map[string]any {
	if len(src) == 0 {
		return map[string]any{}
	}
	dst := make(map[string]any, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func autoHealthStatusFromResult(result *ScheduledTestResult) string {
	if result == nil {
		return "unchecked"
	}
	if strings.EqualFold(result.Status, "success") {
		return "healthy"
	}
	lower := strings.ToLower(strings.TrimSpace(result.ErrorMessage))
	switch {
	case lower == "":
		return "unavailable"
	case strings.Contains(lower, "too many requests"),
		strings.Contains(lower, "rate limit"),
		strings.Contains(lower, "rate_limit"),
		strings.Contains(lower, "rate-limited"),
		strings.Contains(lower, "retry after"),
		strings.Contains(lower, "api returned 429"),
		strings.Contains(lower, "(429)"):
		return "constrained"
	case strings.Contains(lower, "quota exhausted"),
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
		return "constrained"
	case strings.Contains(lower, "banned"),
		strings.Contains(lower, "suspend"),
		strings.Contains(lower, "violation"):
		return "unavailable"
	default:
		return "unavailable"
	}
}

func (s *ScheduledTestRunnerService) runOnePlan(ctx context.Context, plan *ScheduledTestPlan) {
	result, err := s.accountTestSvc.RunTestBackground(ctx, plan.AccountID, plan.ModelID)
	if err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d RunTestBackground error: %v", plan.ID, err)
		return
	}

	if err := s.scheduledSvc.SaveResult(ctx, plan.ID, plan.MaxResults, result); err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d SaveResult error: %v", plan.ID, err)
	}

	// Auto-recover account if test succeeded and auto_recover is enabled.
	if result.Status == "success" && plan.AutoRecover {
		s.tryRecoverAccount(ctx, plan.AccountID, plan.ID)
	}

	nextRun, err := computeNextRun(plan.CronExpression, time.Now())
	if err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d computeNextRun error: %v", plan.ID, err)
		return
	}

	if err := s.planRepo.UpdateAfterRun(ctx, plan.ID, time.Now(), nextRun); err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d UpdateAfterRun error: %v", plan.ID, err)
	}
}

// tryRecoverAccount attempts to recover an account from recoverable runtime state.
func (s *ScheduledTestRunnerService) tryRecoverAccount(ctx context.Context, accountID int64, planID int64) {
	if s.rateLimitSvc == nil {
		return
	}

	recovery, err := s.rateLimitSvc.RecoverAccountAfterSuccessfulTest(ctx, accountID)
	if err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d auto-recover failed: %v", planID, err)
		return
	}
	if recovery == nil {
		return
	}

	if recovery.ClearedError {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d auto-recover: account=%d recovered from error status", planID, accountID)
	}
	if recovery.ClearedRateLimit {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d auto-recover: account=%d cleared rate-limit/runtime state", planID, accountID)
	}
}
