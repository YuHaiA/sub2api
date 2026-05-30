package admin

import (
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// GetAccountHealthAutoCheckConfig returns the scheduled account health check settings.
func (h *SettingHandler) GetAccountHealthAutoCheckConfig(c *gin.Context) {
	cfg, err := h.settingService.GetAccountHealthAutoCheckConfig(c.Request.Context())
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, cfg)
}

// UpdateAccountHealthAutoCheckConfig saves the scheduled account health check settings.
func (h *SettingHandler) UpdateAccountHealthAutoCheckConfig(c *gin.Context) {
	var cfg service.AccountHealthAutoCheckConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.settingService.SaveAccountHealthAutoCheckConfig(c.Request.Context(), &cfg); response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, cfg)
}

// GetAccountTokenAutoRefreshConfig returns the scheduled account token refresh settings.
func (h *SettingHandler) GetAccountTokenAutoRefreshConfig(c *gin.Context) {
	cfg, err := h.settingService.GetAccountTokenAutoRefreshConfig(c.Request.Context())
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, cfg)
}

// UpdateAccountTokenAutoRefreshConfig saves the scheduled account token refresh settings.
func (h *SettingHandler) UpdateAccountTokenAutoRefreshConfig(c *gin.Context) {
	var cfg service.AccountTokenAutoRefreshConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.settingService.SaveAccountTokenAutoRefreshConfig(c.Request.Context(), &cfg); response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, cfg)
}

// RunAccountTokenAutoRefreshNow starts a manual account token refresh batch.
func (h *SettingHandler) RunAccountTokenAutoRefreshNow(c *gin.Context) {
	if h.tokenRefreshService == nil {
		response.Error(c, http.StatusServiceUnavailable, "token refresh service is unavailable")
		return
	}
	result, err := h.tokenRefreshService.RunManualBatchRefresh(c.Request.Context())
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, result)
}
