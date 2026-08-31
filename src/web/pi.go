// pi 集成（V3.0.0）：requiz 右侧聊天框 → exec pi CLI（题库目录为工作目录）
package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// piPath 定位 pi CLI（可被全局配置覆盖）
func piPath() string {
	candidates := []string{
		"C:/nvm/v25.1.0/pi.cmd",
		"pi",
	}
	for _, c := range candidates {
		if _, err := exec.LookPath(c); err == nil {
			return c
		}
	}
	return "pi"
}

// POST /api/pi/chat {message} ：以当前题库为工作目录调用 pi，返回回复
func apiPiChatHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "仅支持 POST", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil || strings.TrimSpace(body.Message) == "" {
			http.Error(w, `{"error":"需要 message 字段"}`, http.StatusBadRequest)
			return
		}
		// pi 工作目录 = 当前打开的题库目录
		cwd := s.main.Dir
		// 构造 pi 命令：pi --no-session -p <消息>（一次性问答）
		ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, piPath(), "--no-session", "-p", body.Message)
		cmd.Dir = cwd
		out, err := cmd.CombinedOutput()
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				http.Error(w, `{"error":"pi 响应超时（90s）"}`, http.StatusGatewayTimeout)
				return
			}
			http.Error(w, fmt.Sprintf(`{"error":"pi 调用失败: %s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		reply := strings.TrimSpace(string(out))
		if reply == "" {
			reply = "（pi 无回复）"
		}
		writeJSON(w, map[string]any{"ok": true, "reply": reply, "cwd": cwd})
	}
}
