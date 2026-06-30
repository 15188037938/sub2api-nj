package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
)

// 当前版本信息（编译时注入）
var (
	GitCommit = "unknown"
	BuildTime = "unknown"
)

// UpdateHandler handles online update operations
type UpdateHandler struct {
	repoPath   string // 代码路径，如 /opt/sub2api-nj
	deployPath string // deploy 目录
	githubUser string // GitHub 用户名 15188037938
	githubRepo string // GitHub 仓库 sub2api-nj
}

// NewUpdateHandler creates a new UpdateHandler
func NewUpdateHandler() *UpdateHandler {
	return &UpdateHandler{
		repoPath:   "/opt/sub2api-nj",
		deployPath: "/opt/sub2api-nj/deploy",
		githubUser: "15188037938",
		githubRepo: "sub2api-nj",
	}
}

// GetStatus returns the current update status
// GET /api/v1/admin/update/status
func (h *UpdateHandler) GetStatus(c *gin.Context) {
	latestCommit, _ := fetchLatestCommit(h.githubUser, h.githubRepo)

	isUpToDate := (latestCommit == "" || latestCommit == GitCommit)

	c.JSON(http.StatusOK, gin.H{
		"current_commit":   GitCommit,
		"latest_commit":    latestCommit,
		"build_time":       BuildTime,
		"is_up_to_date":    isUpToDate,
		"update_available": !isUpToDate,
	})
}

// ApplyUpdate starts the update process
// POST /api/v1/admin/update/apply
func (h *UpdateHandler) ApplyUpdate(c *gin.Context) {
	go h.runUpdate()

	c.JSON(http.StatusOK, gin.H{
		"message": "更新已开始执行，正在拉取最新代码并重建容器...",
		"status":  "in_progress",
	})
}

func (h *UpdateHandler) runUpdate() {
	commands := []string{
		fmt.Sprintf("cd %s && git pull origin main 2>&1", h.repoPath),
		fmt.Sprintf("cd %s && docker compose -f docker-compose.nj.yml --env-file .env build sub2api 2>&1", h.deployPath),
		fmt.Sprintf("cd %s && docker compose -f docker-compose.nj.yml --env-file .env up -d sub2api 2>&1", h.deployPath),
	}
	for _, cmd := range commands {
		exec.Command("bash", "-c", cmd).Run()
	}
}

func fetchLatestCommit(owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/main", owner, repo)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if sha, ok := result["sha"].(string); ok {
		if len(sha) >= 8 {
			return sha[:8], nil
		}
		return sha, nil
	}
	return "", fmt.Errorf("cannot get sha")
}
