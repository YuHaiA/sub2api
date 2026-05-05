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
METADATA_DIR="${METADATA_DIR:-$COMPOSE_PROJECT_DIR/.deploy-state}"
ARCHIVE_ETAG_FILE="${ARCHIVE_ETAG_FILE:-$METADATA_DIR/archive.etag}"
KEEP_BACKUPS="${KEEP_BACKUPS:-2}"
HEALTH_WAIT_SECONDS="${HEALTH_WAIT_SECONDS:-120}"
HEALTH_POLL_INTERVAL="${HEALTH_POLL_INTERVAL:-5}"

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

fetch_archive_etag() {
  curl -fsSLI "$ARCHIVE_URL" | awk -F': ' 'tolower($1)=="etag" {gsub("\r","",$2); print $2}' | tail -n 1
}

load_cached_archive_etag() {
  if [[ -f "$ARCHIVE_ETAG_FILE" ]]; then
    tr -d '\r\n' < "$ARCHIVE_ETAG_FILE"
  fi
}

save_cached_archive_etag() {
  local etag="$1"
  if [[ -n "$etag" ]]; then
    mkdir -p "$METADATA_DIR"
    printf '%s\n' "$etag" > "$ARCHIVE_ETAG_FILE"
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

cleanup_archive() {
  if [[ -f "$ARCHIVE_PATH" ]]; then
    rm -f "$ARCHIVE_PATH"
    log "cleanup archive: $ARCHIVE_PATH"
  fi
}

get_running_image_id() {
  "$DOCKER_BINARY" inspect --format '{{.Image}}' "$SERVICE_NAME" 2>/dev/null || true
}

get_loaded_image_id() {
  "$DOCKER_BINARY" image inspect "$LOADED_IMAGE" --format '{{.Id}}' 2>/dev/null || true
}

get_latest_image_id() {
  "$DOCKER_BINARY" image inspect "$IMAGE_TAG" --format '{{.Id}}' 2>/dev/null || true
}

already_up_to_date() {
  local running_image_id latest_image_id
  running_image_id="$(get_running_image_id)"
  latest_image_id="$(get_latest_image_id)"
  [[ -n "$running_image_id" && -n "$latest_image_id" && "$running_image_id" == "$latest_image_id" ]]
}

archive_matches_running() {
  local running_image_id loaded_image_id
  running_image_id="$(get_running_image_id)"
  loaded_image_id="$(get_loaded_image_id)"
  [[ -n "$running_image_id" && -n "$loaded_image_id" && "$running_image_id" == "$loaded_image_id" ]]
}

release_unchanged() {
  local current_etag cached_etag
  current_etag="$(fetch_archive_etag || true)"
  cached_etag="$(load_cached_archive_etag || true)"
  [[ -n "$current_etag" && -n "$cached_etag" && "$current_etag" == "$cached_etag" ]]
}

prune_old_backups() {
  local repo="${IMAGE_TAG%:*}"
  local keep="${KEEP_BACKUPS}"
  mapfile -t backups < <("$DOCKER_BINARY" images --format '{{.Repository}}:{{.Tag}} {{.CreatedAt}}' | awk -v repo="$repo" '$1 ~ ("^" repo ":backup-") {print $0}')
  local count=${#backups[@]}
  if (( count <= keep )); then
    log "backup images within keep limit: $count/$keep"
    return
  fi

  mapfile -t backup_refs < <(printf '%s\n' "${backups[@]}" | sort -rk2 | awk '{print $1}')
  local idx=0
  for image_ref in "${backup_refs[@]}"; do
    idx=$((idx + 1))
    if (( idx <= keep )); then
      continue
    fi
    log "remove old backup image: $image_ref"
    "$DOCKER_BINARY" rmi "$image_ref" >/dev/null 2>&1 || log "warn: failed to remove $image_ref"
  done
}

prune_unused_images() {
  log "prune unused docker images"
  "$DOCKER_BINARY" image prune -a -f >/dev/null 2>&1 || log "warn: failed to prune unused images"
}

wait_for_health() {
  local deadline=$(( $(date +%s) + HEALTH_WAIT_SECONDS ))
  while (( $(date +%s) <= deadline )); do
    local status
    status=$("$DOCKER_BINARY" inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$SERVICE_NAME" 2>/dev/null || true)
    if [[ "$status" == "healthy" || "$status" == "running" ]]; then
      log "service status: $status"
      return 0
    fi
    sleep "$HEALTH_POLL_INTERVAL"
  done
  log "health check timeout after ${HEALTH_WAIT_SECONDS}s"
  "$DOCKER_BINARY" ps --filter "name=^/${SERVICE_NAME}$" || true
  "$DOCKER_BINARY" logs --tail 100 "$SERVICE_NAME" || true
  return 1
}

show_result() {
  local image_id container_image started_at
  image_id=$("$DOCKER_BINARY" image inspect "$IMAGE_TAG" --format '{{.Id}}' 2>/dev/null || true)
  container_image=$("$DOCKER_BINARY" inspect --format '{{.Image}}' "$SERVICE_NAME" 2>/dev/null || true)
  started_at=$("$DOCKER_BINARY" inspect --format '{{.State.StartedAt}}' "$SERVICE_NAME" 2>/dev/null || true)
  log "result image_tag=$IMAGE_TAG image_id=$image_id container_image=$container_image started_at=$started_at"
}

main() {
  require_command curl
  require_command "$DOCKER_BINARY"
  require_command "$COMPOSE_BINARY"

  mkdir -p "$(dirname "$ARCHIVE_PATH")"
  mkdir -p "$METADATA_DIR"

  trap cleanup_archive EXIT

  if already_up_to_date && release_unchanged; then
    log "already up to date; skip download and deploy"
    prune_old_backups
    prune_unused_images
    show_result
    log "deploy completed successfully"
    return 0
  fi

  download_archive

  log "docker load package"
  "$DOCKER_BINARY" load -i "$ARCHIVE_PATH"

  if archive_matches_running; then
    log "already up to date; skip deploy"
    save_cached_archive_etag "$(fetch_archive_etag || true)"
    prune_old_backups
    prune_unused_images
    show_result
    log "deploy completed successfully"
    return 0
  fi

  backup_current_image

  log "docker tag $LOADED_IMAGE -> $IMAGE_TAG"
  "$DOCKER_BINARY" tag "$LOADED_IMAGE" "$IMAGE_TAG"

  cd "$COMPOSE_PROJECT_DIR"
  log "docker compose restart target service: $SERVICE_NAME"
  compose_up

  wait_for_health
  save_cached_archive_etag "$(fetch_archive_etag || true)"
  prune_old_backups
  prune_unused_images
  show_result
  log "deploy completed successfully"
}

main "$@"
