#!/usr/bin/env bash
set -euo pipefail

REPO_URL="${REPO_URL:-https://github.com/YuHaiA/sub2api.git}"
BRANCH="${BRANCH:-main}"
REPO_DIR="${REPO_DIR:-/home/ec2-user/sub2api-source}"
COMPOSE_PROJECT_DIR="${COMPOSE_PROJECT_DIR:-/home/ec2-user/sub2api-deploy}"
COMPOSE_FILE="${COMPOSE_FILE:-}"
SERVICE_NAME="${SERVICE_NAME:-sub2api}"
IMAGE_TAG="${IMAGE_TAG:-weishaw/sub2api:latest}"
DOCKER_BINARY="${DOCKER_BINARY:-docker}"
COMPOSE_BINARY="${COMPOSE_BINARY:-docker-compose}"

timestamp() {
  date '+%Y-%m-%d %H:%M:%S'
}

log() {
  printf '[%s] %s\n' "$(timestamp)" "$*"
}

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    log "missing command: $1"
    exit 1
  fi
}

compose_up() {
  if [[ -n "$COMPOSE_FILE" ]]; then
    "$COMPOSE_BINARY" -f "$COMPOSE_FILE" up -d --no-deps "$SERVICE_NAME"
  else
    "$COMPOSE_BINARY" up -d --no-deps "$SERVICE_NAME"
  fi
}

backup_current_image() {
  if ! "$DOCKER_BINARY" image inspect "$IMAGE_TAG" >/dev/null 2>&1; then
    log "skip backup; image not found: $IMAGE_TAG"
    return
  fi

  local backup_tag
  if [[ "$IMAGE_TAG" == *:* ]]; then
    backup_tag="${IMAGE_TAG%:*}:backup-$(date '+%Y%m%d%H%M%S')"
  else
    backup_tag="${IMAGE_TAG}:backup-$(date '+%Y%m%d%H%M%S')"
  fi

  log "backup image: $IMAGE_TAG -> $backup_tag"
  "$DOCKER_BINARY" tag "$IMAGE_TAG" "$backup_tag"
}

main() {
  require_command git
  require_command "$DOCKER_BINARY"
  require_command "$COMPOSE_BINARY"

  mkdir -p "$(dirname "$REPO_DIR")"

  if [[ ! -d "$REPO_DIR/.git" ]]; then
    log "clone repository: $REPO_URL -> $REPO_DIR"
    git clone "$REPO_URL" "$REPO_DIR"
  fi

  cd "$REPO_DIR"
  log "git fetch origin"
  git fetch origin
  log "git checkout $BRANCH"
  git checkout "$BRANCH"
  log "git pull --ff-only origin $BRANCH"
  git pull --ff-only origin "$BRANCH"

  backup_current_image

  log "docker build -t $IMAGE_TAG ."
  "$DOCKER_BINARY" build -t "$IMAGE_TAG" .

  cd "$COMPOSE_PROJECT_DIR"
  log "docker compose restart target service: $SERVICE_NAME"
  compose_up

  log "deploy completed successfully"
}

main "$@"
