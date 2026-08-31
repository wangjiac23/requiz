// 由 server.go 拆分生成（V1.6.0 模块化，功能零变化）
package web

import (
	"requiz/src/model"
	"requiz/src/config"
	"requiz/src/quiz"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func apiMetaValuesHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := s.bankByDir(r.URL.Query().Get("bank"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		pc, _ := config.ReadProjectConfig(b.Dir)
		merged := config.MergeFieldDefs(s.global.MetaFields, pc.MetaFields)
		out := map[string][]string{}
		for _, f := range merged {
			out[f.Name] = f.Values
		}
		writeJSON(w, out)
	}
}

// POST /api/meta-value/add {bank, field, value}：新值写入配置（全局字段→全局配置；否则→项目配置）

func apiMetaValueAddHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "仅支持 POST", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Bank string `json:"bank"`
			Field string `json:"field"`
			Value string `json:"value"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil || body.Field == "" || body.Value == "" {
			http.Error(w, `{"error":"需要 field 与 value 字段"}`, http.StatusBadRequest)
			return
		}
		b, err := s.bankByDir(body.Bank)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusNotFound)
			return
		}
		inGlobal := false
		for _, f := range s.global.MetaFields {
			if f.Name == body.Field {
				inGlobal = true
				break
			}
		}
		if inGlobal {
			// 值加入全局配置字段
			gc := s.global
			for i := range gc.MetaFields {
				if gc.MetaFields[i].Name == body.Field {
					if !containsStr(gc.MetaFields[i].Values, body.Value) {
						gc.MetaFields[i].Values = append(gc.MetaFields[i].Values, body.Value)
						if err := config.WriteGlobalConfig(gc); err != nil {
							http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
							return
						}
						s.mu.Lock()
						s.global = gc
						s.mu.Unlock()
					}
					break
				}
			}
		} else {
			// 自定义字段：写入该题库项目配置（字段不存在则新建）
			pc, _ := config.ReadProjectConfig(b.Dir)
			found := false
			for i := range pc.MetaFields {
				if pc.MetaFields[i].Name == body.Field {
					if !containsStr(pc.MetaFields[i].Values, body.Value) {
						pc.MetaFields[i].Values = append(pc.MetaFields[i].Values, body.Value)
					}
					found = true
					break
				}
			}
			if !found {
				pc.MetaFields = append(pc.MetaFields, model.FieldDef{Name: body.Field, Label: body.Field, Values: []string{body.Value}})
			}
			if err := config.WriteProjectConfig(b.Dir, pc); err != nil {
				http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
				return
			}
		}
		writeJSON(w, map[string]bool{"ok": true})
	}
}

// ---------- V1.3.0：配置管理 API ----------

// GET /api/config/global ：返回全局配置与路径

func apiConfigGlobalHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := s.global
		writeJSON(w, map[string]any{
			"path":        config.GlobalConfigPath(),
			"defaults":    gc.Defaults,
			"meta_fields": gc.MetaFields,
			"links":       gc.Links,
			"favorites":   gc.Favorites,
			"pi":          gc.Pi,
		})
	}
}

// POST /api/config/global {links?} ：更新全局配置（当前支持链接题库列表）

func apiConfigGlobalSaveHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "仅支持 POST", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Links []string `json:"links"`
			Pi    struct {
				Path  string `json:"path"`
				Model string `json:"model"`
			} `json:"pi"` // V3.1.0
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil {
			http.Error(w, `{"error":"body 解析失败"}`, http.StatusBadRequest)
			return
		}
		gc := s.global
		gc.Links = body.Links
		gc.Pi.Path = body.Pi.Path
		gc.Pi.Model = body.Pi.Model
		// 保护：不允许移除当前打开的题库
		if !containsStr(gc.Links, s.main.Dir) {
			http.Error(w, `{"error":"不能移除当前打开的题库"}`, http.StatusBadRequest)
			return
		}
		if err := config.WriteGlobalConfig(gc); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		s.mu.Lock()
		s.global = gc
		// 重新连接链接题库
		s.links = nil
		for _, linkDir := range gc.Links {
			if b, err := quiz.ConnectBank(linkDir); err == nil {
				s.links = append(s.links, b)
			}
		}
		s.mu.Unlock()
		writeJSON(w, map[string]bool{"ok": true})
	}
}

// GET /api/config/project?bank= ：返回项目配置与路径

func apiConfigProjectHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := s.bankByDir(r.URL.Query().Get("bank"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		pc, _ := config.ReadProjectConfig(b.Dir)
		writeJSON(w, map[string]any{
			"path":        config.ProjectConfigPath(b.Dir),
			"app":         pc.App,
			"bank":        pc.Bank,
			"meta_fields": pc.MetaFields,
		})
	}
}

// ---------- V1.5.0：收藏（项目配置）/ 清单 / 导出 ----------

// POST /api/favorite {bank, id}：收藏/取消（toggle，V1.5.0 存项目配置）
