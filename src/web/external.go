// 外部题库注入（V2.1.0）：Obsidian 插件复用 Obsidian API 扫描后注入
package web

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"requiz/src/model"
	"requiz/src/parser"
)

type externalQuestion struct {
	Path    string            `json:"path"`
	Content string            `json:"content"`
}

type externalBank struct {
	Name      string             `json:"name"`
	Dir       string             `json:"dir"`
	Questions []externalQuestion `json:"questions"`
}

// POST /api/external/banks ：注入题库（替换当前 main），插件扫描 Vault 后调用
func apiExternalBanksHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "仅支持 POST", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Bank externalBank `json:"bank"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 50<<20)).Decode(&body); err != nil {
			http.Error(w, `{"error":"body 解析失败"}`, http.StatusBadRequest)
			return
		}
		b := body.Bank
		if b.Name == "" {
			b.Name = "Obsidian Vault"
		}
		qs := make([]*model.Question, 0, len(b.Questions))
		for _, eq := range b.Questions {
			if strings.TrimSpace(eq.Content) == "" {
				continue
			}
			q, err := parser.ParseContent(eq.Content, eq.Path)
			if err != nil {
				continue
			}
			// 只收录 app: requiz 的题目
			if q.Meta["app"] != "requiz" {
				continue
			}
			qs = append(qs, q)
		}
		s.InjectBank(b.Name, b.Dir, qs)
		writeJSON(w, map[string]any{"ok": true, "name": b.Name, "count": len(qs)})
	}
}
