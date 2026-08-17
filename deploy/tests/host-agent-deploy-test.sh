#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
test_root="$(mktemp -d "${TMPDIR:-/tmp}/sub2api-host-deploy-test.XXXXXX")"
state_dir="$test_root/state"
bin_dir="$test_root/bin"
compose_dir="$test_root/compose"

cleanup() {
  rm -rf "$test_root"
}
trap cleanup EXIT

mkdir -p "$state_dir" "$bin_dir" "$compose_dir"

cat > "$bin_dir/docker" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

old_id='sha256:old-image-id'
new_id='sha256:new-image-id'

if [[ "$1" == "inspect" ]]; then
  format="$3"
  case "$format" in
    '{{.Image}}') [[ -f "$FAKE_DEPLOY_STATE/running-new" ]] && echo "$new_id" || echo "$old_id" ;;
    '{{.Config.Image}}') echo 'ghcr.io/yuhaia/sub2api:main-old' ;;
    *Health.Status*) echo 'healthy' ;;
    '{{.State.StartedAt}}') echo '2026-08-17T00:00:00Z' ;;
  esac
  exit 0
fi

if [[ "$1 $2" == "image inspect" ]]; then
  image="$3"
  if [[ "$image" == 'sub2api-gha:docker-deploy' || -f "$FAKE_DEPLOY_STATE/tagged-${image//\//_}" ]]; then
    echo "$new_id"
    exit 0
  fi
  exit 1
fi

if [[ "$1" == "load" ]]; then
  touch "$FAKE_DEPLOY_STATE/loaded"
  echo 'Loaded image: sub2api-gha:docker-deploy'
  exit 0
fi

if [[ "$1" == "tag" ]]; then
  destination="$3"
  touch "$FAKE_DEPLOY_STATE/tagged-${destination//\//_}"
  [[ "$destination" == 'ghcr.io/yuhaia/sub2api:main-old' ]] && touch "$FAKE_DEPLOY_STATE/runtime-tagged"
  exit 0
fi

if [[ "$1" == "images" ]]; then
  exit 0
fi

if [[ "$1 $2" == "image prune" || "$1" == "rmi" || "$1" == "ps" || "$1" == "logs" ]]; then
  exit 0
fi

printf 'unexpected docker invocation: %s\n' "$*" >&2
exit 1
EOF

cat > "$bin_dir/docker-compose" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

[[ " $* " == *' --force-recreate '* ]] || { echo 'missing --force-recreate' >&2; exit 1; }
[[ -f "$FAKE_DEPLOY_STATE/runtime-tagged" ]] || { echo 'compose image reference was not retagged' >&2; exit 1; }
touch "$FAKE_DEPLOY_STATE/running-new"
EOF

cat > "$bin_dir/curl" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

if [[ " $* " == *' --output '* ]]; then
  while [[ $# -gt 0 ]]; do
    if [[ "$1" == '--output' ]]; then
      touch "$2"
      exit 0
    fi
    shift
  done
fi
echo 'release-sha256'
EOF

chmod +x "$bin_dir/docker" "$bin_dir/docker-compose" "$bin_dir/curl"
export FAKE_DEPLOY_STATE="$state_dir"
export PATH="$bin_dir:$PATH"

output="$({
  ARCHIVE_URL='https://example.invalid/sub2api.tar' \
  ARCHIVE_SHA_URL='https://example.invalid/sub2api.tar.sha256' \
  IMAGE_TAG='sub2api:rollback' \
  COMPOSE_PROJECT_DIR="$compose_dir" \
  ARCHIVE_PATH="$test_root/deploy-update.tar" \
  METADATA_DIR="$test_root/metadata" \
  HEALTH_WAIT_SECONDS=1 \
  HEALTH_POLL_INTERVAL=1 \
  "$repo_root/deploy/host-agent/deploy-from-package.sh"
} 2>&1)"

grep -q 'docker tag sub2api-gha:docker-deploy -> ghcr.io/yuhaia/sub2api:main-old' <<<"$output"
grep -q 'running image verified: sha256:new-image-id' <<<"$output"
grep -q 'image_id=sha256:new-image-id running_image_id=sha256:new-image-id' <<<"$output"
[[ -f "$state_dir/running-new" ]]

printf 'host agent deploy test passed\n'
