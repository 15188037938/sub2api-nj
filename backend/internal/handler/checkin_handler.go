package handler

import (
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// CheckInHandler 签到抽奖 HTTP 处理器（用户端）
type CheckInHandler struct {
	checkInService *service.CheckInService
}

// NewCheckInHandler creates a new CheckInHandler
func NewCheckInHandler(checkInService *service.CheckInService) *CheckInHandler {
	return &CheckInHandler{checkInService: checkInService}
}

// GetStatus 获取签到状态
func (h *CheckInHandler) GetStatus(c *gin.Context) {
	userID := response.GetUserID(c)

	status, err := h.checkInService.GetStatus(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, status)
}

// DoCheckIn 执行签到
func (h *CheckInHandler) DoCheckIn(c *gin.Context) {
	userID := response.GetUserID(c)

	status, err := h.checkInService.DoCheckIn(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, status)
}

// DrawLottery 抽奖
func (h *CheckInHandler) DrawLottery(c *gin.Context) {
	userID := response.GetUserID(c)

	result, err := h.checkInService.DrawLottery(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, result)
}

// GetLotteryHistory 获取抽奖记录
func (h *CheckInHandler) GetLotteryHistory(c *gin.Context) {
	userID := response.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	records, total, err := h.checkInService.GetLotteryHistory(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"total":   total,
		"records": records,
	})
}

// GetLotteryConfig 获取抽奖配置（奖品列表）
func (h *CheckInHandler) GetLotteryConfig(c *gin.Context) {
	config, err := h.checkInService.GetConfig(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	prizes, err := h.checkInService.ListPrizes(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	count, maxDraws, _ := h.checkInService.GetTodayDrawCount(c.Request.Context(), response.GetUserID(c))

	response.Success(c, gin.H{
		"config":         config,
		"prizes":         prizes,
		"today_draws":    count,
		"daily_max_draws": maxDraws,
		"lottery_cost":   config.LotteryCost,
	})
}

// GetCheckInRecords 获取签到记录
func (h *CheckInHandler) GetCheckInRecords(c *gin.Context) {
	userID := response.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	records, total, err := h.checkInService.GetCheckInRecords(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"total":   total,
		"records": records,
	})
}

// ClaimLotteryReward 领取奖品
func (h *CheckInHandler) ClaimLotteryReward(c *gin.Context) {
	// 奖品已在抽中时自动发放（balance/points类型），此处仅做状态更新
	recordID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的记录ID")
		return
	}

	_ = recordID // 后续扩展：标记为已领取
	response.Success(c, gin.H{"message": "领取成功", "record_id": recordID})
}

// AdminCheckInHandler 签到抽奖管理端 Handler
type AdminCheckInHandler struct {
	checkInService *service.CheckInService
}

// NewAdminCheckInHandler creates a new AdminCheckInHandler
func NewAdminCheckInHandler(checkInService *service.CheckInService) *AdminCheckInHandler {
	return &AdminCheckInHandler{checkInService: checkInService}
}

// GetConfig 获取签到配置
func (h *AdminCheckInHandler) GetConfig(c *gin.Context) {
	config, err := h.checkInService.GetConfig(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, config)
}

// UpdateConfig 更新签到配置
func (h *AdminCheckInHandler) UpdateConfig(c *gin.Context) {
	var updates map[string]any
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.Error(c, http.StatusBadRequest, "参数格式错误")
		return
	}

	if err := h.checkInService.UpdateConfig(c.Request.Context(), updates); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	config, _ := h.checkInService.GetConfig(c.Request.Context())
	response.Success(c, config)
}

// ListPrizes 获取奖品列表
func (h *AdminCheckInHandler) ListPrizes(c *gin.Context) {
	prizes, err := h.checkInService.ListPrizes(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, prizes)
}

// CreatePrizeReq 创建奖品请求
type CreatePrizeReq struct {
	Name         string  `json:"name" binding:"required"`
	PrizeType    string  `json:"prize_type" binding:"required"`
	Amount       float64 `json:"amount"`
	Weight       int     `json:"weight"`
	TotalStock   int     `json:"total_stock"`
	SortOrder    int     `json:"sort_order"`
	Icon         string  `json:"icon"`
}

// CreatePrize 创建奖品
func (h *AdminCheckInHandler) CreatePrize(c *gin.Context) {
	var req CreatePrizeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数格式错误")
		return
	}
	if req.Weight <= 0 {
		req.Weight = 10
	}
	if req.TotalStock <= 0 {
		req.TotalStock = -1
	}

	prize, err := h.checkInService.CreatePrize(c.Request.Context(), req.Name, req.PrizeType, req.Amount, req.Weight, req.TotalStock, req.SortOrder, req.Icon)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, prize)
}

// UpdatePrize 更新奖品
func (h *AdminCheckInHandler) UpdatePrize(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的奖品ID")
		return
	}

	var updates map[string]any
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.Error(c, http.StatusBadRequest, "参数格式错误")
		return
	}

	if err := h.checkInService.UpdatePrize(c.Request.Context(), id, updates); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "更新成功"})
}

// DeletePrize 删除奖品
func (h *AdminCheckInHandler) DeletePrize(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的奖品ID")
		return
	}

	if err := h.checkInService.DeletePrize(c.Request.Context(), id); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// GetLotteryRecords 获取所有抽奖记录
func (h *AdminCheckInHandler) GetLotteryRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	records, total, err := h.checkInService.GetAllLotteryRecords(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"total":   total,
		"records": records,
	})
}

// GetCheckInRecords 获取所有签到记录
func (h *AdminCheckInHandler) GetCheckInRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	records, total, err := h.checkInService.GetAllCheckInRecords(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"total":   total,
		"records": records,
	})
}

// GetLotteryStats 获取抽奖统计
func (h *AdminCheckInHandler) GetLotteryStats(c *gin.Context) {
	stats, err := h.checkInService.GetLotteryStats(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, stats)
}
