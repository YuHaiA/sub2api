#!/usr/bin/env bash
set -euo pipefail

ARCHIVE_URL="${ARCHIVE_URL:-https://github.com/YuHaiA/sub2api/releases/download/docker-deploy/sub2api-docker-image.tar}"
LOADED_IMAGE="${LOADED_IMAGE:-sub2api-gha:docker-deploy}"
IMAGE_TAG="${IMAGE_TAG:-weishaw/sub2api:latest}"
COMPOSE_PROJECT_DIR="${COMPOSE_PROJECT_DIR:-/home/ec2-user/sub2api-deploy}"
COMPOSE_FILE="${COMPOSE_FILE:-}"
SERVICE_NAME="${SERVICE_NAME:-sub2api}"
DOCKER_BINARY="${DOCKER_BINARY:-docker}"
COMPOSE_BINARY="${COMPOSE_BINARY:-docker-compose}"
ARCHIVE_PATH="${ARCHIVE_PATH:-$COMPOSE_PROJECT_DIR/deploy-update.tar}"

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

download_archive() {
  log "download package: $ARCHIVE_URL -> $ARCHIVE_PATH"
  curl -L --fail --output "$ARCHIVE_PATH" "$ARCHIVE_URL"
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

cleanup_archive() {
  if [[ -f "$ARCHIVE_PATH" ]]; then
    rm -f "$ARCHIVE_PATH"
    log "cleanup archive: $ARCHIVE_PATH"
  fi
}

main() {
  require_command curl
  require_command "$DOCKER_BINARY"
  require_command "$COMPOSE_BINARY"

  mkdir -p "$(dirname "$ARCHIVE_PATH")"

  trap cleanup_archive EXIT

  download_archive
  backup_current_image

  log "docker load package"
  "$DOCKER_BINARY" load -i "$ARCHIVE_PATH"

  log "docker tag $LOADED_IMAGE -> $IMAGE_TAG"
  "$DOCKER_BINARY" tag "$LOADED_IMAGE" "$IMAGE_TAG"

  cd "$COMPOSE_PROJECT_DIR"
  log "docker compose restart target service: $SERVICE_NAME"
  compose_up

  log "deploy completed successfully"
}

main "$@"
