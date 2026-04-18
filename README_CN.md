# Sub2API

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.25.7-00ADD8.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.4+-4FC08D.svg)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791.svg)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7+-DC382D.svg)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg)](https://www.docker.com/)

**面向账号管理、API Key 分发、计费与订阅运营的 AI API 网关平台**

[English](README.md) | 中文 | [日本語](README_JA.md)

</div>

## 项目说明

Sub2API 是一个 AI API 网关和运营管理平台，核心目标是把上游账号能力统一对外提供，并通过后台完成账号管理、API Key 分发、计费、订阅和运维。

当前仓库包含：

- Go 后端
- Vue 3 管理端 / 用户端前端
- PostgreSQL + Redis 支持
- Docker Compose 部署文件
- GitHub 镜像包 + 宿主机更新脚本

## 致谢与来源

本仓库是基于原项目进行二次开发和定制维护的版本：

- 原项目： [Wei-Shaw/sub2api](https://github.com/Wei-Shaw/sub2api)

感谢 `Wei-Shaw/sub2api` 原作者和贡献者提供的基础架构与实现。

## 主要功能

- 多账号管理：支持 OAuth、API Key、Setup Token、Upstream 等账号类型
- API Key 管理：生成、分发、统计和配额控制
- 用户 / 分组 / 渠道 / 订阅管理
- Token 级计费、余额管理、支付集成
- 账号测活、定时测试、重复账号清理、异常账号清理
- 代理管理、用量清理、错误请求处理、运维监控面板
- 备份 / 恢复，以及可选的数据管理宿主机进程联动

## 推荐启动方式：Docker Compose

### 前置条件

- Docker 20.10+
- Docker Compose v2+

### 一键部署

```bash
mkdir -p sub2api-deploy && cd sub2api-deploy
curl -sSL https://raw.githubusercontent.com/YuHaiA/sub2api/main/deploy/docker-deploy.sh | bash
docker compose up -d
docker compose logs -f sub2api
```

这个脚本会自动：

- 下载 `docker-compose.local.yml` 并保存为 `docker-compose.yml`
- 下载 `.env.example`
- 自动生成 `POSTGRES_PASSWORD`、`JWT_SECRET`、`TOTP_ENCRYPTION_KEY`
- 创建 `data/`、`postgres_data/`、`redis_data/`

### 手动部署

```bash
git clone https://github.com/YuHaiA/sub2api.git
cd sub2api/deploy
cp .env.example .env
mkdir -p data postgres_data redis_data
docker compose -f docker-compose.local.yml up -d
docker compose -f docker-compose.local.yml logs -f sub2api
```

## 关键环境变量

完整模板见 [`deploy/.env.example`](deploy/.env.example)。

最重要的几个：

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

说明：

- `POSTGRES_PASSWORD` 必填
- `JWT_SECRET` 生产环境建议固定，否则重启后用户登录态可能失效
- `TOTP_ENCRYPTION_KEY` 生产环境建议固定，否则已有 2FA 数据可能失效
- `ADMIN_PASSWORD` 留空时会在首次启动时自动生成，并打印到日志中

## 启动后访问

服务启动完成后访问：

```text
http://你的服务器IP:8080
```

如果管理员密码是自动生成的，可以这样查：

```bash
docker compose -f docker-compose.local.yml logs sub2api | grep "admin password"
```

## 升级方式

### 方式 1：直接拉最新镜像

```bash
docker compose -f docker-compose.local.yml pull
docker compose -f docker-compose.local.yml up -d
```

### 方式 2：GitHub 镜像包 + 宿主机部署

仓库里已经包含宿主机更新脚本，目录在 [`deploy/host-agent`](deploy/host-agent)。

推荐流程：

1. 本地改代码后 push 到 GitHub
2. GitHub Actions 产出最新 Docker 镜像包
3. 宿主机执行镜像包部署脚本

手动执行示例：

```bash
/home/ec2-user/sub2api-deploy/bin/deploy-from-package.sh
```

这套方式尤其适合磁盘和性能都比较紧张、不适合服务器本地构建镜像的场景。

## 源码编译

如果你要开发或定制版本：

```bash
git clone https://github.com/YuHaiA/sub2api.git
cd sub2api
cd frontend
pnpm install
pnpm run build
cd ../backend
go build -tags embed -o sub2api ./cmd/server
```

## 可选组件

- `datamanagementd`：可选的宿主机数据管理进程
- 宿主机部署代理：可选的 HTTP 触发器，用于后台一键更新

相关说明：

- [`deploy/DATAMANAGEMENTD_CN.md`](deploy/DATAMANAGEMENTD_CN.md)
- [`deploy/host-agent/README.md`](deploy/host-agent/README.md)

## 本次整理说明

- 已去掉 README 里的赞助图片、推广位和演示站宣传内容
- 已把主要启动和部署说明改成当前仓库 `YuHaiA/sub2api`
- 文档重点只保留“项目做什么、怎么启动、怎么配环境变量、有哪些主要功能”
