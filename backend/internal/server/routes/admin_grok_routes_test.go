package routes

import (
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	adminhandler "github.com/Wei-Shaw/sub2api/internal/handler/admin"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAdminGrokSSOToOAuthRouteIsRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	admin := router.Group("/api/v1/admin")
	registerGrokOAuthRoutes(admin, &handler.Handlers{
		Admin: &handler.AdminHandlers{
			GrokOAuth: &adminhandler.GrokOAuthHandler{},
		},
	})

	for _, route := range router.Routes() {
		if route.Method == http.MethodPost && route.Path == "/api/v1/admin/grok/sso-to-oauth" {
			return
		}
	}

	require.Fail(t, "POST /api/v1/admin/grok/sso-to-oauth should be registered")
}
