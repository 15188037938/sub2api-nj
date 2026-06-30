package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/lotteryrecord"
)

// CheckInRepo 签到数据访问接口
type CheckInRepo interface {
	GetCheckInConfig(ctx context.Context) (*ent.CheckInConfig, error)
	GetTodayCheckIn(ctx context.Context, userID int64, date string) (*ent.CheckInRecord, error)
	GetLastCheckIn(ctx context.Context, userID int64) (*ent.CheckInRecord, error)
	CreateCheckIn(ctx context.Context, userID int64, date time.Time, points, consecutiveDays, totalPoints int) (*ent.CheckInRecord, error)
	GetTodayDrawCount(ctx context.Context, userID int64, todayStart, todayEnd time.Time) (int, error)
	ListPrizes(ctx context.Context) ([]*ent.LotteryPrize, error)
	DecrementPrizeStock(ctx context.Context, id int64) error
	CreateLotteryRecord(ctx context.Context, userID, prizeID int64, prizeName, prizeType string, amount float64, costPoints int) (*ent.LotteryRecord, error)
	ClaimLotteryRecord(ctx context.Context, id int64) error
	GetLotteryRecords(ctx context.Context, userID int64, page, pageSize int) ([]*ent.LotteryRecord, int, error)
	GetCheckInRecords(ctx context.Context, userID int64, page, pageSize int) ([]*ent.CheckInRecord, int, error)
	UpdateCheckInConfig(ctx context.Context, id int, updates map[string]any) (*ent.CheckInConfig, error)
	CreatePrize(ctx context.Context, name, prizeType string, amount float64, weight, totalStock, sortOrder int, icon string) (*ent.LotteryPrize, error)
	UpdatePrize(ctx context.Context, id int64, updates map[string]any) (*ent.LotteryPrize, error)
	DeletePrize(ctx context.Context, id int64) error
	GetAllLotteryRecords(ctx context.Context, page, pageSize int) ([]*ent.LotteryRecord, int, error)
	GetAllCheckInRecords(ctx context.Context, page, pageSize int) ([]*ent.CheckInRecord, int, error)
}

// CheckInService 签到抽奖业务逻辑
type CheckInService struct {
	repo    CheckInRepo
	userRepo UserRepository
}

// NewCheckInService creates a new CheckInService
func NewCheckInService(repo CheckInRepo, userRepo UserRepository) *CheckInService {
	return &CheckInService{repo: repo, userRepo: userRepo}
}

// ConsecutiveBonus 连续签到加成规则
type ConsecutiveBonus struct {
	Days  int `json:"days"`
	Bonus int `json:"bonus"`
}

// CheckInStatus 签到状态响�?type CheckInStatus struct {
	CheckedIn       bool   `json:"checked_in"`
	ConsecutiveDays int    `json:"consecutive_days"`
	TotalPoints     int    `json:"total_points"`
	TodayPoints     int    `json:"today_points"`
	Enabled         bool   `json:"enabled"`
}

// LotteryResult 抽奖结果
type LotteryResult struct {
	ID         int64   `json:"id"`
	PrizeID    int64   `json:"prize_id"`
	PrizeName  string  `json:"prize_name"`
	PrizeType  string  `json:"prize_type"`
	Amount     float64 `json:"amount"`
	CostPoints int     `json:"cost_points"`
	Claimed    bool    `json:"claimed"`
	CreatedAt  string  `json:"created_at"`
}

// GetStatus 获取用户签到状�?func (s *CheckInService) GetStatus(ctx context.Context, userID int64) (*CheckInStatus, error) {
	config, err := s.repo.GetCheckInConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}

	today := time.Now().Format("2006-01-02")
	record, err := s.repo.GetTodayCheckIn(ctx, userID, today)
	checkedIn := err == nil && record != nil

	status := &CheckInStatus{
		CheckedIn: checkedIn,
		Enabled:   config.Enabled,
	}

	if checkedIn {
		status.ConsecutiveDays = record.ConsecutiveDays
		status.TotalPoints = record.TotalPoints
		status.TodayPoints = record.PointsEarned
	} else {
		// 获取上次签到记录来确定连续天�?		lastRecord, err := s.repo.GetLastCheckIn(ctx, userID)
		if err == nil {
			yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
			lastDate := lastRecord.CheckInDate.Format("2006-01-02")
			if lastDate == yesterday {
				status.ConsecutiveDays = lastRecord.ConsecutiveDays
			}
			status.TotalPoints = lastRecord.TotalPoints
		}
	}

	return status, nil
}

// DoCheckIn 执行签到
func (s *CheckInService) DoCheckIn(ctx context.Context, userID int64) (*CheckInStatus, error) {
	config, err := s.repo.GetCheckInConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("签到功能暂未开�?)
	}

	today := time.Now()
	todayStr := today.Format("2006-01-02")

	// 检查今日是否已签到
	existing, _ := s.repo.GetTodayCheckIn(ctx, userID, todayStr)
	if existing != nil {
		return nil, fmt.Errorf("今日已签�?)
	}

	// 计算连续签到天数
	consecutiveDays := 1
	lastRecord, err := s.repo.GetLastCheckIn(ctx, userID)
	if err == nil {
		yesterday := today.AddDate(0, 0, -1).Format("2006-01-02")
		lastDate := lastRecord.CheckInDate.Format("2006-01-02")
		if lastDate == yesterday {
			consecutiveDays = lastRecord.ConsecutiveDays + 1
		}
		// 如果上次签到不是昨天，连续天数重置为1
	}

	// 随机积分
	pointsRange := config.DailyMaxPoints - config.DailyMinPoints + 1
	points := config.DailyMinPoints + rand.Intn(pointsRange)

	// 连续签到加成
	var bonusConsecutive []ConsecutiveBonus
	if err := json.Unmarshal([]byte(config.ConsecutiveBonusJSON), &bonusConsecutive); err == nil {
		for _, b := range bonusConsecutive {
			if consecutiveDays >= b.Days {
				// 当天恰好到达里程碑天数时给加�?				if consecutiveDays == b.Days {
					points += b.Bonus
					slog.Info("checkin: consecutive bonus applied",
						"userID", userID,
						"consecutiveDays", consecutiveDays,
						"bonus", b.Bonus,
					)
				}
			}
		}
	}

	totalPoints := points
	if lastRecord != nil {
		totalPoints = lastRecord.TotalPoints + points
	}

	// 创建签到记录
	record, err := s.repo.CreateCheckIn(ctx, userID, today, points, consecutiveDays, totalPoints)
	if err != nil {
		return nil, fmt.Errorf("create checkin record: %w", err)
	}

	return &CheckInStatus{
		CheckedIn:       true,
		ConsecutiveDays: record.ConsecutiveDays,
		TotalPoints:     record.TotalPoints,
		TodayPoints:     record.PointsEarned,
		Enabled:         config.Enabled,
	}, nil
}

// DrawLottery 执行抽奖
func (s *CheckInService) DrawLottery(ctx context.Context, userID int64) (*LotteryResult, error) {
	config, err := s.repo.GetCheckInConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("抽奖功能暂未开�?)
	}

	// 检查今日已抽奖次数
	todayStart := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
	todayEnd := todayStart.Add(24 * time.Hour)
	drawCount, err := s.repo.GetTodayDrawCount(ctx, userID, todayStart, todayEnd)
	if err != nil {
		return nil, fmt.Errorf("get draw count: %w", err)
	}
	if drawCount >= config.DailyMaxDraws {
		return nil, fmt.Errorf("今日抽奖次数已用�?)
	}

	// 检查积分是否足�?	lastRecord, err := s.repo.GetLastCheckIn(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("请先签到获取积分")
	}
	if lastRecord.TotalPoints < config.LotteryCost {
		return nil, fmt.Errorf("积分不足，需�?d积分，当�?d积分", config.LotteryCost, lastRecord.TotalPoints)
	}

	// 获取奖品列表
	prizes, err := s.repo.ListPrizes(ctx)
	if err != nil || len(prizes) == 0 {
		return nil, fmt.Errorf("暂无可用奖品")
	}

	// 按权重抽取奖�?	totalWeight := 0
	availablePrizes := make([]*ent.LotteryPrize, 0)
	for _, p := range prizes {
		if p.RemainingStock != 0 && p.Status == "active" {
			totalWeight += p.Weight
			availablePrizes = append(availablePrizes, p)
		}
	}
	if len(availablePrizes) == 0 {
		return nil, fmt.Errorf("奖品已全部抽�?)
	}

	// 随机抽取
	r := rand.Intn(totalWeight)
	cumulative := 0
	var selectedPrize *ent.LotteryPrize
	for _, p := range availablePrizes {
		cumulative += p.Weight
		if r < cumulative {
			selectedPrize = p
			break
		}
	}
	if selectedPrize == nil {
		selectedPrize = availablePrizes[0]
	}

	// 扣库�?	if selectedPrize.RemainingStock > 0 {
		_ = s.repo.DecrementPrizeStock(ctx, selectedPrize.ID)
	}

	// 创建抽奖记录
	record, err := s.repo.CreateLotteryRecord(ctx, userID, selectedPrize.ID, selectedPrize.Name, selectedPrize.PrizeType, selectedPrize.Amount, config.LotteryCost)
	if err != nil {
		return nil, fmt.Errorf("create lottery record: %w", err)
	}

	// 处理奖品发放
	if selectedPrize.PrizeType != "none" && selectedPrize.Amount > 0 {
		s.processPrize(ctx, userID, selectedPrize.PrizeType, selectedPrize.Amount, record.ID)
	}

	return &LotteryResult{
		ID:         record.ID,
		PrizeID:    record.PrizeID,
		PrizeName:  record.PrizeName,
		PrizeType:  record.PrizeType,
		Amount:     record.Amount,
		CostPoints: record.CostPoints,
		Claimed:    record.Claimed,
		CreatedAt:  record.CreatedAt.Format(time.RFC3339),
	}, nil
}

// processPrize 处理奖品发放
func (s *CheckInService) processPrize(ctx context.Context, userID int64, prizeType string, amount float64, recordID int64) {
	switch prizeType {
	case "balance":
		// 发放余额
		err := s.userRepo.UpdateBalance(ctx, userID, amount)
		if err != nil {
			slog.Error("lottery: failed to add balance", "userID", userID, "amount", amount, "error", err)
		}
		// 自动标记为已领取
		_ = s.repo.ClaimLotteryRecord(ctx, recordID)
	case "points":
		// 积分返还已在上层处理
		_ = s.repo.ClaimLotteryRecord(ctx, recordID)
	case "none":
		// 谢谢参与，自动完�?		_ = s.repo.ClaimLotteryRecord(ctx, recordID)
	}
}

// GetLotteryHistory 获取用户抽奖记录
func (s *CheckInService) GetLotteryHistory(ctx context.Context, userID int64, page, pageSize int) ([]*LotteryResult, int, error) {
	records, total, err := s.repo.GetLotteryRecords(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	results := make([]*LotteryResult, len(records))
	for i, r := range records {
		results[i] = &LotteryResult{
			ID:         r.ID,
			PrizeID:    r.PrizeID,
			PrizeName:  r.PrizeName,
			PrizeType:  r.PrizeType,
			Amount:     r.Amount,
			CostPoints: r.CostPoints,
			Claimed:    r.Claimed,
			CreatedAt:  r.CreatedAt.Format(time.RFC3339),
		}
	}

	return results, total, nil
}

// GetTodayDrawCount 获取今日已抽次数
func (s *CheckInService) GetTodayDrawCount(ctx context.Context, userID int64) (int, int, error) {
	config, err := s.repo.GetCheckInConfig(ctx)
	if err != nil {
		return 0, 0, err
	}

	todayStart := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
	todayEnd := todayStart.Add(24 * time.Hour)
	count, err := s.repo.GetTodayDrawCount(ctx, userID, todayStart, todayEnd)
	if err != nil {
		return 0, 0, err
	}

	return count, config.DailyMaxDraws, nil
}

// GetCheckInRecords 获取用户签到记录
func (s *CheckInService) GetCheckInRecords(ctx context.Context, userID int64, page, pageSize int) ([]*ent.CheckInRecord, int, error) {
	return s.repo.GetCheckInRecords(ctx, userID, page, pageSize)
}

// ---------- 管理端接�?----------

// GetConfig 获取签到配置
func (s *CheckInService) GetConfig(ctx context.Context) (*ent.CheckInConfig, error) {
	return s.repo.GetCheckInConfig(ctx)
}

// UpdateConfig 更新签到配置
func (s *CheckInService) UpdateConfig(ctx context.Context, updates map[string]any) error {
	config, err := s.repo.GetCheckInConfig(ctx)
	if err != nil {
		return err
	}
	_, err = s.repo.UpdateCheckInConfig(ctx, config.ID, updates)
	return err
}

// ListPrizes 获取所有奖�?func (s *CheckInService) ListPrizes(ctx context.Context) ([]*ent.LotteryPrize, error) {
	return s.repo.ListPrizes(ctx)
}

// CreatePrize 创建奖品
func (s *CheckInService) CreatePrize(ctx context.Context, name, prizeType string, amount float64, weight, totalStock, sortOrder int, icon string) (*ent.LotteryPrize, error) {
	return s.repo.CreatePrize(ctx, name, prizeType, amount, weight, totalStock, sortOrder, icon)
}

// UpdatePrize 更新奖品
func (s *CheckInService) UpdatePrize(ctx context.Context, id int64, updates map[string]any) error {
	_, err := s.repo.UpdatePrize(ctx, id, updates)
	return err
}

// DeletePrize 删除奖品
func (s *CheckInService) DeletePrize(ctx context.Context, id int64) error {
	return s.repo.DeletePrize(ctx, id)
}

// GetAllLotteryRecords 获取所有抽奖记录（管理员）
func (s *CheckInService) GetAllLotteryRecords(ctx context.Context, page, pageSize int) ([]*ent.LotteryRecord, int, error) {
	return s.repo.GetAllLotteryRecords(ctx, page, pageSize)
}

// GetAllCheckInRecords 获取所有签到记录（管理员）
func (s *CheckInService) GetAllCheckInRecords(ctx context.Context, page, pageSize int) ([]*ent.CheckInRecord, int, error) {
	return s.repo.GetAllCheckInRecords(ctx, page, pageSize)
}

// GetLotteryStats 获取抽奖统计
func (s *CheckInService) GetLotteryStats(ctx context.Context) (map[string]any, error) {
	prizes, err := s.repo.ListPrizes(ctx)
	if err != nil {
		return nil, err
	}

	totalRecords, err := s.repo.GetAllLotteryRecords(ctx, 1, 1)
	if err != nil {
		return nil, err
	}

	stats := map[string]any{
		"total_prizes":  len(prizes),
		"total_records": totalRecords,
	}
	return stats, nil
}

// GetAllLotteryRecordsSlice 获取所有抽奖记录切片（管理员，不分页）
func (s *CheckInService) GetAllLotteryRecordsSlice(ctx context.Context) ([]*ent.LotteryRecord, error) {
	records, _, err := s.repo.GetAllLotteryRecords(ctx, 1, 10000)
	return records, err
}

// GetAllCheckInRecordsSlice 获取所有签到记录切片（管理员，不分页）
func (s *CheckInService) GetAllCheckInRecordsSlice(ctx context.Context) ([]*ent.CheckInRecord, error) {
	records, _, err := s.repo.GetAllCheckInRecords(ctx, 1, 10000)
	return records, err
}
