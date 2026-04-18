# Sub2API

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.25.7-00ADD8.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.4+-4FC08D.svg)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791.svg)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7+-DC382D.svg)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg)](https://www.docker.com/)

**AI API gateway platform for account management, API key distribution, billing, and subscription operations**

English | [中文](README_CN.md) | [日本語](README_JA.md)

</div>

## Overview

Sub2API is an AI API gateway and admin platform built around upstream account management.
It lets you manage multiple upstream accounts, expose unified API keys to end users, track usage and billing, and operate the service through a web admin panel.

This repository includes:

- Go backend
- Vue 3 admin / user frontend
- PostgreSQL + Redis support
- Docker Compose deployment files
- Host-side package deployment scripts for GitHub release artifacts

## Acknowledgement

This repository is a secondary development / customized fork based on the original project:

- Upstream project: [Wei-Shaw/sub2api](https://github.com/Wei-Shaw/sub2api)

Thanks to the original author and contributors of `Wei-Shaw/sub2api` for the foundational architecture and implementation.

## Main Features

- Multi-account management for OAuth, API key, setup token, and upstream account types
- API key creation, distribution, and usage tracking
- User, group, channel, and subscription management
- Token-level billing, balance management, and payment integration
- Account health check, scheduled testing, deduplication, and unhealthy cleanup
- Proxy management, usage cleanup, request error handling, and ops dashboard
- Backup / restore and optional host-side data management integration

## Recommended Start: Docker Compose

### Prerequisites

- Docker 20.10+
- Docker Compose v2+

### One-click deployment

```bash
mkdir -p sub2api-deploy && cd sub2api-deploy
curl -sSL https://raw.githubusercontent.com/YuHaiA/sub2api/main/deploy/docker-deploy.sh | bash
docker compose up -d
docker compose logs -f sub2api
```

The script will:

- download `docker-compose.local.yml` as `docker-compose.yml`
- download `.env.example`
- generate secure defaults for `POSTGRES_PASSWORD`, `JWT_SECRET`, and `TOTP_ENCRYPTION_KEY`
- create `data/`, `postgres_data/`, and `redis_data/`

### Manual deployment

```bash
git clone https://github.com/YuHaiA/sub2api.git
cd sub2api/deploy
cp .env.example .env
mkdir -p data postgres_data redis_data
docker compose -f docker-compose.local.yml up -d
docker compose -f docker-compose.local.yml logs -f sub2api
```

## Important Environment Variables

The full template is in [`deploy/.env.example`](deploy/.env.example).

Most important variables:

```bash
POSTGRES_PASSWORD=change_this_secure_password
POSTGRES_USER=sub2api
POSTGRES_DB=sub2api

SERVER_PORT=8080
SERVER_MODE=release
RUN_MODE=standard
TZ=Asia/Shanghai

ADMIN_EMAIL=admin@sub2api.local
ADMIN_PASSWORD=

JWT_SECRET=
TOTP_ENCRYPTION_KEY=
```

Notes:

- `POSTGRES_PASSWORD` is required
- `JWT_SECRET` should be fixed in production, otherwise users may be logged out after restart
- `TOTP_ENCRYPTION_KEY` should be fixed in production, otherwise existing 2FA data may become invalid
- leaving `ADMIN_PASSWORD` empty will auto-generate it on first startup and print it in logs

## Access and First Login

After the service starts:

```text
http://YOUR_SERVER_IP:8080
```

If the admin password was auto-generated:

```bash
docker compose -f docker-compose.local.yml logs sub2api | grep "admin password"
```

## Upgrade

### Option 1: Pull latest image

```bash
docker compose -f docker-compose.local.yml pull
docker compose -f docker-compose.local.yml up -d
```

### Option 2: GitHub package + host deployment

This repository includes host-side package deployment scripts under [`deploy/host-agent`](deploy/host-agent).

Recommended flow:

1. Push code to GitHub
2. Let GitHub Actions publish the latest Docker package artifact
3. On the host machine, run the package deployment script

Manual example:

```bash
/home/ec2-user/sub2api-deploy/bin/deploy-from-package.sh
```

This flow is useful when the server is too small to build Docker images locally.

## Build From Source

For development or custom builds:

```bash
git clone https://github.com/YuHaiA/sub2api.git
cd sub2api
cd frontend
pnpm install
pnpm run build
cd ../backend
go build -tags embed -o sub2api ./cmd/server
```

## Optional Components

- `datamanagementd`: optional host-side data management component
- host deploy agent: optional host-side HTTP trigger for package deployment

Related files:

- [`deploy/DATAMANAGEMENTD_CN.md`](deploy/DATAMANAGEMENTD_CN.md)
- [`deploy/host-agent/README.md`](deploy/host-agent/README.md)

## Repo Notes

- This README intentionally removes sponsor banners, demo promotions, and unrelated marketing blocks
- Deployment examples in this fork point to `YuHaiA/sub2api`
- The main purpose here is clarity: what the project does, how to start it, how to configure it, and which features are actually present
