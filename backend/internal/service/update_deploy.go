package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	deployStatusIdle      = "idle"
	deployStatusPending   = "pending"
	deployStatusRunning   = "running"
	deployStatusSucceeded = "succeeded"
	deployStatusFailed    = "failed"

	deployExecutionModeInProcess = "in_process"
	deployExecutionModeHostAgent = "host_agent"

	deploySourceTypeGitSync = "git_sync"
)

type DeployConfig struct {
	Enabled           bool   `json:"enabled"`
	Mode              string `json:"mode"`
	ExecutionMode     string `json:"execution_mode,omitempty"`
	SourceType        string `json:"source_type,omitempty"`
	DefaultImage      string `json:"default_image"`
	RepoURL           string `json:"repo_url,omitempty"`
	Branch            string `json:"branch,omitempty"`
	RepoDir           string `json:"repo_dir,omitempty"`
	ServiceName       string `json:"service_name"`
	ComposeProjectDir string `json:"compose_project_dir"`
	ComposeFile       string `json:"compose_file,omitempty"`
	DockerBinary      string `json:"docker_binary,omitempty"`
	ComposeBinary     string `json:"compose_binary,omitempty"`
	AgentURL          string `json:"agent_url,omitempty"`
	AgentToken        string `json:"agent_token,omitempty"`
	AgentTimeoutSecs  int    `json:"agent_timeout_seconds,omitempty"`
	AgentInsecureTLS  bool   `json:"agent_insecure_tls,omitempty"`
}

type DeployState struct {
	Status         string `json:"status"`
	RequestedImage string `json:"requested_image,omitempty"`
	LastMessage    string `json:"last_message,omitempty"`
	LastError      string `json:"last_error,omitempty"`
	StartedAt      *int64 `json:"started_at,omitempty"`
	FinishedAt     *int64 `json:"finished_at,omitempty"`
}

type DeployTriggerRequest struct {
	Image  string `json:"image,omitempty"`
	DryRun bool   `json:"dry_run,omitempty"`
}

type DeployResult struct {
	Status            string   `json:"status"`
	Image             string   `json:"image"`
	ServiceName       string   `json:"service_name"`
	ComposeProjectDir string   `json:"compose_project_dir"`
	Commands          []string `json:"commands,omitempty"`
	Message           string   `json:"message"`
	NeedRestart       bool     `json:"need_restart"`
}

type deployAgentRequest struct {
	SourceType        string   `json:"source_type"`
	DefaultImage      string   `json:"default_image"`
	RepoURL           string   `json:"repo_url"`
	Branch            string   `json:"branch"`
	RepoDir           string   `json:"repo_dir"`
	ServiceName       string   `json:"service_name"`
	ComposeProjectDir string   `json:"compose_project_dir"`
	ComposeFile       string   `json:"compose_file,omitempty"`
	DockerBinary      string   `json:"docker_binary,omitempty"`
	ComposeBinary     string   `json:"compose_binary,omitempty"`
	Commands          []string `json:"commands,omitempty"`
}

type deployAgentResponse struct {
	Status            string   `json:"status"`
	Image             string   `json:"image,omitempty"`
	ServiceName       string   `json:"service_name,omitempty"`
	ComposeProjectDir string   `json:"compose_project_dir,omitempty"`
	Message           string   `json:"message,omitempty"`
	NeedRestart       bool     `json:"need_restart,omitempty"`
	Commands          []string `json:"commands,omitempty"`
	Output            string   `json:"output,omitempty"`
	Error             string   `json:"error,omitempty"`
}

type DeployCommandRunner interface {
	Run(ctx context.Context, dir string, name string, args ...string) (string, error)
}

type execDeployCommandRunner struct{}

func (execDeployCommandRunner) Run(ctx context.Context, dir string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func defaultDeployConfig() *DeployConfig {
	return &DeployConfig{
		Enabled:           false,
		Mode:              "docker_compose",
		ExecutionMode:     deployExecutionModeHostAgent,
		SourceType:        deploySourceTypeGitSync,
		DefaultImage:      "weishaw/sub2api:latest",
		RepoURL:           "https://github.com/YuHaiA/sub2api.git",
		Branch:            "main",
		RepoDir:           "/home/ec2-user/sub2api-source",
		ServiceName:       "sub2api",
		ComposeProjectDir: "/home/ec2-user/sub2api-deploy",
		DockerBinary:      "docker",
		ComposeBinary:     "docker-compose",
		AgentURL:          "http://172.17.0.1:18080",
		AgentTimeoutSecs:  900,
	}
}

func normalizeDeployConfig(cfg *DeployConfig) *DeployConfig {
	if cfg == nil {
		return defaultDeployConfig()
	}
	out := *cfg
	if strings.TrimSpace(out.Mode) == "" {
		out.Mode = "docker_compose"
	}
	if strings.TrimSpace(out.ExecutionMode) == "" {
		out.ExecutionMode = deployExecutionModeHostAgent
	}
	switch strings.TrimSpace(strings.ToLower(out.SourceType)) {
	case "", "docker_archive_url", "docker_image":
		out.SourceType = deploySourceTypeGitSync
	default:
		out.SourceType = strings.TrimSpace(strings.ToLower(out.SourceType))
	}
	if strings.TrimSpace(out.DefaultImage) == "" {
		out.DefaultImage = "weishaw/sub2api:latest"
	}
	if strings.TrimSpace(out.RepoURL) == "" {
		out.RepoURL = "https://github.com/YuHaiA/sub2api.git"
	}
	if strings.TrimSpace(out.Branch) == "" {
		out.Branch = "main"
	}
	if strings.TrimSpace(out.RepoDir) == "" {
		out.RepoDir = "/home/ec2-user/sub2api-source"
	}
	if strings.TrimSpace(out.ServiceName) == "" {
		out.ServiceName = "sub2api"
	}
	if strings.TrimSpace(out.ComposeProjectDir) == "" {
		out.ComposeProjectDir = "/home/ec2-user/sub2api-deploy"
	}
	if strings.TrimSpace(out.DockerBinary) == "" {
		out.DockerBinary = "docker"
	}
	if strings.TrimSpace(out.ComposeBinary) == "" {
		out.ComposeBinary = "docker-compose"
	}
	if strings.TrimSpace(out.AgentURL) == "" {
		out.AgentURL = "http://172.17.0.1:18080"
	}
	if out.AgentTimeoutSecs <= 0 {
		out.AgentTimeoutSecs = 900
	}
	out.Mode = strings.TrimSpace(strings.ToLower(out.Mode))
	out.ExecutionMode = strings.TrimSpace(strings.ToLower(out.ExecutionMode))
	out.DefaultImage = strings.TrimSpace(out.DefaultImage)
	out.RepoURL = strings.TrimSpace(out.RepoURL)
	out.Branch = strings.TrimSpace(out.Branch)
	out.RepoDir = strings.TrimSpace(out.RepoDir)
	out.ServiceName = strings.TrimSpace(out.ServiceName)
	out.ComposeProjectDir = strings.TrimSpace(out.ComposeProjectDir)
	out.ComposeFile = strings.TrimSpace(out.ComposeFile)
	out.DockerBinary = strings.TrimSpace(out.DockerBinary)
	out.ComposeBinary = strings.TrimSpace(out.ComposeBinary)
	out.AgentURL = normalizeDeployAgentBaseURL(strings.TrimSpace(out.AgentURL))
	out.AgentToken = strings.TrimSpace(out.AgentToken)
	return &out
}

func validateDeployConfig(cfg *DeployConfig) error {
	if cfg == nil {
		return nil
	}
	if cfg.Mode != "docker_compose" {
		return fmt.Errorf("unsupported mode: %s", cfg.Mode)
	}
	if cfg.ExecutionMode != deployExecutionModeInProcess && cfg.ExecutionMode != deployExecutionModeHostAgent {
		return fmt.Errorf("unsupported execution_mode: %s", cfg.ExecutionMode)
	}
	if cfg.SourceType != deploySourceTypeGitSync {
		return fmt.Errorf("unsupported source_type: %s", cfg.SourceType)
	}
	if !cfg.Enabled {
		return nil
	}
	if cfg.ExecutionMode != deployExecutionModeHostAgent {
		return fmt.Errorf("git_sync deployment requires host_agent execution mode")
	}
	if cfg.AgentURL == "" {
		return fmt.Errorf("agent_url is required when deployment is enabled")
	}
	if err := validateDeployAgentURL(cfg.AgentURL); err != nil {
		return err
	}
	if cfg.DefaultImage == "" {
		return fmt.Errorf("default_image is required when deployment is enabled")
	}
	if cfg.RepoURL == "" {
		return fmt.Errorf("repo_url is required when deployment is enabled")
	}
	if parsed, err := neturl.Parse(cfg.RepoURL); err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || strings.TrimSpace(parsed.Host) == "" {
		return fmt.Errorf("repo_url must be a valid http or https URL")
	}
	if cfg.Branch == "" {
		return fmt.Errorf("branch is required when deployment is enabled")
	}
	if cfg.RepoDir == "" {
		return fmt.Errorf("repo_dir is required when deployment is enabled")
	}
	if !filepath.IsAbs(cfg.RepoDir) {
		return fmt.Errorf("repo_dir must be an absolute path")
	}
	if cfg.ServiceName == "" {
		return fmt.Errorf("service_name is required when deployment is enabled")
	}
	if cfg.ComposeProjectDir == "" {
		return fmt.Errorf("compose_project_dir is required when deployment is enabled")
	}
	if !filepath.IsAbs(cfg.ComposeProjectDir) {
		return fmt.Errorf("compose_project_dir must be an absolute path")
	}
	if cfg.ComposeFile != "" && !filepath.IsAbs(cfg.ComposeFile) {
		return fmt.Errorf("compose_file must be an absolute path")
	}
	return nil
}

func defaultDeployState() *DeployState {
	return &DeployState{Status: deployStatusIdle}
}

func parseDeployConfig(raw string) *DeployConfig {
	cfg := defaultDeployConfig()
	if strings.TrimSpace(raw) == "" {
		return cfg
	}
	if err := json.Unmarshal([]byte(raw), cfg); err != nil {
		return defaultDeployConfig()
	}
	return normalizeDeployConfig(cfg)
}

func parseDeployState(raw string) *DeployState {
	state := defaultDeployState()
	if strings.TrimSpace(raw) == "" {
		return state
	}
	if err := json.Unmarshal([]byte(raw), state); err != nil {
		return defaultDeployState()
	}
	if state.Status == "" {
		state.Status = deployStatusIdle
	}
	return state
}

func (s *UpdateService) GetDeployConfig(ctx context.Context) (*DeployConfig, error) {
	if s.settingRepo == nil {
		return defaultDeployConfig(), nil
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeySystemDeployConfig)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return defaultDeployConfig(), nil
		}
		return defaultDeployConfig(), err
	}
	return parseDeployConfig(raw), nil
}

func (s *UpdateService) SaveDeployConfig(ctx context.Context, cfg *DeployConfig) error {
	if s.settingRepo == nil {
		return infraerrors.InternalServer("DEPLOY_CONFIG_STORE_UNAVAILABLE", "deployment configuration store unavailable")
	}
	cfg = normalizeDeployConfig(cfg)
	if err := validateDeployConfig(cfg); err != nil {
		return infraerrors.BadRequest("INVALID_DEPLOY_CONFIG", err.Error())
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal deploy config: %w", err)
	}
	return s.settingRepo.Set(ctx, SettingKeySystemDeployConfig, string(data))
}

func (s *UpdateService) GetDeployState(ctx context.Context) (*DeployState, error) {
	if s.settingRepo == nil {
		return defaultDeployState(), nil
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeySystemDeployState)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return defaultDeployState(), nil
		}
		return defaultDeployState(), err
	}
	return parseDeployState(raw), nil
}

func (s *UpdateService) saveDeployState(ctx context.Context, state *DeployState) error {
	if s.settingRepo == nil {
		return nil
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return s.settingRepo.Set(ctx, SettingKeySystemDeployState, string(data))
}

func (s *UpdateService) TriggerDeploy(ctx context.Context, req *DeployTriggerRequest) (*DeployResult, error) {
	cfg, err := s.GetDeployConfig(ctx)
	if err != nil {
		return nil, err
	}
	if cfg == nil || !cfg.Enabled {
		return nil, infraerrors.BadRequest("DEPLOY_DISABLED", "deployment mode is not enabled")
	}

	commands := s.buildDeployCommands(cfg)
	result := &DeployResult{
		Status:            deployStatusPending,
		Image:             cfg.DefaultImage,
		ServiceName:       cfg.ServiceName,
		ComposeProjectDir: cfg.ComposeProjectDir,
		Commands:          commands,
		Message:           "Deploy request recorded",
		NeedRestart:       false,
	}

	now := time.Now().Unix()
	state := &DeployState{
		Status:         deployStatusPending,
		RequestedImage: cfg.DefaultImage,
		LastMessage:    result.Message,
		StartedAt:      &now,
	}
	_ = s.saveDeployState(ctx, state)

	if req != nil && req.DryRun {
		result.Message = "Dry-run only; no deploy command executed"
		return result, nil
	}

	output, err := s.executeDeployCommands(ctx, cfg)
	finishedAt := time.Now().Unix()
	state.FinishedAt = &finishedAt
	if err != nil {
		state.Status = deployStatusFailed
		state.LastError = err.Error()
		if output != "" {
			state.LastMessage = output
		}
		_ = s.saveDeployState(context.Background(), state)
		return nil, infraerrors.InternalServer("DEPLOY_EXECUTION_FAILED", strings.TrimSpace(strings.Join([]string{err.Error(), output}, "\n")))
	}

	successMessage := "Deploy completed successfully"
	if strings.TrimSpace(output) != "" {
		successMessage = strings.TrimSpace(output)
	}
	state.Status = deployStatusSucceeded
	state.LastMessage = successMessage
	state.LastError = ""
	_ = s.saveDeployState(context.Background(), state)

	result.Status = deployStatusSucceeded
	result.Message = successMessage
	return result, nil
}

func (s *UpdateService) buildDeployCommands(cfg *DeployConfig) []string {
	commands := []string{}
	if cfg.ExecutionMode == deployExecutionModeHostAgent {
		commands = append(commands, fmt.Sprintf("POST %s/deploy", cfg.AgentURL))
	}
	commands = append(commands, buildDeployCommandsForExecution(cfg)...)
	return commands
}

func buildDeployCommandsForExecution(cfg *DeployConfig) []string {
	backupTag := fmt.Sprintf("%s:backup-<timestamp>", strings.Split(cfg.DefaultImage, ":")[0])
	commands := []string{
		fmt.Sprintf("if [ ! -d %s/.git ]; then git clone %s %s; fi", cfg.RepoDir, cfg.RepoURL, cfg.RepoDir),
		fmt.Sprintf("cd %s && git fetch origin", cfg.RepoDir),
		fmt.Sprintf("cd %s && git checkout %s", cfg.RepoDir, cfg.Branch),
		fmt.Sprintf("cd %s && git pull --ff-only origin %s", cfg.RepoDir, cfg.Branch),
		fmt.Sprintf("%s image inspect %s && %s tag %s %s", cfg.DockerBinary, cfg.DefaultImage, cfg.DockerBinary, cfg.DefaultImage, backupTag),
		fmt.Sprintf("cd %s && %s build -t %s .", cfg.RepoDir, cfg.DockerBinary, cfg.DefaultImage),
	}
	composeTarget := fmt.Sprintf("cd %s && %s up -d --no-deps %s", cfg.ComposeProjectDir, cfg.ComposeBinary, cfg.ServiceName)
	if cfg.ComposeFile != "" {
		composeTarget = fmt.Sprintf("cd %s && %s -f %s up -d --no-deps %s", cfg.ComposeProjectDir, cfg.ComposeBinary, cfg.ComposeFile, cfg.ServiceName)
	}
	commands = append(commands, composeTarget)
	return commands
}

func (s *UpdateService) executeDeployCommands(ctx context.Context, cfg *DeployConfig) (string, error) {
	if cfg.ExecutionMode != deployExecutionModeHostAgent {
		return "", fmt.Errorf("git_sync deployment requires host_agent execution mode")
	}
	return s.executeDeployViaAgent(ctx, cfg)
}

func normalizeDeployAgentBaseURL(raw string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(raw), "/")
	trimmed = strings.TrimSuffix(trimmed, "/deploy")
	trimmed = strings.TrimSuffix(trimmed, "/health")
	return strings.TrimRight(trimmed, "/")
}

func validateDeployAgentURL(raw string) error {
	parsed, err := neturl.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid agent_url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("agent_url must use http or https")
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return fmt.Errorf("agent_url host is required")
	}
	return nil
}

func (s *UpdateService) executeDeployViaAgent(ctx context.Context, cfg *DeployConfig) (string, error) {
	if err := validateDeployAgentURL(cfg.AgentURL); err != nil {
		return "", err
	}

	agentReq := deployAgentRequest{
		SourceType:        cfg.SourceType,
		DefaultImage:      cfg.DefaultImage,
		RepoURL:           cfg.RepoURL,
		Branch:            cfg.Branch,
		RepoDir:           cfg.RepoDir,
		ServiceName:       cfg.ServiceName,
		ComposeProjectDir: cfg.ComposeProjectDir,
		ComposeFile:       cfg.ComposeFile,
		DockerBinary:      cfg.DockerBinary,
		ComposeBinary:     cfg.ComposeBinary,
		Commands:          buildDeployCommandsForExecution(cfg),
	}

	payload, err := json.Marshal(agentReq)
	if err != nil {
		return "", fmt.Errorf("marshal deploy agent request: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.AgentTimeoutSecs)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, cfg.AgentURL+"/deploy", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("build deploy agent request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.AgentToken != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.AgentToken)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.AgentInsecureTLS}, //nolint:gosec // Admin-controlled optional self-signed agent TLS.
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request deploy agent: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if readErr != nil {
		return "", fmt.Errorf("read deploy agent response: %w", readErr)
	}

	var agentResp deployAgentResponse
	_ = json.Unmarshal(body, &agentResp)

	messageParts := []string{}
	if msg := strings.TrimSpace(agentResp.Message); msg != "" {
		messageParts = append(messageParts, msg)
	}
	if out := strings.TrimSpace(agentResp.Output); out != "" {
		messageParts = append(messageParts, out)
	}
	message := strings.TrimSpace(strings.Join(messageParts, "\n"))
	if message == "" {
		message = strings.TrimSpace(string(body))
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		errMsg := strings.TrimSpace(agentResp.Error)
		if errMsg != "" {
			if message != "" {
				return "", fmt.Errorf("%s\n%s", errMsg, message)
			}
			return "", fmt.Errorf("%s", errMsg)
		}
		if message == "" {
			message = fmt.Sprintf("deploy agent returned HTTP %d", resp.StatusCode)
		}
		return "", fmt.Errorf(message)
	}

	if message == "" {
		message = "Deploy completed successfully"
	}
	return message, nil
}
