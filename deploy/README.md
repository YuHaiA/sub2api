# Sub2API Deploy Guide

This directory contains deployment files used by this repository.

## Main Files

- `docker-deploy.sh`: one-click Docker deployment preparation script
- `.env.example`: Docker environment variable template
- `docker-compose.local.yml`: local-directory Docker deployment
- `docker-compose.yml`: named-volume Docker deployment
- `install.sh`: binary installation script
- `config.example.yaml`: example runtime config
- `host-agent/`: host-side package deployment scripts

## Recommended Docker Start

```bash
mkdir -p sub2api-deploy && cd sub2api-deploy
curl -sSL https://raw.githubusercontent.com/YuHaiA/sub2api/main/deploy/docker-deploy.sh | bash
docker compose up -d
docker compose logs -f sub2api
```

## Manual Docker Start

```bash
git clone https://github.com/YuHaiA/sub2api.git
cd sub2api/deploy
cp .env.example .env
mkdir -p data postgres_data redis_data
docker compose -f docker-compose.local.yml up -d
```

## Host Package Deploy

If the server is too small to build Docker images locally, use the host-side package deployment flow:

- package deployment script: `host-agent/deploy-from-package.sh`
- optional HTTP trigger: `host-agent/sub2api_host_deploy_agent.py`

See [`host-agent/README.md`](host-agent/README.md) for details.
