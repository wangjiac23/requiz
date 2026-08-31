// 由 server.go 拆分生成（V1.6.0 模块化，功能零变化）
package web

import (
	"requiz/src/model"
	"requiz/src/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func apiFavoriteHandler(s *Store) http.HandlerFunc {
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
		pc, _ := config.ReadProjectConfig(b.Dir)
		idx := -1
		for i, f := range pc.Favorites {
			if f == body.ID {
				idx = i
				break
			}
		}
		favorited := idx < 0
		if idx >= 0 {
			pc.Favorites = append(pc.Favorites[:idx], pc.Favorites[idx+1:]...)
		} else {
			pc.Favorites = append(pc.Favorites, body.ID)
		}
		if err := config.WriteProjectConfig(b.Dir, pc); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]any{"ok": true, "favorited": favorited})
	}
}

// GET /api/favorites?bank= ：收藏列表（id + 动态元数据）

func apiFavoritesHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := s.bankByDir(r.URL.Query().Get("bank"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		pc, _ := config.ReadProjectConfig(b.Dir)
		out := []map[string]any{}
		for _, id := range pc.Favorites {
			if q, err := b.Find(id); err == nil {
				out = append(out, map[string]any{
					"dir":   b.Dir,
					"id":    q.ID(),
					"file":  q.File,
					"title": firstLine(q.Prompt),
					"meta":  q.Meta,
				})
			}
		}
		writeJSON(w, out)
	}
}

// GET /api/lists?bank= ：自定义清单列表（含题目摘要）

func apiListsHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := s.bankByDir(r.URL.Query().Get("bank"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		pc, _ := config.ReadProjectConfig(b.Dir)
		out := []map[string]any{}
		for _, l := range pc.Lists {
			qs := []map[string]any{}
			for _, id := range l.IDs {
				if q, err := b.Find(id); err == nil {
					qs = append(qs, map[string]any{"id": q.ID(), "title": firstLine(q.Prompt), "meta": q.Meta})
				}
			}
			out = append(out, map[string]any{"name": l.Name, "ids": l.IDs, "count": len(qs), "questions": qs})
		}
		writeJSON(w, out)
	}
}

// POST /api/lists {bank, name, ids}：创建/更新清单；{bank, name, action:delete}：删除

func apiListsSaveHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "仅支持 POST", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Bank string   `json:"bank"`
			Name   string   `json:"name"`
			IDs    []string `json:"ids"`
			Action string   `json:"action"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil || body.Name == "" {
			http.Error(w, `{"error":"需要 name 字段"}`, http.StatusBadRequest)
			return
		}
		b, err := s.bankByDir(body.Bank)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusNotFound)
			return
		}
		pc, _ := config.ReadProjectConfig(b.Dir)
		if body.Action == "delete" {
			newLists := []model.QuestionList{}
			for _, l := range pc.Lists {
				if l.Name != body.Name {
					newLists = append(newLists, l)
				}
			}
			pc.Lists = newLists
		} else {
			found := false
			for i := range pc.Lists {
				if pc.Lists[i].Name == body.Name {
					pc.Lists[i].IDs = body.IDs
					found = true
					break
				}
			}
			if !found {
				pc.Lists = append(pc.Lists, model.QuestionList{Name: body.Name, IDs: body.IDs})
			}
		}
		if err := config.WriteProjectConfig(b.Dir, pc); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]bool{"ok": true})
	}
}

// POST /api/export {bank, ids, parts, format}：导出题目到 output/（json / html）
