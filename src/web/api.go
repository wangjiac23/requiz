// 由 server.go 拆分生成（V1.6.0 模块化，功能零变化）
package web

import (
	"requiz/src/parser"
	"requiz/src/quiz"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func apiOpenHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "仅支持 POST", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Bank string `json:"bank"`
			ID   string `json:"id"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil || body.ID == "" {
			http.Error(w, `{"error":"需要 bank 与 id 字段"}`, http.StatusBadRequest)
			return
		}
		b, err := s.bankByDir(body.Bank)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusNotFound)
			return
		}
		q, err := b.Find(body.ID)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusNotFound)
			return
		}
		bankAbs, err := filepath.Abs(b.Dir)
		if err != nil {
			http.Error(w, `{"error":"题库目录解析失败"}`, http.StatusInternalServerError)
			return
		}
		qAbs, err := filepath.Abs(q.Path)
		if err != nil {
			http.Error(w, `{"error":"文件路径解析失败"}`, http.StatusInternalServerError)
			return
		}
		rel, err := filepath.Rel(bankAbs, qAbs)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			http.Error(w, `{"error":"文件不在题库目录内，拒绝打开"}`, http.StatusForbidden)
			return
		}
		cmd := exec.Command("explorer.exe", "/select,", qAbs)
		if err := cmd.Start(); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]bool{"ok": true})
	}
}

// apiQuestionSaveHandler POST /api/question/save：编辑内容写回本地 md

func apiQuestionSaveHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "仅支持 POST", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Bank string            `json:"bank"`
			ID      string            `json:"id"`
			File    string            `json:"file"`
			Meta    map[string]string `json:"meta"`
			Prompt  string            `json:"prompt"`
			Answer  string            `json:"answer"`
			Explain string            `json:"explain"`
			Note    string            `json:"note"`
			Links   string            `json:"links"` // V2.2.0 链接笔记
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil || body.ID == "" {
			http.Error(w, `{"error":"需要 id 字段"}`, http.StatusBadRequest)
			return
		}
		b, err := s.bankByDir(body.Bank)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusNotFound)
			return
		}
		q, err := b.Find(body.ID)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusNotFound)
			return
		}
		// 改名（题目名 = 文件名，同目录内重命名）
		if body.File != "" && body.File != q.File {
			if strings.ContainsAny(body.File, `/\`) || body.File == "." || body.File == ".." {
				http.Error(w, `{"error":"非法文件名"}`, http.StatusBadRequest)
				return
			}
			newPath := filepath.Join(filepath.Dir(q.Path), body.File)
			if _, err := os.Stat(newPath); err == nil {
				http.Error(w, `{"error":"同名文件已存在"}`, http.StatusBadRequest)
				return
			}
			if err := os.Rename(q.Path, newPath); err != nil {
				http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
				return
			}
			q.Path = newPath
			q.File = body.File
		}
		// 元数据整体替换（app/bank/path 系统字段除外；删除行/留空即删除该字段）
		for k := range q.Meta {
			if k != "app" && k != "bank" && k != "path" {
				delete(q.Meta, k)
			}
		}
		if body.Meta != nil {
			for k, v := range body.Meta {
				if k == "app" || k == "bank" || k == "path" {
					continue
				}
				if strings.TrimSpace(v) != "" {
					q.Meta[k] = v
				}
			}
		}
		// path 元数据 = 相对题库目录路径（改名后自动维护）
		if rel, err := filepath.Rel(b.Dir, q.Path); err == nil {
			q.Meta["path"] = filepath.ToSlash(rel)
		}
		q.Prompt = body.Prompt
		q.Answer = body.Answer
		q.Explain = body.Explain
		q.Note = body.Note
		q.Links = body.Links
		q.Links = body.Links
		if err := os.WriteFile(q.Path, []byte(parser.SerializeQuestion(q)), 0644); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]bool{"ok": true})
	}
}

// apiReloadHandler POST /api/reload：重新扫描题库目录（同步本地/网页改动）

func apiReloadHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "仅支持 POST", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Bank string `json:"bank"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"error":"body 解析失败"}`, http.StatusBadRequest)
			return
		}
		b, err := s.bankByDir(body.Bank)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusNotFound)
			return
		}
		nb, err := quiz.ConnectBank(b.Dir)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		s.mu.Lock()
		if s.main.Dir == b.Dir {
			s.main = nb
		} else {
			for i, l := range s.links {
				if l.Dir == b.Dir {
					s.links[i] = nb
					break
				}
			}
		}
		s.mu.Unlock()
		writeJSON(w, map[string]any{"ok": true, "count": len(nb.Questions)})
	}
}

// ---------- V1.3.0：元数据字段定义（全局配置 + 项目配置） ----------

// GET /api/meta-values?bank= ：返回合并后的字段已知值 {field: [values]}

func apiBanksHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		banks := s.banks()
		list := make([]bankJSON, 0, len(banks))
		for i, b := range banks {
			list = append(list, bankJSON{
				Name:    b.Name,
				Dir:     b.Dir,
				Count:   len(b.Questions),
				Current: i == 0,
			})
		}
		writeJSON(w, list)
	}
}


func apiTreeHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := s.bankByDir(r.URL.Query().Get("bank"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSON(w, treeOf(b))
	}
}


func apiQuestionsHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := s.bankByDir(r.URL.Query().Get("bank"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		tag, value := r.URL.Query().Get("tag"), r.URL.Query().Get("value")
		list := make([]questionJSON, 0, len(b.Questions))
		for _, q := range b.Questions {
			if tag != "" {
				if q.Meta[tag] != value {
					continue
				}
			}
			list = append(list, toJSON(q))
		}
		writeJSON(w, list)
	}
}


func apiQuestionHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, `{"error":"缺少参数 id"}`, http.StatusBadRequest)
			return
		}
		b, err := s.bankByDir(r.URL.Query().Get("bank"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		q, err := b.Find(id)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusNotFound)
			return
		}
		writeJSON(w, toJSON(q))
	}
}


func apiLinkHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "仅支持 POST", http.StatusMethodNotAllowed)
			return
		}
		var body struct{ Dir string `json:"dir"` }
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil || body.Dir == "" {
			http.Error(w, `{"error":"需要 dir 字段"}`, http.StatusBadRequest)
			return
		}
		if err := s.addLink(body.Dir); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusBadRequest)
			return
		}
		writeJSON(w, map[string]bool{"ok": true})
	}
}

// ---------- 页面 ----------

