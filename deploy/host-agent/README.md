# Sub2API Git Deploy Host Agent

这套方案保留 Docker 运行方式，只把部署动作挪到宿主机执行。

## 组成

- 宿主机真实部署脚本：`deploy-from-git.sh`
- 轻量 HTTP 触发器：`sub2api_host_deploy_agent.py`
- systemd 服务文件：`sub2api-host-deploy-agent.service`

## 手动部署

把脚本复制到服务器：

```bash
mkdir -p /home/ec2-user/sub2api-deploy/bin
cp deploy/host-agent/deploy-from-git.sh /home/ec2-user/sub2api-deploy/bin/
chmod +x /home/ec2-user/sub2api-deploy/bin/deploy-from-git.sh
```

直接执行：

```bash
/home/ec2-user/sub2api-deploy/bin/deploy-from-git.sh
```

自定义参数执行：

```bash
REPO_URL="https://github.com/YuHaiA/sub2api.git" \
BRANCH="main" \
REPO_DIR="/home/ec2-user/sub2api-source" \
COMPOSE_PROJECT_DIR="/home/ec2-user/sub2api-deploy" \
SERVICE_NAME="sub2api" \
/home/ec2-user/sub2api-deploy/bin/deploy-from-git.sh
```

## 后台一键部署

安装轻量触发器：

```bash
sudo mkdir -p /opt/sub2api-host-agent
sudo cp deploy/host-agent/sub2api_host_deploy_agent.py /opt/sub2api-host-agent/
sudo chmod +x /opt/sub2api-host-agent/sub2api_host_deploy_agent.py
sudo cp deploy/host-agent/sub2api-host-deploy-agent.service /etc/systemd/system/
sudo sed -i 's/change-me/请替换成你的长随机令牌/' /etc/systemd/system/sub2api-host-deploy-agent.service
sudo systemctl daemon-reload
sudo systemctl enable --now sub2api-host-deploy-agent
curl http://127.0.0.1:18080/health
```

后台建议配置：

- 仓库地址：`https://github.com/YuHaiA/sub2api.git`
- 分支：`main`
- 源码目录：`/home/ec2-user/sub2api-source`
- 部署目录：`/home/ec2-user/sub2api-deploy`
- 服务名：`sub2api`
- 代理地址：`http://172.17.0.1:18080`

## 依赖

宿主机需要：

- `git`
- `docker`
- `docker-compose`
- `python3`

数据库和 Redis 不需要重装，也不会被这套部署脚本重建。
