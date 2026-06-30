package repository

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/checkinrecord"
	"github.com/Wei-Shaw/sub2api/ent/lotteryrecord"
	"github.com/Wei-Shaw/sub2api/ent/lotteryprize"
	"github.com/Wei-Shaw/sub2api/ent/checkinconfig"
)

// CheckInRepo 签到数据访问�?type CheckInRepo struct {
	client *ent.Client
}

// NewCheckInRepo creates a new CheckInRepo
func NewCheckInRepo(client *ent.Client) *CheckInRepo {
	return &CheckInRepo{client: client}
}

// GetTodayCheckIn 查询用户今日是否已签�?func (r *CheckInRepo) GetTodayCheckIn(ctx context.Context, userID int64, date string) (*ent.CheckInRecord, error) {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, err
	}
	return r.client.CheckInRecord.Query().
		Where(
			checkinrecord.UserIDEQ(userID),
			checkinrecord.CheckInDateEQ(d),
		).
		Only(ctx)
}

// CreateCheckIn 创建签到记录
func (r *CheckInRepo) CreateCheckIn(ctx context.Context, userID int64, date time.Time, points, consecutiveDays, totalPoints int) (*ent.CheckInRecord, error) {
	return r.client.CheckInRecord.Create().
		SetUserID(userID).
		SetCheckInDate(date).
		SetPointsEarned(points).
		SetConsecutiveDays(consecutiveDays).
		SetTotalPoints(totalPoints).
		Save(ctx)
}

// GetLastCheckIn 获取用户最近一次签到记录（判断连续天数�?func (r *CheckInRepo) GetLastCheckIn(ctx context.Context, userID int64) (*ent.CheckInRecord, error) {
	return r.client.CheckInRecord.Query().
		Where(checkinrecord.UserIDEQ(userID)).
		Order(ent.Desc(checkinrecord.FieldCheckInDate)).
		First(ctx)
}

// GetCheckInRecords 分页查询签到记录
func (r *CheckInRepo) GetCheckInRecords(ctx context.Context, userID int64, page, pageSize int) ([]*ent.CheckInRecord, int, error) {
	query := r.client.CheckInRecord.Query().
		Where(checkinrecord.UserIDEQ(userID))

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	records, err := query.
		Order(ent.Desc(checkinrecord.FieldCreatedAt)).
		Offset(offset).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// GetCheckInConfig 获取签到配置
func (r *CheckInRepo) GetCheckInConfig(ctx context.Context) (*ent.CheckInConfig, error) {
	return r.client.CheckInConfig.Query().First(ctx)
}

// UpdateCheckInConfig 更新签到配置
func (r *CheckInRepo) UpdateCheckInConfig(ctx context.Context, id int, updates map[string]any) (*ent.CheckInConfig, error) {
	update := r.client.CheckInConfig.UpdateOneID(id)
	for k, v := range updates {
		switch k {
		case "daily_min_points":
			update.SetDailyMinPoints(v.(int))
		case "daily_max_points":
			update.SetDailyMaxPoints(v.(int))
		case "lottery_cost":
			update.SetLotteryCost(v.(int))
		case "daily_max_draws":
			update.SetDailyMaxDraws(v.(int))
		case "consecutive_bonus_json":
			update.SetConsecutiveBonusJSON(v.(string))
		case "enabled":
			update.SetEnabled(v.(bool))
		}
	}
	return update.Save(ctx)
}

// ListPrizes 获取所有活跃奖品（按权重排序）
func (r *CheckInRepo) ListPrizes(ctx context.Context) ([]*ent.LotteryPrize, error) {
	return r.client.LotteryPrize.Query().
		Where(lotteryprize.StatusEQ("active")).
		Order(ent.Asc(lotteryprize.FieldSortOrder)).
		All(ctx)
}

// GetPrize 获取单个奖品
func (r *CheckInRepo) GetPrize(ctx context.Context, id int64) (*ent.LotteryPrize, error) {
	return r.client.LotteryPrize.Get(ctx, id)
}

// CreatePrize 创建奖品
func (r *CheckInRepo) CreatePrize(ctx context.Context, name, prizeType string, amount float64, weight, totalStock, sortOrder int, icon string) (*ent.LotteryPrize, error) {
	remaining := totalStock
	return r.client.LotteryPrize.Create().
		SetName(name).
		SetPrizeType(prizeType).
		SetAmount(amount).
		SetWeight(weight).
		SetTotalStock(totalStock).
		SetRemainingStock(remaining).
		SetSortOrder(sortOrder).
		SetIcon(icon).
		Save(ctx)
}

// UpdatePrize 更新奖品
func (r *CheckInRepo) UpdatePrize(ctx context.Context, id int64, updates map[string]any) (*ent.LotteryPrize, error) {
	update := r.client.LotteryPrize.UpdateOneID(id)
	for k, v := range updates {
		switch k {
		case "name":
			update.SetName(v.(string))
		case "prize_type":
			update.SetPrizeType(v.(string))
		case "amount":
			update.SetAmount(v.(float64))
		case "weight":
			update.SetWeight(v.(int))
		case "total_stock":
			update.SetTotalStock(v.(int))
		case "remaining_stock":
			update.SetRemainingStock(v.(int))
		case "sort_order":
			update.SetSortOrder(v.(int))
		case "icon":
			update.SetIcon(v.(string))
		case "status":
			update.SetStatus(v.(string))
		}
	}
	return update.Save(ctx)
}

// DeletePrize 删除奖品
func (r *CheckInRepo) DeletePrize(ctx context.Context, id int64) error {
	return r.client.LotteryPrize.DeleteOneID(id).Exec(ctx)
}

// DecrementPrizeStock 减少奖品库存
func (r *CheckInRepo) DecrementPrizeStock(ctx context.Context, id int64) error {
	return r.client.LotteryPrize.UpdateOneID(id).
		AddRemainingStock(-1).
		Exec(ctx)
}

// CreateLotteryRecord 创建抽奖记录
func (r *CheckInRepo) CreateLotteryRecord(ctx context.Context, userID, prizeID int64, prizeName, prizeType string, amount float64, costPoints int) (*ent.LotteryRecord, error) {
	return r.client.LotteryRecord.Create().
		SetUserID(userID).
		SetPrizeID(prizeID).
		SetPrizeName(prizeName).
		SetPrizeType(prizeType).
		SetAmount(amount).
		SetCostPoints(costPoints).
		Save(ctx)
}

// ClaimLotteryRecord 领取奖品
func (r *CheckInRepo) ClaimLotteryRecord(ctx context.Context, id int64) error {
	now := time.Now()
	return r.client.LotteryRecord.UpdateOneID(id).
		SetClaimed(true).
		SetClaimedAt(now).
		Exec(ctx)
}

// GetTodayDrawCount 获取用户今日抽奖次数
func (r *CheckInRepo) GetTodayDrawCount(ctx context.Context, userID int64, todayStart, todayEnd time.Time) (int, error) {
	return r.client.LotteryRecord.Query().
		Where(
			lotteryrecord.UserIDEQ(userID),
			lotteryrecord.CreatedAtGTE(todayStart),
			lotteryrecord.CreatedAtLTE(todayEnd),
		).
		Count(ctx)
}

// GetLotteryRecords 分页查询抽奖记录
func (r *CheckInRepo) GetLotteryRecords(ctx context.Context, userID int64, page, pageSize int) ([]*ent.LotteryRecord, int, error) {
	query := r.client.LotteryRecord.Query().
		Where(lotteryrecord.UserIDEQ(userID))

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	records, err := query.
		Order(ent.Desc(lotteryrecord.FieldCreatedAt)).
		Offset(offset).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// GetAllLotteryRecords 管理员查询所有抽奖记�?func (r *CheckInRepo) GetAllLotteryRecords(ctx context.Context, page, pageSize int) ([]*ent.LotteryRecord, int, error) {
	query := r.client.LotteryRecord.Query()

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	records, err := query.
		Order(ent.Desc(lotteryrecord.FieldCreatedAt)).
		Offset(offset).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// GetAllCheckInRecords 管理员查询所有签到记�?func (r *CheckInRepo) GetAllCheckInRecords(ctx context.Context, page, pageSize int) ([]*ent.CheckInRecord, int, error) {
	query := r.client.CheckInRecord.Query()

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	records, err := query.
		Order(ent.Desc(checkinrecord.FieldCreatedAt)).
		Offset(offset).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}
