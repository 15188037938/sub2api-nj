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

// CheckInRepo 绛惧埌鏁版嵁璁块棶灞?type CheckInRepo struct {
	client *ent.Client
}

// NewCheckInRepo creates a new CheckInRepo
func NewCheckInRepo(client *ent.Client) *CheckInRepo {
	return &CheckInRepo{client: client}
}

// GetTodayCheckIn 鏌ヨ鐢ㄦ埛浠婃棩鏄惁宸茬鍒?func (r *CheckInRepo) GetTodayCheckIn(ctx context.Context, userID int64, date string) (*ent.CheckInRecord, error) {
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

// CreateCheckIn 鍒涘缓绛惧埌璁板綍
func (r *CheckInRepo) CreateCheckIn(ctx context.Context, userID int64, date time.Time, points, consecutiveDays, totalPoints int) (*ent.CheckInRecord, error) {
	return r.client.CheckInRecord.Create().
		SetUserID(userID).
		SetCheckInDate(date).
		SetPointsEarned(points).
		SetConsecutiveDays(consecutiveDays).
		SetTotalPoints(totalPoints).
		Save(ctx)
}

// GetLastCheckIn 鑾峰彇鐢ㄦ埛鏈€杩戜竴娆＄鍒拌褰曪紙鍒ゆ柇杩炵画澶╂暟锛?func (r *CheckInRepo) GetLastCheckIn(ctx context.Context, userID int64) (*ent.CheckInRecord, error) {
	return r.client.CheckInRecord.Query().
		Where(checkinrecord.UserIDEQ(userID)).
		Order(ent.Desc(checkinrecord.FieldCheckInDate)).
		First(ctx)
}

// GetCheckInRecords 鍒嗛〉鏌ヨ绛惧埌璁板綍
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

// GetCheckInConfig 鑾峰彇绛惧埌閰嶇疆
func (r *CheckInRepo) GetCheckInConfig(ctx context.Context) (*ent.CheckInConfig, error) {
	return r.client.CheckInConfig.Query().First(ctx)
}

// UpdateCheckInConfig 鏇存柊绛惧埌閰嶇疆
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

// ListPrizes 鑾峰彇鎵€鏈夋椿璺冨鍝侊紙鎸夋潈閲嶆帓搴忥級
func (r *CheckInRepo) ListPrizes(ctx context.Context) ([]*ent.LotteryPrize, error) {
	return r.client.LotteryPrize.Query().
		Where(lotteryprize.StatusEQ("active")).
		Order(ent.Asc(lotteryprize.FieldSortOrder)).
		All(ctx)
}

// GetPrize 鑾峰彇鍗曚釜濂栧搧
func (r *CheckInRepo) GetPrize(ctx context.Context, id int64) (*ent.LotteryPrize, error) {
	return r.client.LotteryPrize.Get(ctx, id)
}

// CreatePrize 鍒涘缓濂栧搧
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

// UpdatePrize 鏇存柊濂栧搧
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

// DeletePrize 鍒犻櫎濂栧搧
func (r *CheckInRepo) DeletePrize(ctx context.Context, id int64) error {
	return r.client.LotteryPrize.DeleteOneID(id).Exec(ctx)
}

// DecrementPrizeStock 鍑忓皯濂栧搧搴撳瓨
func (r *CheckInRepo) DecrementPrizeStock(ctx context.Context, id int64) error {
	return r.client.LotteryPrize.UpdateOneID(id).
		AddRemainingStock(-1).
		Exec(ctx)
}

// CreateLotteryRecord 鍒涘缓鎶藉璁板綍
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

// ClaimLotteryRecord 棰嗗彇濂栧搧
func (r *CheckInRepo) ClaimLotteryRecord(ctx context.Context, id int64) error {
	now := time.Now()
	return r.client.LotteryRecord.UpdateOneID(id).
		SetClaimed(true).
		SetClaimedAt(now).
		Exec(ctx)
}

// GetTodayDrawCount 鑾峰彇鐢ㄦ埛浠婃棩鎶藉娆℃暟
func (r *CheckInRepo) GetTodayDrawCount(ctx context.Context, userID int64, todayStart, todayEnd time.Time) (int, error) {
	return r.client.LotteryRecord.Query().
		Where(
			lotteryrecord.UserIDEQ(userID),
			lotteryrecord.CreatedAtGTE(todayStart),
			lotteryrecord.CreatedAtLTE(todayEnd),
		).
		Count(ctx)
}

// GetLotteryRecords 鍒嗛〉鏌ヨ鎶藉璁板綍
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

// GetAllLotteryRecords 绠＄悊鍛樻煡璇㈡墍鏈夋娊濂栬褰?func (r *CheckInRepo) GetAllLotteryRecords(ctx context.Context, page, pageSize int) ([]*ent.LotteryRecord, int, error) {
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

// GetAllCheckInRecords 绠＄悊鍛樻煡璇㈡墍鏈夌鍒拌褰?func (r *CheckInRepo) GetAllCheckInRecords(ctx context.Context, page, pageSize int) ([]*ent.CheckInRecord, int, error) {
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
