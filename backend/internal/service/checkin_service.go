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
// CheckInRepo defines the repository interface for check-in/lottery
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

)

// CheckInService 绛惧埌鎶藉涓氬姟閫昏緫
type CheckInService struct {
	repo    CheckInRepo
	userRepo UserRepository
}

// NewCheckInService creates a new CheckInService
func NewCheckInService(repo *repository.CheckInRepo, userRepo UserRepository) *CheckInService {
	return &CheckInService{repo: repo, userRepo: userRepo}
}

// ConsecutiveBonus 杩炵画绛惧埌鍔犳垚瑙勫垯
type ConsecutiveBonus struct {
	Days  int `json:"days"`
	Bonus int `json:"bonus"`
}

// CheckInStatus 绛惧埌鐘舵€佸搷搴?type CheckInStatus struct {
	CheckedIn       bool   `json:"checked_in"`
	ConsecutiveDays int    `json:"consecutive_days"`
	TotalPoints     int    `json:"total_points"`
	TodayPoints     int    `json:"today_points"`
	Enabled         bool   `json:"enabled"`
}

// LotteryResult 鎶藉缁撴灉
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

// GetStatus 鑾峰彇鐢ㄦ埛绛惧埌鐘舵€?func (s *CheckInService) GetStatus(ctx context.Context, userID int64) (*CheckInStatus, error) {
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
		// 鑾峰彇涓婃绛惧埌璁板綍鏉ョ‘瀹氳繛缁ぉ鏁?		lastRecord, err := s.repo.GetLastCheckIn(ctx, userID)
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

// DoCheckIn 鎵ц绛惧埌
func (s *CheckInService) DoCheckIn(ctx context.Context, userID int64) (*CheckInStatus, error) {
	config, err := s.repo.GetCheckInConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("绛惧埌鍔熻兘鏆傛湭寮€鏀?)
	}

	today := time.Now()
	todayStr := today.Format("2006-01-02")

	// 妫€鏌ヤ粖鏃ユ槸鍚﹀凡绛惧埌
	existing, _ := s.repo.GetTodayCheckIn(ctx, userID, todayStr)
	if existing != nil {
		return nil, fmt.Errorf("浠婃棩宸茬鍒?)
	}

	// 璁＄畻杩炵画绛惧埌澶╂暟
	consecutiveDays := 1
	lastRecord, err := s.repo.GetLastCheckIn(ctx, userID)
	if err == nil {
		yesterday := today.AddDate(0, 0, -1).Format("2006-01-02")
		lastDate := lastRecord.CheckInDate.Format("2006-01-02")
		if lastDate == yesterday {
			consecutiveDays = lastRecord.ConsecutiveDays + 1
		}
		// 濡傛灉涓婃绛惧埌涓嶆槸鏄ㄥぉ锛岃繛缁ぉ鏁伴噸缃负1
	}

	// 闅忔満绉垎
	pointsRange := config.DailyMaxPoints - config.DailyMinPoints + 1
	points := config.DailyMinPoints + rand.Intn(pointsRange)

	// 杩炵画绛惧埌鍔犳垚
	var bonusConsecutive []ConsecutiveBonus
	if err := json.Unmarshal([]byte(config.ConsecutiveBonusJSON), &bonusConsecutive); err == nil {
		for _, b := range bonusConsecutive {
			if consecutiveDays >= b.Days {
				// 褰撳ぉ鎭板ソ鍒拌揪閲岀▼纰戝ぉ鏁版椂缁欏姞鎴?				if consecutiveDays == b.Days {
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

	// 鍒涘缓绛惧埌璁板綍
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

// DrawLottery 鎵ц鎶藉
func (s *CheckInService) DrawLottery(ctx context.Context, userID int64) (*LotteryResult, error) {
	config, err := s.repo.GetCheckInConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("鎶藉鍔熻兘鏆傛湭寮€鏀?)
	}

	// 妫€鏌ヤ粖鏃ュ凡鎶藉娆℃暟
	todayStart := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
	todayEnd := todayStart.Add(24 * time.Hour)
	drawCount, err := s.repo.GetTodayDrawCount(ctx, userID, todayStart, todayEnd)
	if err != nil {
		return nil, fmt.Errorf("get draw count: %w", err)
	}
	if drawCount >= config.DailyMaxDraws {
		return nil, fmt.Errorf("浠婃棩鎶藉娆℃暟宸茬敤瀹?)
	}

	// 妫€鏌ョН鍒嗘槸鍚﹁冻澶?	lastRecord, err := s.repo.GetLastCheckIn(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("璇峰厛绛惧埌鑾峰彇绉垎")
	}
	if lastRecord.TotalPoints < config.LotteryCost {
		return nil, fmt.Errorf("绉垎涓嶈冻锛岄渶瑕?d绉垎锛屽綋鍓?d绉垎", config.LotteryCost, lastRecord.TotalPoints)
	}

	// 鑾峰彇濂栧搧鍒楄〃
	prizes, err := s.repo.ListPrizes(ctx)
	if err != nil || len(prizes) == 0 {
		return nil, fmt.Errorf("鏆傛棤鍙敤濂栧搧")
	}

	// 鎸夋潈閲嶆娊鍙栧鍝?	totalWeight := 0
	availablePrizes := make([]*ent.LotteryPrize, 0)
	for _, p := range prizes {
		if p.RemainingStock != 0 && p.Status == "active" {
			totalWeight += p.Weight
			availablePrizes = append(availablePrizes, p)
		}
	}
	if len(availablePrizes) == 0 {
		return nil, fmt.Errorf("濂栧搧宸插叏閮ㄦ娊瀹?)
	}

	// 闅忔満鎶藉彇
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

	// 鎵ｅ簱瀛?	if selectedPrize.RemainingStock > 0 {
		_ = s.repo.DecrementPrizeStock(ctx, selectedPrize.ID)
	}

	// 鍒涘缓鎶藉璁板綍
	record, err := s.repo.CreateLotteryRecord(ctx, userID, selectedPrize.ID, selectedPrize.Name, selectedPrize.PrizeType, selectedPrize.Amount, config.LotteryCost)
	if err != nil {
		return nil, fmt.Errorf("create lottery record: %w", err)
	}

	// 澶勭悊濂栧搧鍙戞斁
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

// processPrize 澶勭悊濂栧搧鍙戞斁
func (s *CheckInService) processPrize(ctx context.Context, userID int64, prizeType string, amount float64, recordID int64) {
	switch prizeType {
	case "balance":
		// 鍙戞斁浣欓
		err := s.userRepo.UpdateBalance(ctx, userID, amount, "+", fmt.Sprintf("鎶藉涓 #%d", recordID))
		if err != nil {
			slog.Error("lottery: failed to add balance", "userID", userID, "amount", amount, "error", err)
		}
		// 鑷姩鏍囪涓哄凡棰嗗彇
		_ = s.repo.ClaimLotteryRecord(ctx, recordID)
	case "points":
		// 绉垎杩旇繕宸插湪涓婂眰澶勭悊
		_ = s.repo.ClaimLotteryRecord(ctx, recordID)
	case "none":
		// 璋㈣阿鍙備笌锛岃嚜鍔ㄥ畬鎴?		_ = s.repo.ClaimLotteryRecord(ctx, recordID)
	}
}

// GetLotteryHistory 鑾峰彇鐢ㄦ埛鎶藉璁板綍
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

// GetTodayDrawCount 鑾峰彇浠婃棩宸叉娊娆℃暟
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

// GetCheckInRecords 鑾峰彇鐢ㄦ埛绛惧埌璁板綍
func (s *CheckInService) GetCheckInRecords(ctx context.Context, userID int64, page, pageSize int) ([]*ent.CheckInRecord, int, error) {
	return s.repo.GetCheckInRecords(ctx, userID, page, pageSize)
}

// ---------- 绠＄悊绔帴鍙?----------

// GetConfig 鑾峰彇绛惧埌閰嶇疆
func (s *CheckInService) GetConfig(ctx context.Context) (*ent.CheckInConfig, error) {
	return s.repo.GetCheckInConfig(ctx)
}

// UpdateConfig 鏇存柊绛惧埌閰嶇疆
func (s *CheckInService) UpdateConfig(ctx context.Context, updates map[string]any) error {
	config, err := s.repo.GetCheckInConfig(ctx)
	if err != nil {
		return err
	}
	_, err = s.repo.UpdateCheckInConfig(ctx, config.ID, updates)
	return err
}

// ListPrizes 鑾峰彇鎵€鏈夊鍝?func (s *CheckInService) ListPrizes(ctx context.Context) ([]*ent.LotteryPrize, error) {
	return s.repo.ListPrizes(ctx)
}

// CreatePrize 鍒涘缓濂栧搧
func (s *CheckInService) CreatePrize(ctx context.Context, name, prizeType string, amount float64, weight, totalStock, sortOrder int, icon string) (*ent.LotteryPrize, error) {
	return s.repo.CreatePrize(ctx, name, prizeType, amount, weight, totalStock, sortOrder, icon)
}

// UpdatePrize 鏇存柊濂栧搧
func (s *CheckInService) UpdatePrize(ctx context.Context, id int64, updates map[string]any) error {
	_, err := s.repo.UpdatePrize(ctx, id, updates)
	return err
}

// DeletePrize 鍒犻櫎濂栧搧
func (s *CheckInService) DeletePrize(ctx context.Context, id int64) error {
	return s.repo.DeletePrize(ctx, id)
}

// GetAllLotteryRecords 鑾峰彇鎵€鏈夋娊濂栬褰曪紙绠＄悊鍛橈級
func (s *CheckInService) GetAllLotteryRecords(ctx context.Context, page, pageSize int) ([]*ent.LotteryRecord, int, error) {
	return s.repo.GetAllLotteryRecords(ctx, page, pageSize)
}

// GetAllCheckInRecords 鑾峰彇鎵€鏈夌鍒拌褰曪紙绠＄悊鍛橈級
func (s *CheckInService) GetAllCheckInRecords(ctx context.Context, page, pageSize int) ([]*ent.CheckInRecord, int, error) {
	return s.repo.GetAllCheckInRecords(ctx, page, pageSize)
}

// GetLotteryStats 鑾峰彇鎶藉缁熻
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

// GetAllLotteryRecordsSlice 鑾峰彇鎵€鏈夋娊濂栬褰曞垏鐗囷紙绠＄悊鍛橈紝涓嶅垎椤碉級
func (s *CheckInService) GetAllLotteryRecordsSlice(ctx context.Context) ([]*ent.LotteryRecord, error) {
	records, _, err := s.repo.GetAllLotteryRecords(ctx, 1, 10000)
	return records, err
}

// GetAllCheckInRecordsSlice 鑾峰彇鎵€鏈夌鍒拌褰曞垏鐗囷紙绠＄悊鍛橈紝涓嶅垎椤碉級
func (s *CheckInService) GetAllCheckInRecordsSlice(ctx context.Context) ([]*ent.CheckInRecord, error) {
	records, _, err := s.repo.GetAllCheckInRecords(ctx, 1, 10000)
	return records, err
}
