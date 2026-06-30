-- 157_add_checkin_lottery.sql
-- 签到抽奖系统数据库迁移

-- 签到配置表
CREATE TABLE IF NOT EXISTS check_in_configs (
    id BIGSERIAL PRIMARY KEY,
    daily_min_points INTEGER NOT NULL DEFAULT 1,
    daily_max_points INTEGER NOT NULL DEFAULT 10,
    lottery_cost INTEGER NOT NULL DEFAULT 10,
    daily_max_draws INTEGER NOT NULL DEFAULT 5,
    consecutive_bonus_json TEXT NOT NULL DEFAULT '[]',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 插入默认配置
INSERT INTO check_in_configs (daily_min_points, daily_max_points, lottery_cost, daily_max_draws, consecutive_bonus_json, enabled)
VALUES (1, 10, 10, 5, '[{"days":3,"bonus":2},{"days":7,"bonus":5},{"days":30,"bonus":20}]', true);

-- 签到记录表
CREATE TABLE IF NOT EXISTS check_in_records (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    check_in_date DATE NOT NULL,
    points_earned INTEGER NOT NULL DEFAULT 0,
    consecutive_days INTEGER NOT NULL DEFAULT 1,
    total_points INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 每位用户每天只能签到一次
CREATE UNIQUE INDEX IF NOT EXISTS idx_check_in_records_user_date ON check_in_records(user_id, check_in_date);
CREATE INDEX IF NOT EXISTS idx_check_in_records_user_id ON check_in_records(user_id);
CREATE INDEX IF NOT EXISTS idx_check_in_records_date ON check_in_records(check_in_date);

-- 奖品配置表
CREATE TABLE IF NOT EXISTS lottery_prizes (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    prize_type VARCHAR(50) NOT NULL,
    amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    weight INTEGER NOT NULL DEFAULT 10,
    total_stock INTEGER NOT NULL DEFAULT -1,
    remaining_stock INTEGER NOT NULL DEFAULT -1,
    icon VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 插入默认奖品
INSERT INTO lottery_prizes (name, prize_type, amount, weight, total_stock, remaining_stock, icon, status, sort_order) VALUES
('谢谢参与', 'none', 0, 30, -1, -1, '', 'active', 0),
('1元余额', 'balance', 1, 25, -1, -1, '', 'active', 1),
('3元余额', 'balance', 3, 15, -1, -1, '', 'active', 2),
('5元余额', 'balance', 5, 10, -1, -1, '', 'active', 3),
('10元余额', 'balance', 10, 5, -1, -1, '', 'active', 4),
('积分返还5分', 'points', 5, 10, -1, -1, '', 'active', 5),
('积分返还10分', 'points', 10, 5, -1, -1, '', 'active', 6);

CREATE INDEX IF NOT EXISTS idx_lottery_prizes_status ON lottery_prizes(status);
CREATE INDEX IF NOT EXISTS idx_lottery_prizes_sort ON lottery_prizes(sort_order);

-- 抽奖记录表
CREATE TABLE IF NOT EXISTS lottery_records (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    prize_id BIGINT NOT NULL,
    prize_name VARCHAR(100) NOT NULL DEFAULT '',
    prize_type VARCHAR(50) NOT NULL DEFAULT '',
    amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    cost_points INTEGER NOT NULL DEFAULT 0,
    claimed BOOLEAN NOT NULL DEFAULT false,
    claimed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lottery_records_user_id ON lottery_records(user_id);
CREATE INDEX IF NOT EXISTS idx_lottery_records_prize_id ON lottery_records(prize_id);
CREATE INDEX IF NOT EXISTS idx_lottery_records_created_at ON lottery_records(created_at);
CREATE INDEX IF NOT EXISTS idx_lottery_records_claimed ON lottery_records(claimed);
