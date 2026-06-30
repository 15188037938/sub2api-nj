# Sub2API-NJ 部署指南

含签到抽奖系统的 Sub2API 自定义版本。

**仓库:** <https://github.com/15188037938/sub2api-nj>

## 一键安装（推荐）

在**全新 Linux 服务器**上，一条命令搞定：

```bash
# 方式 A: curl
curl -sSL https://raw.githubusercontent.com/15188037938/sub2api-nj/main/deploy/quick-install.sh | sudo bash

# 方式 B: wget
wget -qO- https://raw.githubusercontent.com/15188037938/sub2api-nj/main/deploy/quick-install.sh | sudo bash
```

脚本会自动完成：
- 安装 Docker + Docker Compose（如未安装）
- 克隆项目代码
- 生成随机安全密钥（数据库密码、JWT 密钥、TOTP 密钥）
- 使用 Docker 本地构建完全镜像（含签到抽奖代码）
- 启动 PostgreSQL、Redis、应用服务
- 执行数据库迁移（创建签到/抽奖相关表）
- 输出访问地址和管理后台账号

> 首次构建约 5-10 分钟，后续更新只需改代码后重新 `docker compose build`。

## 手动部署

### 前置条件

| 组件 | 版本 | 用途 |
|------|------|------|
| Docker | 24+ | 容器化运行 |
| Docker Compose | 2.20+ | 服务编排 |
| 或 Go | 1.21+ | 直接编译运行 |

### 方案 A: Docker Compose

```bash
# 1. 克隆仓库
git clone --depth 1 https://github.com/15188037938/sub2api-nj.git
cd sub2api-nj/deploy

# 2. 配置环境变量
cp .env.example .env
# 编辑 .env，至少修改：
#   POSTGRES_PASSWORD=your_secure_password
#   JWT_SECRET=$(openssl rand -hex 32)

# 3. 启动（本地构建，100% 含签到抽奖代码）
docker compose -f docker-compose.nj.yml --env-file .env up -d

# 4. 执行数据库迁移
docker compose -f docker-compose.nj.yml --env-file .env up migrate
```

### 方案 B: 直接编译

```bash
# 1. 安装依赖
#    Go 1.21+, Node.js 18+, pnpm

# 2. 克隆仓库
git clone https://github.com/15188037938/sub2api-nj.git
cd sub2api-nj/deploy

# 3. 设置 PostgreSQL 和 Redis（自行安装）
#    确保数据库可达

# 4. 编译后端（嵌入前端）
make build-embed

# 5. 配置
cp config.example.yaml config.yaml
# 编辑 config.yaml: 填写数据库连接、Redis地址、JWT_SECRET

# 6. 执行迁移
psql -U sub2api -d sub2api -f ../backend/migrations/157_add_checkin_lottery.sql

# 7. 启动
./bin/server --config config.yaml
```

## 数据库迁移

签到抽奖系统新增 4 张表。自动迁移已在 `docker-compose.nj.yml` 中集成 - `migrate` 容器会按文件名顺序执行 `migrations/` 目录下的所有 SQL。

如需手动执行：

```bash
export PGPASSWORD=your_password
psql -h localhost -U sub2api -d sub2api -f backend/migrations/157_add_checkin_lottery.sql
```

迁移内容：`checkin_configs` / `checkin_records` / `lottery_prizes` / `lottery_records` + 7 条默认奖品数据。

## 功能预览

### 签到系统
| 功能 | 默认值 | 管理后台可配置 |
|------|--------|---------------|
| 每日积分 | 1-10 分随机 | 是 |
| 连续 3 天加成 | +2 分 | 是 |
| 连续 7 天加成 | +5 分 | 是 |
| 连续 30 天加成 | +20 分 | 是 |

### 抽奖系统
| 奖品 | 权重 | 说明 |
|------|------|------|
| 谢谢参与 | 30 | 概率最高 |
| 1 元余额 | 25 | 自动到账 |
| 3 元余额 | 15 | 自动到账 |
| 5 元余额 | 10 | 自动到账 |
| 10 元余额 | 5 | 自动到账 |
| 积分返还+5 | 10 | 返还 5 积分 |
| 积分返还+10 | 5 | 返还 10 积分 |

### 管理后台
- 签到配置：积分范围、抽奖消耗、每日上限、启用开关
- 奖品管理：CRUD、权重、库存、排序、状态

## API 端点

路径 | 方法 | 权限 | 说明
-----|------|------|-----
`/user/checkin/status` | GET | 用户 | 签到状态
`/user/checkin` | POST | 用户 | 执行签到
`/user/checkin/records` | GET | 用户 | 签到记录
`/user/lottery/config` | GET | 用户 | 奖品列表
`/user/lottery/draw` | POST | 用户 | 抽奖
`/user/lottery/history` | GET | 用户 | 抽奖记录
`/admin/checkin/config` | GET/PUT | 管理员 | 签到配置
`/admin/lottery/prizes` | GET/POST | 管理员 | 奖品管理
`/admin/lottery/prizes/:id` | PUT/DELETE | 管理员 | 奖品编辑
`/admin/lottery/records` | GET | 管理员 | 抽奖记录
`/admin/lottery/stats` | GET | 管理员 | 抽奖统计

## 升级更新

```bash
cd /opt/sub2api-nj
git pull origin main
cd deploy
docker compose -f docker-compose.nj.yml --env-file .env build sub2api
docker compose -f docker-compose.nj.yml --env-file .env up -d
```
