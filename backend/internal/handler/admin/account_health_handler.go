package admin

import (
	"context"
	"errors"
	"io"
	"log"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

func (h *AccountHandler) Deduplicate(c *gin.Context) {
	var req DeduplicateAccountsRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	filters := normalizeAccountHealthFilters(req.Filters)
	accounts, err := h.listAccountsForHealthFilters(c.Request.Context(), filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	grouped := make(map[string][]service.Account)
	for _, account := range accounts {
		key := strings.ToLower(strings.TrimSpace(account.Platform)) + "|" +
			strings.ToLower(strings.TrimSpace(account.Type)) + "|" +
			strings.ToLower(strings.TrimSpace(account.Name))
		grouped[key] = append(grouped[key], account)
	}

	result := DeduplicateAccountsResult{}
	for _, group := range grouped {
		if len(group) <= 1 {
			continue
		}
		result.DuplicateGroups++
		sort.Slice(group, func(i, j int) bool {
			return group[i].ID < group[j].ID
		})

		result.KeptCount++
		for _, duplicate := range group[1:] {
			if err := h.adminService.DeleteAccount(c.Request.Context(), duplicate.ID); err != nil {
				response.ErrorFrom(c, err)
				return
			}
			result.DeletedCount++
		}
	}

	response.Success(c, result)
}

// DeleteUnhealthy handles deleting accounts matching selected account or health statuses.
// POST /api/v1/admin/accounts/delete-unhealthy
func (h *AccountHandler) DeleteUnhealthy(c *gin.Context) {
	var req DeleteUnhealthyAccountsRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	filters := normalizeAccountHealthFilters(req.Filters)
	accounts, err := h.listAccountsForHealthFilters(c.Request.Context(), filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	accountStatusSet := normalizeDeleteAccountStatuses(req.AccountStatuses)
	healthStatusSet := normalizeDeleteHealthStatuses(req.HealthStatuses, len(req.AccountStatuses) == 0 && len(req.HealthStatuses) == 0)
	result := DeleteUnhealthyAccountsResult{}
	for _, account := range accounts {
		if !shouldDeleteAccountBySelection(account, accountStatusSet, healthStatusSet) {
			continue
		}
		if err := h.adminService.DeleteAccount(c.Request.Context(), account.ID); err != nil {
			response.ErrorFrom(c, err)
			return
		}
		result.DeletedCount++
	}

	response.Success(c, result)
}

const (
	accountHealthStatusUnchecked         = "unchecked"
	accountHealthStatusHealthy           = "healthy"
	accountHealthStatusConstrained       = "constrained"
	accountHealthStatusUnavailable       = "unavailable"
	accountHealthCheckExtraKey           = "health_check"
	defaultAccountHealthCheckConcurrency = 2
	defaultAccountHealthCheckBatchSize   = 10
	accountHealthCheckBatchPause         = 2 * time.Second
	defaultAccountHealthFetchPageSize    = 200
)

type AccountHealthCheckFilters struct {
	Platform     string `json:"platform"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	HealthStatus string `json:"health_status"`
	Group        string `json:"group"`
	Search       string `json:"search"`
	PrivacyMode  string `json:"privacy_mode"`
	SortBy       string `json:"sort_by"`
	SortOrder    string `json:"sort_order"`
}

type AccountHealthCheckRequest struct {
	AccountIDs []int64                    `json:"account_ids"`
	Filters    *AccountHealthCheckFilters `json:"filters"`
	ModelID    string                     `json:"model_id"`
}

type DeduplicateAccountsRequest struct {
	Filters *AccountHealthCheckFilters `json:"filters"`
}

type DeduplicateAccountsResult struct {
	DuplicateGroups int `json:"duplicate_groups"`
	DeletedCount    int `json:"deleted_count"`
	KeptCount       int `json:"kept_count"`
}

type DeleteUnhealthyAccountsRequest struct {
	Filters         *AccountHealthCheckFilters `json:"filters"`
	AccountStatuses []string                   `json:"account_statuses"`
	HealthStatuses  []string                   `json:"health_statuses"`
}

type DeleteUnhealthyAccountsResult struct {
	DeletedCount int `json:"deleted_count"`
}

type AccountHealthAutoConfigRequest struct {
	Enabled         bool   `json:"enabled"`
	IntervalMinutes int    `json:"interval_minutes"`
	ModelID         string `json:"model_id"`
}

type AccountHealthSummary struct {
	TotalAccounts       int    `json:"total_accounts"`
	HealthyAccounts     int    `json:"healthy_accounts"`
	ConstrainedAccounts int    `json:"constrained_accounts"`
	UnavailableAccounts int    `json:"unavailable_accounts"`
	UncheckedAccounts   int    `json:"unchecked_accounts"`
	LastCheckedAt       string `json:"last_checked_at,omitempty"`
}

type AccountHealthCheckItem struct {
	AccountID     int64  `json:"account_id"`
	Name          string `json:"name"`
	Platform      string `json:"platform"`
	Type          string `json:"type"`
	HealthStatus  string `json:"health_status"`
	ResultStatus  string `json:"result_status"`
	Message       string `json:"message,omitempty"`
	LatencyMs     int64  `json:"latency_ms"`
	LastCheckedAt string `json:"last_checked_at"`
}

type AccountHealthCheckRunResult struct {
	Started bool   `json:"started"`
	Running bool   `json:"running"`
	Message string `json:"message"`
	RunAt   int64  `json:"run_at"`
	Total   int    `json:"total"`
}

var accountHealthManualRunInProgress atomic.Bool

type accountHealthSnapshot struct {
	Status        string
	ResultStatus  string
	Message       string
	LatencyMs     int64
	LastCheckedAt string
}

func normalizeAccountHealthFilters(filters *AccountHealthCheckFilters) AccountHealthCheckFilters {
	if filters == nil {
		return AccountHealthCheckFilters{
			SortBy:    "name",
			SortOrder: "asc",
		}
	}
	normalized := *filters
	normalized.Platform = strings.TrimSpace(normalized.Platform)
	normalized.Type = strings.TrimSpace(normalized.Type)
	normalized.Status = strings.TrimSpace(normalized.Status)
	normalized.HealthStatus = normalizeAccountHealthStatusFilter(normalized.HealthStatus)
	normalized.Group = strings.TrimSpace(normalized.Group)
	normalized.Search = strings.TrimSpace(normalized.Search)
	normalized.PrivacyMode = strings.TrimSpace(normalized.PrivacyMode)
	normalized.SortBy = strings.TrimSpace(normalized.SortBy)
	if normalized.SortBy == "" {
		normalized.SortBy = "name"
	}
	normalized.SortOrder = strings.TrimSpace(strings.ToLower(normalized.SortOrder))
	if normalized.SortOrder != "desc" {
		normalized.SortOrder = "asc"
	}
	return normalized
}

func normalizeDeleteAccountStatuses(statuses []string) map[string]struct{} {
	allowed := map[string]string{
		service.StatusActive:   service.StatusActive,
		service.StatusDisabled: service.StatusDisabled,
		"inactive":             service.StatusDisabled,
		service.StatusError:    service.StatusError,
		"rate_limited":         "rate_limited",
		"temp_unschedulable":   "temp_unschedulable",
		"unschedulable":        "unschedulable",
	}
	return normalizeDeleteStatusSet(statuses, allowed)
}

func normalizeDeleteHealthStatuses(statuses []string, useDefault bool) map[string]struct{} {
	allowed := map[string]string{
		accountHealthStatusUnchecked:   accountHealthStatusUnchecked,
		accountHealthStatusHealthy:     accountHealthStatusHealthy,
		accountHealthStatusConstrained: accountHealthStatusConstrained,
		accountHealthStatusUnavailable: accountHealthStatusUnavailable,
	}
	result := normalizeDeleteStatusSet(statuses, allowed)
	if useDefault {
		result[accountHealthStatusUnavailable] = struct{}{}
	}
	return result
}

func normalizeDeleteStatusSet(statuses []string, allowed map[string]string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, status := range statuses {
		normalized, ok := allowed[strings.TrimSpace(strings.ToLower(status))]
		if !ok {
			continue
		}
		result[normalized] = struct{}{}
	}
	return result
}

func shouldDeleteAccountBySelection(account service.Account, accountStatuses map[string]struct{}, healthStatuses map[string]struct{}) bool {
	for status := range accountStatuses {
		if accountMatchesDeleteAccountStatus(account, status) {
			return true
		}
	}
	if len(healthStatuses) == 0 {
		return false
	}
	_, ok := healthStatuses[parseStoredAccountHealth(&account).Status]
	return ok
}

func accountMatchesDeleteAccountStatus(account service.Account, status string) bool {
	now := time.Now()
	switch status {
	case "rate_limited":
		return account.Status == service.StatusActive && account.RateLimitResetAt != nil && account.RateLimitResetAt.After(now)
	case "temp_unschedulable":
		return account.Status == service.StatusActive && account.TempUnschedulableUntil != nil && account.TempUnschedulableUntil.After(now)
	case "unschedulable":
		return account.Status == service.StatusActive && !account.Schedulable && !accountRateLimited(account, now) && !accountTempUnschedulable(account, now)
	case service.StatusDisabled, "inactive":
		return account.Status == service.StatusDisabled || account.Status == "inactive"
	default:
		return account.Status == status
	}
}

func accountRateLimited(account service.Account, now time.Time) bool {
	return account.RateLimitResetAt != nil && account.RateLimitResetAt.After(now)
}

func accountTempUnschedulable(account service.Account, now time.Time) bool {
	return account.TempUnschedulableUntil != nil && account.TempUnschedulableUntil.After(now)
}

func parseAccountHealthGroupID(raw string) (int64, error) {
	if raw == "" {
		return 0, nil
	}
	if raw == accountListGroupUngroupedQueryValue {
		return service.AccountListGroupUngrouped, nil
	}
	groupID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || groupID < 0 {
		return 0, infraerrors.BadRequest("INVALID_GROUP_FILTER", "invalid group filter")
	}
	return groupID, nil
}

func (h *AccountHandler) listAccountsForHealthFilters(ctx context.Context, filters AccountHealthCheckFilters) ([]service.Account, error) {
	groupID, err := parseAccountHealthGroupID(filters.Group)
	if err != nil {
		return nil, err
	}

	page := 1
	accounts := make([]service.Account, 0)
	for {
		rows, total, err := h.adminService.ListAccounts(
			ctx,
			page,
			defaultAccountHealthFetchPageSize,
			filters.Platform,
			filters.Type,
			filters.Status,
			filters.Search,
			groupID,
			filters.PrivacyMode,
			filters.SortBy,
			filters.SortOrder,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, rows...)
		if int64(len(accounts)) >= total || len(rows) < defaultAccountHealthFetchPageSize {
			break
		}
		page++
	}
	accounts = filterAccountsByHealthStatus(accounts, filters.HealthStatus)
	return accounts, nil
}

func cloneMap(src map[string]any) map[string]any {
	if len(src) == 0 {
		return map[string]any{}
	}
	dst := make(map[string]any, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func classifyAccountHealthStatus(result *service.ScheduledTestResult) string {
	if result == nil {
		return accountHealthStatusUnchecked
	}
	if strings.EqualFold(result.Status, "success") {
		return accountHealthStatusHealthy
	}
	lower := strings.ToLower(strings.TrimSpace(result.ErrorMessage))
	switch {
	case lower == "":
		return accountHealthStatusUnavailable
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
		strings.Contains(lower, "(402)"),
		// Transient network/proxy pressure is not proof the account credential is dead.
		strings.Contains(lower, "eof"),
		strings.Contains(lower, "timeout"),
		strings.Contains(lower, "deadline exceeded"),
		strings.Contains(lower, "connection reset"),
		strings.Contains(lower, "proxy connection failed"),
		strings.Contains(lower, "chatgpt codex upstream connection failed"),
		strings.Contains(lower, "stream ended before response.completed"),
		strings.Contains(lower, "stream read error"):
		return accountHealthStatusConstrained
	case strings.Contains(lower, "banned"),
		strings.Contains(lower, "suspend"),
		strings.Contains(lower, "violation"):
		return accountHealthStatusUnavailable
	default:
		return accountHealthStatusUnavailable
	}
}

func normalizeStoredAccountHealthStatus(status string, message string) string {
	normalizedStatus := strings.TrimSpace(strings.ToLower(status))
	switch normalizedStatus {
	case "", accountHealthStatusUnchecked:
		return accountHealthStatusUnchecked
	case accountHealthStatusHealthy:
		return accountHealthStatusHealthy
	case accountHealthStatusConstrained:
		return accountHealthStatusConstrained
	case accountHealthStatusUnavailable:
		return accountHealthStatusUnavailable
	case "rate_limited":
		return accountHealthStatusConstrained
	case "banned_or_exhausted":
		return classifyAccountHealthStatus(&service.ScheduledTestResult{
			Status:       "failed",
			ErrorMessage: message,
		})
	default:
		return normalizedStatus
	}
}

func normalizeAccountHealthStatusFilter(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case accountHealthStatusUnchecked,
		accountHealthStatusHealthy,
		accountHealthStatusConstrained,
		accountHealthStatusUnavailable:
		return strings.TrimSpace(strings.ToLower(raw))
	default:
		return ""
	}
}

func filterAccountsByHealthStatus(accounts []service.Account, healthStatus string) []service.Account {
	if healthStatus == "" {
		return accounts
	}
	filtered := make([]service.Account, 0, len(accounts))
	for i := range accounts {
		if parseStoredAccountHealth(&accounts[i]).Status == healthStatus {
			filtered = append(filtered, accounts[i])
		}
	}
	return filtered
}

func buildAccountHealthSnapshot(result *service.ScheduledTestResult) accountHealthSnapshot {
	snapshot := accountHealthSnapshot{
		Status:        classifyAccountHealthStatus(result),
		LastCheckedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if result != nil {
		snapshot.ResultStatus = result.Status
		snapshot.Message = result.ErrorMessage
		snapshot.LatencyMs = result.LatencyMs
	}
	return snapshot
}

func parseStoredAccountHealth(account *service.Account) accountHealthSnapshot {
	if account == nil {
		return accountHealthSnapshot{Status: accountHealthStatusUnchecked}
	}
	if runtimeStatus := currentRuntimeAccountHealthStatus(account); runtimeStatus != "" {
		return accountHealthSnapshot{Status: runtimeStatus}
	}
	if account.Extra == nil {
		return accountHealthSnapshot{Status: accountHealthStatusUnchecked}
	}
	raw, ok := account.Extra[accountHealthCheckExtraKey]
	if !ok {
		return accountHealthSnapshot{Status: accountHealthStatusUnchecked}
	}
	data, ok := raw.(map[string]any)
	if !ok {
		return accountHealthSnapshot{Status: accountHealthStatusUnchecked}
	}

	status, _ := data["status"].(string)
	resultStatus, _ := data["result_status"].(string)
	message, _ := data["message"].(string)
	status = normalizeStoredAccountHealthStatus(status, message)
	lastCheckedAt, _ := data["last_checked_at"].(string)

	var latencyMs int64
	switch value := data["latency_ms"].(type) {
	case int64:
		latencyMs = value
	case int:
		latencyMs = int64(value)
	case float64:
		latencyMs = int64(value)
	}

	return accountHealthSnapshot{
		Status:        status,
		ResultStatus:  resultStatus,
		Message:       message,
		LatencyMs:     latencyMs,
		LastCheckedAt: lastCheckedAt,
	}
}

func currentRuntimeAccountHealthStatus(account *service.Account) string {
	if account == nil {
		return ""
	}
	now := time.Now()
	if account.RateLimitResetAt != nil && now.Before(*account.RateLimitResetAt) {
		return accountHealthStatusConstrained
	}
	if account.OverloadUntil != nil && now.Before(*account.OverloadUntil) {
		return accountHealthStatusConstrained
	}
	if account.TempUnschedulableUntil != nil && now.Before(*account.TempUnschedulableUntil) {
		return accountHealthStatusConstrained
	}
	return ""
}

func summarizeAccountHealthSnapshots(snapshots []accountHealthSnapshot) AccountHealthSummary {
	summary := AccountHealthSummary{}
	for _, snapshot := range snapshots {
		summary.TotalAccounts++
		switch snapshot.Status {
		case accountHealthStatusHealthy:
			summary.HealthyAccounts++
		case accountHealthStatusConstrained:
			summary.ConstrainedAccounts++
		case accountHealthStatusUnavailable:
			summary.UnavailableAccounts++
		default:
			summary.UncheckedAccounts++
		}
		if snapshot.LastCheckedAt > summary.LastCheckedAt {
			summary.LastCheckedAt = snapshot.LastCheckedAt
		}
	}
	return summary
}

func buildAccountHealthItem(account *service.Account, snapshot accountHealthSnapshot) AccountHealthCheckItem {
	item := AccountHealthCheckItem{
		HealthStatus:  accountHealthStatusUnchecked,
		LastCheckedAt: snapshot.LastCheckedAt,
		ResultStatus:  snapshot.ResultStatus,
		Message:       snapshot.Message,
		LatencyMs:     snapshot.LatencyMs,
	}
	if account != nil {
		item.AccountID = account.ID
		item.Name = account.Name
		item.Platform = account.Platform
		item.Type = account.Type
	}
	if snapshot.Status != "" {
		item.HealthStatus = snapshot.Status
	}
	return item
}

func (h *AccountHandler) persistAccountHealthSnapshot(ctx context.Context, account *service.Account, snapshot accountHealthSnapshot) error {
	if account == nil {
		return nil
	}
	extra := cloneMap(account.Extra)
	extra[accountHealthCheckExtraKey] = map[string]any{
		"status":          snapshot.Status,
		"result_status":   snapshot.ResultStatus,
		"message":         snapshot.Message,
		"latency_ms":      snapshot.LatencyMs,
		"last_checked_at": snapshot.LastCheckedAt,
	}
	updated, err := h.adminService.UpdateAccount(ctx, account.ID, &service.UpdateAccountInput{
		Extra: extra,
	})
	if err != nil {
		return err
	}
	if updated != nil {
		account.Extra = updated.Extra
	}
	return nil
}

// GetHealthSummary handles aggregate account health summary for the current filters.
// GET /api/v1/admin/accounts/health-summary
func (h *AccountHandler) GetHealthSummary(c *gin.Context) {
	filters := normalizeAccountHealthFilters(&AccountHealthCheckFilters{
		Platform:     strings.TrimSpace(c.Query("platform")),
		Type:         strings.TrimSpace(c.Query("type")),
		Status:       strings.TrimSpace(c.Query("status")),
		HealthStatus: strings.TrimSpace(c.Query("health_status")),
		Group:        strings.TrimSpace(c.Query("group")),
		Search:       strings.TrimSpace(c.Query("search")),
		PrivacyMode:  strings.TrimSpace(c.Query("privacy_mode")),
		SortBy:       strings.TrimSpace(c.Query("sort_by")),
		SortOrder:    strings.TrimSpace(c.Query("sort_order")),
	})

	accounts, err := h.listAccountsForHealthFilters(c.Request.Context(), filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	snapshots := make([]accountHealthSnapshot, 0, len(accounts))
	for i := range accounts {
		snapshots = append(snapshots, parseStoredAccountHealth(&accounts[i]))
	}

	response.Success(c, summarizeAccountHealthSnapshots(snapshots))
}

// RunHealthCheck handles batch account health checks for selected or filtered accounts.
// POST /api/v1/admin/accounts/health-check
func (h *AccountHandler) RunHealthCheck(c *gin.Context) {
	var req AccountHealthCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	ctx := c.Request.Context()
	modelID := strings.TrimSpace(req.ModelID)
	accounts := make([]*service.Account, 0)
	if len(req.AccountIDs) > 0 {
		normalizedIDs := normalizeInt64IDList(req.AccountIDs)
		fetched, err := h.adminService.GetAccountsByIDs(ctx, normalizedIDs)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		accounts = fetched
	} else {
		filters := normalizeAccountHealthFilters(req.Filters)
		rows, err := h.listAccountsForHealthFilters(ctx, filters)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		for i := range rows {
			account := rows[i]
			accounts = append(accounts, &account)
		}
	}

	if !accountHealthManualRunInProgress.CompareAndSwap(false, true) {
		response.Success(c, &AccountHealthCheckRunResult{
			Started: false,
			Running: true,
			Message: "Account health check is already running",
			RunAt:   time.Now().Unix(),
			Total:   len(accounts),
		})
		return
	}

	startedAt := time.Now()
	mode, relatedTask := service.EnqueueBackgroundMaintenance(service.BackgroundMaintenanceTask{
		Name: "account_health_manual",
		Run: func() {
			defer accountHealthManualRunInProgress.Store(false)

			runCtx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
			defer cancel()

			finished := false
			defer func() {
				if !finished {
					if err := h.markAccountHealthCheckFinished(time.Now()); err != nil {
						slog.Warn("account_health.manual_batch_mark_failed", "error", err)
					}
				}
			}()

			if len(accounts) == 0 {
				if err := h.markAccountHealthCheckFinished(startedAt); err != nil {
					slog.Warn("account_health.manual_batch_mark_failed", "error", err)
				}
				finished = true
				return
			}

			if h.settingService != nil {
				_ = h.settingService.MarkAccountHealthAutoCheckProgress(context.Background(), len(accounts), 0, 0)
			}
			var progressMu sync.Mutex
			success := 0
			failed := 0
			for start := 0; start < len(accounts); start += defaultAccountHealthCheckBatchSize {
				end := start + defaultAccountHealthCheckBatchSize
				if end > len(accounts) {
					end = len(accounts)
				}

				_, _, batchFailed, err := h.runAccountHealthCheckBatch(runCtx, accounts[start:end], modelID, func(accountFailed bool) {
					progressMu.Lock()
					if accountFailed {
						failed++
					} else {
						success++
					}
					currentSuccess := success
					currentFailed := failed
					progressMu.Unlock()
					if h.settingService != nil {
						_ = h.settingService.MarkAccountHealthAutoCheckProgress(context.Background(), len(accounts), currentSuccess, currentFailed)
					}
				})
				if err != nil {
					slog.Warn("account_health.manual_batch_failed", "error", err)
					return
				}
				_ = batchFailed

				if end < len(accounts) {
					select {
					case <-runCtx.Done():
						slog.Warn("account_health.manual_batch_interrupted", "error", runCtx.Err())
						return
					case <-time.After(accountHealthCheckBatchPause):
					}
				}
			}
			if err := h.markAccountHealthCheckFinished(startedAt); err != nil {
				slog.Warn("account_health.manual_batch_mark_failed", "error", err)
			}
			finished = true
		},
	})

	switch mode {
	case service.BackgroundMaintenanceRunNow:
		response.Success(c, &AccountHealthCheckRunResult{
			Started: true,
			Running: true,
			Message: "Account health check started in background",
			RunAt:   startedAt.Unix(),
			Total:   len(accounts),
		})
	case service.BackgroundMaintenanceQueued:
		response.Success(c, &AccountHealthCheckRunResult{
			Started: true,
			Running: true,
			Message: "Account health check queued behind: " + relatedTask,
			RunAt:   startedAt.Unix(),
			Total:   len(accounts),
		})
	default:
		accountHealthManualRunInProgress.Store(false)
		response.Success(c, &AccountHealthCheckRunResult{
			Started: false,
			Running: true,
			Message: "Account health check is already running or queued",
			RunAt:   startedAt.Unix(),
			Total:   len(accounts),
		})
	}
}

func (h *AccountHandler) runAccountHealthCheckBatch(
	ctx context.Context,
	accounts []*service.Account,
	modelID string,
	onAccountDone func(failed bool),
) ([]AccountHealthCheckItem, []accountHealthSnapshot, int, error) {
	items := make([]AccountHealthCheckItem, 0, len(accounts))
	snapshots := make([]accountHealthSnapshot, 0, len(accounts))
	if len(accounts) == 0 {
		return items, snapshots, 0, nil
	}

	var mu sync.Mutex
	failedCount := 0
	group, gctx := errgroup.WithContext(ctx)
	group.SetLimit(defaultAccountHealthCheckConcurrency)

	for _, account := range accounts {
		if account == nil {
			continue
		}
		accountRef := account
		group.Go(func() error {
			snapshot, failed := h.runSingleAccountHealthCheck(gctx, accountRef, modelID)
			mu.Lock()
			snapshots = append(snapshots, snapshot)
			items = append(items, buildAccountHealthItem(accountRef, snapshot))
			if failed {
				failedCount++
			}
			mu.Unlock()
			if onAccountDone != nil {
				onAccountDone(failed)
			}
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, nil, failedCount, err
	}
	return items, snapshots, failedCount, nil
}

func (h *AccountHandler) runSingleAccountHealthCheck(
	ctx context.Context,
	account *service.Account,
	modelID string,
) (accountHealthSnapshot, bool) {
	release, slotErr := service.AcquireBackgroundTaskSlot(ctx)
	if slotErr != nil {
		log.Printf("account health check slot failed for account=%d: %v", account.ID, slotErr)
		return accountHealthSnapshot{
			Status:        accountHealthStatusUnavailable,
			ResultStatus:  "failed",
			Message:       slotErr.Error(),
			LastCheckedAt: time.Now().UTC().Format(time.RFC3339),
		}, true
	}
	defer release()

	result, err := h.accountTestService.RunTestBackground(ctx, account.ID, service.ResolveHealthCheckModelID(account, modelID))
	failed := false
	if err != nil {
		log.Printf("account health check failed for account=%d: %v", account.ID, err)
		failed = true
	}

	snapshot := buildAccountHealthSnapshot(result)
	// Progress counters should reflect actual health outcomes, not only infra errors.
	if snapshot.Status != accountHealthStatusHealthy && snapshot.Status != accountHealthStatusConstrained {
		failed = true
	}
	if result != nil && strings.EqualFold(result.Status, "success") && h.rateLimitService != nil {
		if _, recoverErr := h.rateLimitService.RecoverAccountAfterSuccessfulTest(ctx, account.ID); recoverErr != nil {
			log.Printf("account health check recover failed for account=%d: %v", account.ID, recoverErr)
		}
	}
	if persistErr := h.persistAccountHealthSnapshot(ctx, account, snapshot); persistErr != nil {
		log.Printf("account health check persist failed for account=%d: %v", account.ID, persistErr)
		failed = true
	}
	return snapshot, failed
}

func (h *AccountHandler) markAccountHealthCheckFinished(runAt time.Time) error {
	if h.settingService == nil {
		return nil
	}
	return h.settingService.MarkAccountHealthAutoCheckRun(context.Background(), runAt)
}
