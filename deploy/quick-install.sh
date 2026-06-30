#!/bin/bash
# =============================================================================
# Sub2API-NJ 一键安装脚本
# 含签到抽奖系统的自定义版本
# =============================================================================
# 使用方法:
#   curl -sSL https://raw.githubusercontent.com/15188037938/sub2api-nj/main/deploy/quick-install.sh | bash
# 或:
#   wget -qO- https://raw.githubusercontent.com/15188037938/sub2api-nj/main/deploy/quick-install.sh | bash
#
# 前置条件: Docker + Docker Compose
# 自动安装 Docker 如果不存在
# =============================================================================

set -e

# ============================================================
# 颜色 & 样式
# ============================================================
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

print_step() { echo -e "\n${CYAN}[${1}/${TOTAL_STEPS}]${NC} $2"; }
print_info() { echo -e "  ${BLUE}->${NC} $1"; }
print_ok()   { echo -e "  ${GREEN}OK${NC} $1"; }
print_warn() { echo -e "  ${YELLOW}WARN${NC} $1"; }
print_err()  { echo -e "  ${RED}ERR${NC} $1"; }

# ============================================================
# 配置
# ============================================================
REPO="15188037938/sub2api-nj"
REPO_URL="https://github.com/${REPO}.git"
INSTALL_DIR="${INSTALL_DIR:-/opt/sub2api-nj}"
SERVER_PORT="${SERVER_PORT:-8080}"
TOTAL_STEPS=7

# ============================================================
# Step 1: 检查 root
# ============================================================
STEP=1
if [ "$(id -u)" != "0" ]; then
  print_err "请以 root 用户运行（或 sudo bash quick-install.sh）"
  exit 1
fi

echo ""
echo "================================================"
echo "  Sub2API-NJ 一键安装"
echo "  含签到抽奖系统"
echo "================================================"

# ============================================================
# Step 2: 检查 / 安装 Docker
# ============================================================
STEP=2
print_step "$STEP" "检查 Docker 环境"

install_docker() {
  print_info "正在安装 Docker..."
  curl -fsSL https://get.docker.com | bash
  systemctl enable docker
  systemctl start docker
  print_ok "Docker 安装完成"
}

install_docker_compose() {
  print_info "正在安装 Docker Compose..."
  DOCKER_CONFIG=${DOCKER_CONFIG:-$HOME/.docker}
  mkdir -p "$DOCKER_CONFIG/cli-plugins"
  curl -SL "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" \
    -o "$DOCKER_CONFIG/cli-plugins/docker-compose"
  chmod +x "$DOCKER_CONFIG/cli-plugins/docker-compose"
  print_ok "Docker Compose 安装完成"
}

if ! command -v docker &>/dev/null; then
  print_warn "Docker 未安装，正在自动安装..."
  install_docker
else
  print_ok "Docker $(docker --version | awk '{print $3}' | sed 's/,//')"
fi

if ! docker compose version &>/dev/null; then
  print_warn "Docker Compose 未安装，正在自动安装..."
  install_docker_compose
else
  print_ok "Docker Compose $(docker compose version | awk '{print $4}')"
fi

# ============================================================
# Step 3: 克隆代码
# ============================================================
STEP=3
print_step "$STEP" "获取项目代码"

if [ -d "$INSTALL_DIR" ]; then
  print_warn "目录 $INSTALL_DIR 已存在，拉取最新代码..."
  cd "$INSTALL_DIR"
  git pull origin main 2>/dev/null || true
else
  print_info "克隆仓库到 $INSTALL_DIR"
  git clone --depth 1 "$REPO_URL" "$INSTALL_DIR"
  cd "$INSTALL_DIR"
  print_ok "克隆完成"
fi

cd "$INSTALL_DIR/deploy"

# ============================================================
# Step 4: 生成配置
# ============================================================
STEP=4
print_step "$STEP" "生成安全配置"

if [ -f ".env" ]; then
  print_warn ".env 已存在，保留现有配置"
  source .env
else
  cp .env.example .env

  # 随机生成密钥
  POSTGRES_PASSWORD=$(openssl rand -hex 32)
  JWT_SECRET=$(openssl rand -hex 32)
  TOTP_KEY=$(openssl rand -hex 32)

  # Linux 兼容 sed
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s/POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD=${POSTGRES_PASSWORD}/" .env
    sed -i '' "s/JWT_SECRET=.*/JWT_SECRET=${JWT_SECRET}/" .env
    sed -i '' "s/TOTP_ENCRYPTION_KEY=.*/TOTP_ENCRYPTION_KEY=${TOTP_KEY}/" .env
  else
    sed -i "s/POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD=${POSTGRES_PASSWORD}/" .env
    sed -i "s/JWT_SECRET=.*/JWT_SECRET=${JWT_SECRET}/" .env
    sed -i "s/TOTP_ENCRYPTION_KEY=.*/TOTP_ENCRYPTION_KEY=${TOTP_KEY}/" .env
  fi

  print_ok "密钥已生成: POSTGRES_PASSWORD / JWT_SECRET / TOTP_KEY"
fi

# ============================================================
# Step 5: 准备 docker-compose
# ============================================================
STEP=5
print_step "$STEP" "配置 Docker Compose（本地构建 + 签到抽奖）"

if [ -f "docker-compose.nj.yml" ]; then
  print_ok "docker-compose.nj.yml 已存在"
else
  # 从标准 docker-compose.yml 生成自定义版本
  # 关键改动：把 image 改成 build，实现本地构建
  cat > docker-compose.nj.yml << 'DOCKERCOMPOSE'
# =============================================================================
# Sub2API-NJ Docker Compose - 含签到抽奖系统
# =============================================================================
# 使用方式:
#   docker compose -f docker-compose.nj.yml --env-file .env up -d
# =============================================================================

services:
  sub2api:
    build:
      context: ..
      dockerfile: deploy/Dockerfile
      args:
        GOPROXY: https://goproxy.cn,direct
    container_name: sub2api-nj
    restart: unless-stopped
    ulimits:
      nofile:
        soft: 100000
        hard: 100000
    ports:
      - "${BIND_HOST:-0.0.0.0}:${SERVER_PORT:-8080}:8080"
    volumes:
      - sub2api_data:/app/data
    environment:
      - AUTO_SETUP=true
      - SERVER_HOST=0.0.0.0
      - SERVER_PORT=8080
      - SERVER_MODE=${SERVER_MODE:-release}
      - RUN_MODE=${RUN_MODE:-standard}
      - DATABASE_HOST=postgres
      - DATABASE_PORT=5432
      - DATABASE_USER=${POSTGRES_USER:-sub2api}
      - DATABASE_PASSWORD=${POSTGRES_PASSWORD}
      - DATABASE_DBNAME=${POSTGRES_DB:-sub2api}
      - DATABASE_SSLMODE=disable
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${REDIS_PASSWORD:-}
      - REDIS_DB=${REDIS_DB:-0}
      - ADMIN_EMAIL=${ADMIN_EMAIL:-admin@sub2api.local}
      - ADMIN_PASSWORD=${ADMIN_PASSWORD:-}
      - JWT_SECRET=${JWT_SECRET}
      - TOTP_ENCRYPTION_KEY=${TOTP_ENCRYPTION_KEY:-}
      - TZ=${TZ:-Asia/Shanghai}
      - SECURITY_URL_ALLOWLIST_ENABLED=${SECURITY_URL_ALLOWLIST_ENABLED:-false}
      - SECURITY_URL_ALLOWLIST_ALLOW_INSECURE_HTTP=${SECURITY_URL_ALLOWLIST_ALLOW_INSECURE_HTTP:-true}
      - SECURITY_URL_ALLOWLIST_ALLOW_PRIVATE_HOSTS=${SECURITY_URL_ALLOWLIST_ALLOW_PRIVATE_HOSTS:-true}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - sub2api-network
    healthcheck:
      test: ["CMD", "wget", "-q", "-T", "5", "-O", "/dev/null", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 60s

  postgres:
    image: postgres:18-alpine
    container_name: sub2api-nj-postgres
    restart: unless-stopped
    volumes:
      - postgres_data:/var/lib/postgresql/data
    environment:
      - PGDATA=/var/lib/postgresql/data
      - POSTGRES_USER=${POSTGRES_USER:-sub2api}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_DB=${POSTGRES_DB:-sub2api}
      - TZ=${TZ:-Asia/Shanghai}
    networks:
      - sub2api-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-sub2api} -d ${POSTGRES_DB:-sub2api}"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 10s

  redis:
    image: redis:8-alpine
    container_name: sub2api-nj-redis
    restart: unless-stopped
    volumes:
      - redis_data:/data
    command: >
      sh -c 'redis-server --save 60 1 --appendonly yes --appendfsync everysec ${REDIS_PASSWORD:+--requirepass "$REDIS_PASSWORD"}'
    environment:
      - TZ=${TZ:-Asia/Shanghai}
      - REDISCLI_AUTH=${REDIS_PASSWORD:-}
    networks:
      - sub2api-network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 5s

  # 迁移容器：启动后自动执行签到抽奖数据库迁移
  migrate:
    image: postgres:18-alpine
    container_name: sub2api-nj-migrate
    restart: "no"
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      - PGPASSWORD=${POSTGRES_PASSWORD}
    volumes:
      - ../backend/migrations:/migrations:ro
    command: >
      sh -c '
        echo "等待 PostgreSQL 就绪...";
        until pg_isready -h postgres -U ${POSTGRES_USER:-sub2api} -d ${POSTGRES_DB:-sub2api}; do sleep 2; done;
        echo "执行数据库迁移...";
        for f in /migrations/*.sql; do
          echo "  执行: $$(basename $$f)";
          psql -h postgres -U ${POSTGRES_USER:-sub2api} -d ${POSTGRES_DB:-sub2api} -f "$$f" 2>&1 || true;
        done;
        echo "迁移完成！";
      '
    networks:
      - sub2api-network

volumes:
  sub2api_data:
  postgres_data:
  redis_data:

networks:
  sub2api-network:
    driver: bridge
DOCKERCOMPOSE
  print_ok "docker-compose.nj.yml 已生成"
fi

# ============================================================
# Step 6: 构建并启动
# ============================================================
STEP=6
print_step "$STEP" "构建 Docker 镜像（约 5-10 分钟）"

print_info "拉取基础镜像..."
docker compose -f docker-compose.nj.yml --env-file .env pull postgres redis 2>&1 | tail -3

print_info "编译应用镜像（首次构建较慢）..."
docker compose -f docker-compose.nj.yml --env-file .env build sub2api 2>&1 | tail -5

print_info "启动所有服务..."
docker compose -f docker-compose.nj.yml --env-file .env up -d 2>&1 | tail -5

print_info "执行数据库迁移（签到抽奖表）..."
docker compose -f docker-compose.nj.yml --env-file .env up migrate 2>&1 | tail -5

print_info "等待应用就绪..."
for i in $(seq 1 30); do
  if curl -sf http://localhost:${SERVER_PORT}/health > /dev/null 2>&1; then
    print_ok "应用已就绪！"
    break
  fi
  sleep 2
done

# ============================================================
# Step 7: 输出部署信息
# ============================================================
STEP=7
print_step "$STEP" "部署完成"

# 获取服务器 IP
SERVER_IP=$(curl -sf https://api.ipify.org 2>/dev/null || curl -sf https://ipinfo.io/ip 2>/dev/null || echo "服务器IP")
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@sub2api.local}"
ADMIN_PASS="${ADMIN_PASSWORD:-（查看日志获取: docker logs sub2api-nj 2>&1 | grep ADMIN_PASSWORD）}"

echo ""
echo -e "${GREEN}================================================"
echo "  Sub2API-NJ 部署成功！"
echo "================================================"
echo -e "${NC}"
echo -e "  管理后台: ${CYAN}http://${SERVER_IP}:${SERVER_PORT}/admin${NC}"
echo -e "   账号: ${ADMIN_EMAIL}"
echo -e "   密码: ${ADMIN_PASS}"
echo ""
echo -e "  签到抽奖: ${CYAN}http://${SERVER_IP}:${SERVER_PORT}${NC} → 登录后访问"
echo ""
echo -e "  查看日志: ${YELLOW}docker logs -f sub2api-nj${NC}"
echo -e "  重启服务: ${YELLOW}docker compose -f ${INSTALL_DIR}/deploy/docker-compose.nj.yml --env-file ${INSTALL_DIR}/deploy/.env restart sub2api${NC}"
echo -e "  停止服务: ${YELLOW}docker compose -f ${INSTALL_DIR}/deploy/docker-compose.nj.yml --env-file ${INSTALL_DIR}/deploy/.env down${NC}"
echo ""
echo -e "  ${GREEN}签到抽奖功能预览:${NC}"
echo "    - 每日签到得积分（1-10分，连续3/7/30天额外加成）"
echo "    - 消耗积分抽奖（余额/积分返还/谢谢参与）"
echo "    - 管理后台配置奖品（权重/库存）"
echo -e "${GREEN}================================================"
echo -e "${NC}"
