//go:build unit

package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTriggerDeployRejectsRunningImageMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status":"succeeded",
			"image_id":"sha256:target",
			"running_image_id":"sha256:old",
			"message":"Deploy completed successfully",
			"output":"compose completed"
		}`))
	}))
	defer server.Close()

	repo := newMockSettingRepo()
	config := defaultDeployConfig()
	config.Enabled = true
	config.AgentURL = server.URL
	rawConfig, err := json.Marshal(config)
	require.NoError(t, err)
	require.NoError(t, repo.Set(context.Background(), SettingKeySystemDeployConfig, string(rawConfig)))

	service := NewUpdateService(nil, nil, repo, "0.1.0", "release")
	result, err := service.TriggerDeploy(context.Background(), &DeployTriggerRequest{})
	require.Nil(t, result)
	require.Error(t, err)
	require.Contains(t, err.Error(), "deployment verification failed")

	rawState, err := repo.GetValue(context.Background(), SettingKeySystemDeployState)
	require.NoError(t, err)
	var state DeployState
	require.NoError(t, json.Unmarshal([]byte(rawState), &state))
	require.Equal(t, deployStatusFailed, state.Status)
	require.Equal(t, "sha256:target", state.RequestedImageID)
	require.Equal(t, "sha256:old", state.RunningImageID)
	require.Contains(t, state.LastError, "deployment verification failed")
}

func TestBuildComposeCommandPreviewForcesRecreate(t *testing.T) {
	command := buildComposeCommandPreview(&DeployConfig{
		ComposeProjectDir: "/srv/sub2api",
		ComposeBinary:     "docker-compose",
		ServiceName:       "sub2api",
	})

	require.True(t, strings.Contains(command, "up -d --no-deps --force-recreate sub2api"))
}
