package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
)

type DeployConfig struct {
	Enabled            bool   `json:"enabled"`
	Mode               string `json:"mode"`
	SourceType         string `json:"source_type,omitempty"`
	DefaultImage       string `json:"default_image"`
	AllowedImagePrefix string `json:"allowed_image_prefix,omitempty"`
	ArchiveURL         string `json:"archive_url,omitempty"`
	LoadedImage        string `json:"loaded_image,omitempty"`
	ServiceName        string `json:"service_name"`
	ComposeProjectDir  string `json:"compose_project_dir"`
	ComposeFile        string `json:"compose_file,omitempty"`
	DockerBinary       string `json:"docker_binary,omitempty"`
	ComposeBinary      string `json:"compose_binary,omitempty"`
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
		SourceType:        "docker_archive_url",
		DefaultImage:      "weishaw/sub2api:latest",
		ArchiveURL:        "https://github.com/YuHaiA/sub2api/releases/download/docker-deploy/sub2api-docker-image.tar",
		LoadedImage:       "sub2api-gha:docker-deploy",
		ServiceName:       "sub2api",
		ComposeProjectDir: "/home/ec2-user/sub2api-deploy",
		DockerBinary:      "docker",
		ComposeBinary:     "docker-compose",
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
	if strings.TrimSpace(out.SourceType) == "" {
		out.SourceType = "docker_image"
	}
	if strings.TrimSpace(out.DefaultImage) == "" {
		out.DefaultImage = "weishaw/sub2api:latest"
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
	out.Mode = strings.TrimSpace(strings.ToLower(out.Mode))
	out.SourceType = strings.TrimSpace(strings.ToLower(out.SourceType))
	out.DefaultImage = strings.TrimSpace(out.DefaultImage)
	out.AllowedImagePrefix = strings.TrimSpace(out.AllowedImagePrefix)
	out.ArchiveURL = strings.TrimSpace(out.ArchiveURL)
	out.LoadedImage = strings.TrimSpace(out.LoadedImage)
	out.ServiceName = strings.TrimSpace(out.ServiceName)
	out.ComposeProjectDir = strings.TrimSpace(out.ComposeProjectDir)
	out.ComposeFile = strings.TrimSpace(out.ComposeFile)
	out.DockerBinary = strings.TrimSpace(out.DockerBinary)
	out.ComposeBinary = strings.TrimSpace(out.ComposeBinary)
	return &out
}

func validateDeployConfig(cfg *DeployConfig) error {
	if cfg == nil {
		return nil
	}
	if cfg.Mode != "docker_compose" {
		return fmt.Errorf("unsupported mode: %s", cfg.Mode)
	}
	if cfg.SourceType != "docker_image" && cfg.SourceType != "docker_archive_url" {
		return fmt.Errorf("unsupported source_type: %s", cfg.SourceType)
	}
	if cfg.Enabled {
		if cfg.SourceType == "docker_image" && cfg.DefaultImage == "" {
			return fmt.Errorf("default_image is required when deployment is enabled")
		}
		if cfg.SourceType == "docker_archive_url" {
			if cfg.ArchiveURL == "" {
				return fmt.Errorf("archive_url is required when archive deployment is enabled")
			}
			if err := validateDownloadURL(cfg.ArchiveURL); err != nil {
				return err
			}
			if cfg.LoadedImage == "" {
				return fmt.Errorf("loaded_image is required when archive deployment is enabled")
			}
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

	image := strings.TrimSpace(cfg.DefaultImage)
	if cfg.SourceType == "docker_image" && req != nil && strings.TrimSpace(req.Image) != "" {
		image = strings.TrimSpace(req.Image)
	}
	if cfg.SourceType == "docker_image" && image == "" {
		return nil, infraerrors.BadRequest("DEPLOY_IMAGE_REQUIRED", "deploy image is required")
	}
	if cfg.SourceType == "docker_image" {
		if prefix := strings.TrimSpace(cfg.AllowedImagePrefix); prefix != "" && !strings.HasPrefix(image, prefix) {
			return nil, infraerrors.BadRequest("DEPLOY_IMAGE_FORBIDDEN", "deploy image does not match allowed prefix")
		}
	} else {
		image = cfg.DefaultImage
		if image == "" {
			image = cfg.LoadedImage
		}
	}

	commands := s.buildDeployCommands(cfg, image)
	result := &DeployResult{
		Status:            deployStatusPending,
		Image:             image,
		ServiceName:       cfg.ServiceName,
		ComposeProjectDir: cfg.ComposeProjectDir,
		Commands:          commands,
		Message:           "Deploy request recorded",
		NeedRestart:       false,
	}

	now := time.Now().Unix()
	state := &DeployState{
		Status:         deployStatusPending,
		RequestedImage: image,
		LastMessage:    result.Message,
		StartedAt:      &now,
	}
	_ = s.saveDeployState(ctx, state)

	if req != nil && req.DryRun {
		result.Message = "Dry-run only; no deploy command executed"
		return result, nil
	}

	output, err := s.executeDeployCommands(ctx, cfg, image)
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

	state.Status = deployStatusSucceeded
	state.LastMessage = "Deploy completed successfully"
	state.LastError = ""
	_ = s.saveDeployState(context.Background(), state)

	result.Status = deployStatusSucceeded
	result.Message = "Deploy completed successfully"
	return result, nil
}

func (s *UpdateService) buildDeployCommands(cfg *DeployConfig, image string) []string {
	if cfg.SourceType == "docker_archive_url" {
		archivePath := filepath.Join(cfg.ComposeProjectDir, "deploy-update.tar")
		commands := []string{
			fmt.Sprintf("download %s -> %s", cfg.ArchiveURL, archivePath),
			fmt.Sprintf("%s load -i %s", cfg.DockerBinary, archivePath),
		}
		if cfg.LoadedImage != "" && cfg.LoadedImage != image {
			commands = append(commands, fmt.Sprintf("%s tag %s %s", cfg.DockerBinary, cfg.LoadedImage, image))
		}
		composeTarget := fmt.Sprintf("%s up -d --no-deps %s", cfg.ComposeBinary, cfg.ServiceName)
		if cfg.ComposeFile != "" {
			composeTarget = fmt.Sprintf("%s -f %s up -d --no-deps %s", cfg.ComposeBinary, cfg.ComposeFile, cfg.ServiceName)
		}
		commands = append(commands, composeTarget)
		return commands
	}
	commands := []string{
		fmt.Sprintf("%s pull %s", cfg.DockerBinary, image),
	}
	if image != cfg.DefaultImage {
		commands = append(commands, fmt.Sprintf("%s tag %s %s", cfg.DockerBinary, image, cfg.DefaultImage))
	}
	composeTarget := fmt.Sprintf("%s up -d --no-deps %s", cfg.ComposeBinary, cfg.ServiceName)
	if cfg.ComposeFile != "" {
		composeTarget = fmt.Sprintf("%s -f %s up -d --no-deps %s", cfg.ComposeBinary, cfg.ComposeFile, cfg.ServiceName)
	}
	commands = append(commands, composeTarget)
	return commands
}

func (s *UpdateService) executeDeployCommands(ctx context.Context, cfg *DeployConfig, image string) (string, error) {
	if s.deployRunner == nil {
		return "", fmt.Errorf("deployment runner is not configured")
	}

	if cfg.SourceType == "docker_archive_url" {
		archivePath := filepath.Join(cfg.ComposeProjectDir, "deploy-update.tar")
		if err := validateDownloadURL(cfg.ArchiveURL); err != nil {
			return "", err
		}
		if err := s.githubClient.DownloadFile(ctx, cfg.ArchiveURL, archivePath, maxDownloadSize); err != nil {
			return "", err
		}
		out, err := s.deployRunner.Run(ctx, cfg.ComposeProjectDir, cfg.DockerBinary, "load", "-i", archivePath)
		outputs := []string{}
		if out != "" {
			outputs = append(outputs, out)
		}
		if err != nil {
			return strings.Join(outputs, "\n"), err
		}
		if cfg.LoadedImage != "" && cfg.LoadedImage != image {
			out, err = s.deployRunner.Run(ctx, cfg.ComposeProjectDir, cfg.DockerBinary, "tag", cfg.LoadedImage, image)
			if out != "" {
				outputs = append(outputs, out)
			}
			if err != nil {
				return strings.Join(outputs, "\n"), err
			}
		}
		composeArgs := []string{"up", "-d", "--no-deps", cfg.ServiceName}
		if cfg.ComposeFile != "" {
			composeArgs = append([]string{"-f", cfg.ComposeFile}, composeArgs...)
		}
		out, err = s.deployRunner.Run(ctx, cfg.ComposeProjectDir, cfg.ComposeBinary, composeArgs...)
		if out != "" {
			outputs = append(outputs, out)
		}
		return strings.Join(outputs, "\n"), err
	}

	var outputs []string
	out, err := s.deployRunner.Run(ctx, cfg.ComposeProjectDir, cfg.DockerBinary, "pull", image)
	if out != "" {
		outputs = append(outputs, out)
	}
	if err != nil {
		return strings.Join(outputs, "\n"), err
	}

	if image != cfg.DefaultImage {
		out, err = s.deployRunner.Run(ctx, cfg.ComposeProjectDir, cfg.DockerBinary, "tag", image, cfg.DefaultImage)
		if out != "" {
			outputs = append(outputs, out)
		}
		if err != nil {
			return strings.Join(outputs, "\n"), err
		}
	}

	composeArgs := []string{"up", "-d", "--no-deps", cfg.ServiceName}
	if cfg.ComposeFile != "" {
		composeArgs = append([]string{"-f", cfg.ComposeFile}, composeArgs...)
	}
	out, err = s.deployRunner.Run(ctx, cfg.ComposeProjectDir, cfg.ComposeBinary, composeArgs...)
	if out != "" {
		outputs = append(outputs, out)
	}
	return strings.Join(outputs, "\n"), err
}
