#!/usr/bin/env python3
import json
import os
import subprocess
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from typing import Any, Dict

DEFAULT_HOST = os.environ.get("SUB2API_DEPLOY_AGENT_HOST", "0.0.0.0")
DEFAULT_PORT = int(os.environ.get("SUB2API_DEPLOY_AGENT_PORT", "18080"))
AGENT_TOKEN = os.environ.get("SUB2API_DEPLOY_AGENT_TOKEN", "").strip()
SCRIPT_PATH = os.environ.get(
    "SUB2API_DEPLOY_SCRIPT",
    "/home/ec2-user/sub2api-deploy/bin/deploy-from-package.sh",
).strip()


def json_response(handler: BaseHTTPRequestHandler, status: int, payload: Dict[str, Any]) -> None:
    data = json.dumps(payload, ensure_ascii=False).encode("utf-8")
    handler.send_response(status)
    handler.send_header("Content-Type", "application/json; charset=utf-8")
    handler.send_header("Content-Length", str(len(data)))
    handler.end_headers()
    handler.wfile.write(data)


def require_auth(handler: BaseHTTPRequestHandler) -> bool:
    if not AGENT_TOKEN:
        return True
    auth = handler.headers.get("Authorization", "")
    if auth == f"Bearer {AGENT_TOKEN}":
        return True
    json_response(handler, 401, {"status": "failed", "error": "unauthorized", "message": "invalid bearer token"})
    return False


def read_json(handler: BaseHTTPRequestHandler) -> Dict[str, Any]:
    length = int(handler.headers.get("Content-Length", "0") or "0")
    raw = handler.rfile.read(length)
    if not raw:
        return {}
    return json.loads(raw.decode("utf-8"))


def build_env(payload: Dict[str, Any]) -> Dict[str, str]:
    env = os.environ.copy()
    mapping = {
        "archive_url": "ARCHIVE_URL",
        "loaded_image": "LOADED_IMAGE",
        "default_image": "IMAGE_TAG",
        "compose_project_dir": "COMPOSE_PROJECT_DIR",
        "compose_file": "COMPOSE_FILE",
        "service_name": "SERVICE_NAME",
        "docker_binary": "DOCKER_BINARY",
        "compose_binary": "COMPOSE_BINARY",
    }
    for payload_key, env_key in mapping.items():
        value = str(payload.get(payload_key, "") or "").strip()
        if value:
            env[env_key] = value
    return env


def handle_deploy(payload: Dict[str, Any]) -> Dict[str, Any]:
    if str(payload.get("source_type", "")).strip() != "docker_archive_url":
        raise RuntimeError("unsupported source_type")
    if not os.path.isfile(SCRIPT_PATH):
        raise RuntimeError(f"deploy script not found: {SCRIPT_PATH}")

    completed = subprocess.run(
        [SCRIPT_PATH],
        env=build_env(payload),
        capture_output=True,
        text=True,
        check=False,
    )
    output = "\n".join(part.strip() for part in [completed.stdout, completed.stderr] if part and part.strip()).strip()
    if completed.returncode != 0:
        raise RuntimeError(output or "deploy script failed")

    return {
        "status": "succeeded",
        "image": str(payload.get("default_image", "")).strip(),
        "service_name": str(payload.get("service_name", "")).strip(),
        "compose_project_dir": str(payload.get("compose_project_dir", "")).strip(),
        "commands": payload.get("commands") or [],
        "message": "Deploy completed successfully",
        "output": output,
        "need_restart": False,
    }


class Handler(BaseHTTPRequestHandler):
    server_version = "sub2api-host-deploy-agent/1.0"

    def do_GET(self) -> None:  # noqa: N802
        if self.path.rstrip("/") == "/health":
            json_response(self, 200, {"status": "ok", "service": "sub2api-host-deploy-agent"})
            return
        json_response(self, 404, {"status": "failed", "error": "not_found", "message": "endpoint not found"})

    def do_POST(self) -> None:  # noqa: N802
        if self.path.rstrip("/") != "/deploy":
            json_response(self, 404, {"status": "failed", "error": "not_found", "message": "endpoint not found"})
            return
        if not require_auth(self):
            return
        try:
            payload = read_json(self)
            result = handle_deploy(payload)
            json_response(self, 200, result)
        except Exception as exc:  # noqa: BLE001
            json_response(self, 500, {"status": "failed", "error": str(exc), "message": "deploy failed"})

    def log_message(self, format: str, *args: Any) -> None:  # noqa: A003
        print(f"[sub2api-host-deploy-agent] {self.address_string()} - {format % args}")


def main() -> None:
    server = ThreadingHTTPServer((DEFAULT_HOST, DEFAULT_PORT), Handler)
    print(f"sub2api host deploy agent listening on {DEFAULT_HOST}:{DEFAULT_PORT}")
    server.serve_forever()


if __name__ == "__main__":
    main()
