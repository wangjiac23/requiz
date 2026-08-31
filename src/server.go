// V1.1.0：多题库链接 + 三栏 Obsidian 风格 UI + 标签筛选
// 技术栈不变：仅 Go 标准库（net/http + html/template），零第三方依赖
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ---------- Store：多题库状态（主题库 + 链接题库） ----------

type Store struct {
	mu     sync.Mutex
	main   *Bank
	links  []*Bank
	global GlobalConfig // V1.3.0 全局配置（用户级）
}

func newStore(main *Bank) *Store {
	s := &Store{main: main}
	// V1.3.0：读全局配置（首次自动创建默认模板），主题库自动注册（Obsidian 式：打开即记录）
	gc, err := readGlobalConfig()
	if err == nil {
		s.global = gc
		// 主题库自动加入全局题库列表
		if !containsStr(gc.Links, main.Dir) {
			gc.Links = append(gc.Links, main.Dir)
			_ = writeGlobalConfig(gc)
		}
		s.global = gc
		migrateProjectLinks(main, &s.global)
	} else {
		s.global = defaultGlobalConfig()
	}
	// 加载全局配置中的全部题库（对等：都加载，主题库跳过重复）
	for _, bankDir := range s.global.Links {
		if bankDir == main.Dir {
			continue
		}
		if b, err := connectBank(bankDir); err == nil {
			s.links = append(s.links, b)
		}
	}
	return s
}

// migrateProjectLinks 旧版：项目配置中的 links 迁移到全局配置
func migrateProjectLinks(main *Bank, gc *GlobalConfig) {
	if len(main.Links) == 0 {
		return
	}
	changed := false
	for _, l := range main.Links {
		if !containsStr(gc.Links, l) {
			gc.Links = append(gc.Links, l)
			changed = true
		}
	}
	if changed {
		_ = writeGlobalConfig(*gc)
	}
	// 重写项目配置（去除旧版 links 节）
	pc, err := readProjectConfig(main.Dir)
	if err == nil {
		_ = writeProjectConfig(main.Dir, pc)
	}
}

func containsStr(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func (s *Store) banks() []*Bank {
	banks := []*Bank{s.main}
	return append(banks, s.links...)
}

// bankByDir 按目录找题库；dir 为空返回主题库；也支持按题库名匹配
func (s *Store) bankByDir(dir string) (*Bank, error) {
	if dir == "" {
		return s.main, nil
	}
	for _, b := range s.banks() {
		if b.Dir == dir {
			return b, nil
		}
	}
	for _, b := range s.banks() {
		if b.Name == dir {
			return b, nil
		}
	}
	return nil, fmt.Errorf("未链接的题库: %s", dir)
}

// addLink 连接新题库并持久化到全局配置（V1.3.0：links 存用户级全局配置）
func (s *Store) addLink(dir string) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if containsStr(s.global.Links, abs) {
		return nil // 已链接
	}
	b, err := connectBank(abs)
	if err != nil {
		return err
	}
	gc := s.global
	if !containsStr(gc.Links, abs) {
		gc.Links = append(gc.Links, abs)
	}
	if err := writeGlobalConfig(gc); err != nil {
		return err
	}
	s.mu.Lock()
	s.global.Links = append(s.global.Links, abs)
	s.links = append(s.links, b)
	s.mu.Unlock()
	return nil
}

// ---------- Web 服务 ----------

func cmdServe(args []string) error {
	dir, port, err := parseServeArgs(args)
	if err != nil {
		return err
	}
	bank, err := connectBank(dir)
	if err != nil {
		return err
	}
	store := newStore(bank)

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler(store))
	mux.HandleFunc("/question", questionHandler(store))
	mux.HandleFunc("/api/banks", apiBanksHandler(store))
	mux.HandleFunc("/api/tree", apiTreeHandler(store))
	mux.HandleFunc("/api/questions", apiQuestionsHandler(store))
	mux.HandleFunc("/api/question", apiQuestionHandler(store))
	mux.HandleFunc("/api/link", apiLinkHandler(store))
	mux.HandleFunc("/api/open", apiOpenHandler(store))
	mux.HandleFunc("/api/question/save", apiQuestionSaveHandler(store))
	mux.HandleFunc("/api/reload", apiReloadHandler(store))
	mux.HandleFunc("/api/meta-values", apiMetaValuesHandler(store))
	mux.HandleFunc("/api/meta-value/add", apiMetaValueAddHandler(store))
	mux.HandleFunc("/api/config/global", apiConfigGlobalHandler(store))
	mux.HandleFunc("/api/config/global/save", apiConfigGlobalSaveHandler(store))
	mux.HandleFunc("/api/config/project", apiConfigProjectHandler(store))
	mux.HandleFunc("/api/favorite", apiFavoriteHandler(store))
	mux.HandleFunc("/api/favorites", apiFavoritesHandler(store))
	mux.HandleFunc("/api/lists", apiListsHandler(store))
	mux.HandleFunc("/api/lists/save", apiListsSaveHandler(store))
	mux.HandleFunc("/api/export", apiExportHandler(store))
	mux.Handle("/katex/", http.StripPrefix("/katex/", http.FileServer(http.Dir(katexDir()))))

	addr := "127.0.0.1:" + port
	fmt.Printf("requiz web   : http://%s/\n", addr)
	fmt.Printf("当前题库     : %s（%d 题）\n", bank.Dir, len(bank.Questions))
	for _, b := range store.banks()[1:] {
		fmt.Printf("其它题库     : %s（%d 题）\n", b.Dir, len(b.Questions))
	}
	fmt.Println("按 Ctrl+C 停止服务")
	return http.ListenAndServe(addr, mux)
}

func parseServeArgs(args []string) (dir, port string, err error) {
	port = "8080"
	dir = "."
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-port", "--port":
			if i+1 >= len(args) {
				err = fmt.Errorf("-port 需要一个端口号")
				return
			}
			port = args[i+1]
			i++
		default:
			if strings.HasPrefix(args[i], "-") {
				err = fmt.Errorf("未知选项: %s", args[i])
				return
			}
			if dir != "." {
				err = fmt.Errorf("serve 参数过多（用法：requiz serve [题库目录] [-port 端口]）")
				return
			}
			dir = args[i]
		}
	}
	return
}

// katexDir 定位 KaTeX 静态资源目录（优先当前工作目录，其次可执行文件上级）
func katexDir() string {
	candidates := []string{"web/katex"}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "..", "web", "katex"))
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return "web/katex"
}

// ---------- V1.2.0 API：打开本地 / 编辑保存 / 刷新 ----------

// apiOpenHandler POST /api/open {bank, id}：explorer 定位本地题目文件
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
		q, err := b.find(body.ID)
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
			Bank    string            `json:"bank"`
			ID      string            `json:"id"`
			File    string            `json:"file"`
			Meta    map[string]string `json:"meta"`
			Prompt  string            `json:"prompt"`
			Answer  string            `json:"answer"`
			Explain string            `json:"explain"`
			Note    string            `json:"note"`
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
		q, err := b.find(body.ID)
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
		if err := os.WriteFile(q.Path, []byte(serializeQuestion(q)), 0644); err != nil {
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
		nb, err := connectBank(b.Dir)
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
func apiMetaValuesHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := s.bankByDir(r.URL.Query().Get("bank"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		pc, _ := readProjectConfig(b.Dir)
		merged := mergeFieldDefs(s.global.MetaFields, pc.MetaFields)
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
			Bank  string `json:"bank"`
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
						if err := writeGlobalConfig(gc); err != nil {
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
			pc, _ := readProjectConfig(b.Dir)
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
				pc.MetaFields = append(pc.MetaFields, FieldDef{Name: body.Field, Label: body.Field, Values: []string{body.Value}})
			}
			if err := writeProjectConfig(b.Dir, pc); err != nil {
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
			"path":        globalConfigPath(),
			"defaults":    gc.Defaults,
			"meta_fields": gc.MetaFields,
			"links":       gc.Links,
			"favorites":   gc.Favorites,
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
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil {
			http.Error(w, `{"error":"body 解析失败"}`, http.StatusBadRequest)
			return
		}
		gc := s.global
		gc.Links = body.Links
		// 保护：不允许移除当前打开的题库
		if !containsStr(gc.Links, s.main.Dir) {
			http.Error(w, `{"error":"不能移除当前打开的题库"}`, http.StatusBadRequest)
			return
		}
		if err := writeGlobalConfig(gc); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		s.mu.Lock()
		s.global = gc
		// 重新连接链接题库
		s.links = nil
		for _, linkDir := range gc.Links {
			if b, err := connectBank(linkDir); err == nil {
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
		pc, _ := readProjectConfig(b.Dir)
		writeJSON(w, map[string]any{
			"path":        projectConfigPath(b.Dir),
			"app":         pc.App,
			"bank":        pc.Bank,
			"meta_fields": pc.MetaFields,
		})
	}
}

// ---------- V1.5.0：收藏（项目配置）/ 清单 / 导出 ----------

// POST /api/favorite {bank, id}：收藏/取消（toggle，V1.5.0 存项目配置）
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
		pc, _ := readProjectConfig(b.Dir)
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
		if err := writeProjectConfig(b.Dir, pc); err != nil {
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
		pc, _ := readProjectConfig(b.Dir)
		out := []map[string]any{}
		for _, id := range pc.Favorites {
			if q, err := b.find(id); err == nil {
				out = append(out, map[string]any{
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
		pc, _ := readProjectConfig(b.Dir)
		out := []map[string]any{}
		for _, l := range pc.Lists {
			qs := []map[string]any{}
			for _, id := range l.IDs {
				if q, err := b.find(id); err == nil {
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
			Bank   string   `json:"bank"`
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
		pc, _ := readProjectConfig(b.Dir)
		if body.Action == "delete" {
			newLists := []QuestionList{}
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
				pc.Lists = append(pc.Lists, QuestionList{Name: body.Name, IDs: body.IDs})
			}
		}
		if err := writeProjectConfig(b.Dir, pc); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]bool{"ok": true})
	}
}

// POST /api/export {bank, ids, parts, format}：导出题目到 output/（json / html）
func apiExportHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "仅支持 POST", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Bank   string   `json:"bank"`
			IDs    []string `json:"ids"`
			Parts  []string `json:"parts"`
			Format string   `json:"format"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 2<<20)).Decode(&body); err != nil || len(body.IDs) == 0 {
			http.Error(w, `{"error":"需要 ids 字段"}`, http.StatusBadRequest)
			return
		}
		if body.Format == "" {
			body.Format = "json"
		}
		b, err := s.bankByDir(body.Bank)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusNotFound)
			return
		}
		// 收集题目
		qs := []*Question{}
		for _, id := range body.IDs {
			if q, err := b.find(id); err == nil {
				qs = append(qs, q)
			}
		}
		if len(qs) == 0 {
			http.Error(w, `{"error":"没有可导出的题目"}`, http.StatusBadRequest)
			return
		}
		if err := os.MkdirAll("output", 0755); err != nil {
			http.Error(w, `{"error":"output 目录创建失败"}`, http.StatusInternalServerError)
			return
		}
		stamp := time.Now().Format("20060102-150405")
		var path string
		var content []byte
		if body.Format == "html" {
			path = filepath.Join("output", fmt.Sprintf("export-%s.html", stamp))
			content = []byte(exportHTML(qs, body.Parts, b.Name))
		} else {
			path = filepath.Join("output", fmt.Sprintf("export-%s.json", stamp))
			content = []byte(exportJSON(qs, body.Parts))
		}
		if err := os.WriteFile(path, content, 0644); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]any{"ok": true, "path": path, "count": len(qs)})
	}
}

// exportJSON 导出题目为 JSON
func exportJSON(qs []*Question, parts []string) string {
	out := []map[string]any{}
	for _, q := range qs {
		item := map[string]any{"id": q.ID(), "file": q.File, "meta": q.Meta}
		for _, p := range parts {
			switch p {
			case "prompt":
				item["prompt"] = q.Prompt
			case "answer":
				item["answer"] = q.Answer
			case "explain":
				item["explain"] = q.Explain
			case "note":
				item["note"] = q.Note
			}
		}
		out = append(out, item)
	}
	data, _ := json.MarshalIndent(out, "", "  ")
	return string(data)
}

// exportHTML 导出题目为可打印 HTML（浏览器打印存 PDF / Word 兼容）
func exportHTML(qs []*Question, parts []string, bankName string) string {
	partName := map[string]string{"prompt": "题目", "answer": "答案", "explain": "解析", "note": "备注"}
	var b strings.Builder
	b.WriteString("<!DOCTYPE html><html lang=\"zh\"><head><meta charset=\"utf-8\"><title>requiz 导出</title>")
	b.WriteString("<style>body{font-family:'Microsoft YaHei',sans-serif;max-width:900px;margin:0 auto;padding:30px;color:#222}.q{border:1px solid #ccc;border-radius:8px;padding:14px 18px;margin:14px 0;page-break-inside:avoid}h2{font-size:16px;margin:0 0 8px}pre{white-space:pre-wrap;font-family:inherit;margin:6px 0}.sec-title{font-weight:bold;margin-top:8px;color:#555}</style></head><body>")
	fmt.Fprintf(&b, "<h1>题库：%s（%d 题）</h1>", bankName, len(qs))
	for i, q := range qs {
		fmt.Fprintf(&b, "<div class=\"q\"><h2>%d. %s</h2>", i+1, q.ID())
		for _, p := range parts {
			var content string
			switch p {
			case "prompt":
				content = q.Prompt
			case "answer":
				content = q.Answer
			case "explain":
				content = q.Explain
			case "note":
				content = q.Note
			}
			if content != "" {
				fmt.Fprintf(&b, "<div class=\"sec-title\">%s</div><pre>%s</pre>", partName[p], htmlEscape(content))
			}
		}
		b.WriteString("</div>")
	}
	b.WriteString("</body></html>")
	return b.String()
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// ---------- 题目视图模型 ----------

type questSummary struct {
	ID     string            `json:"id"`
	File   string            `json:"file"`
	Title  string            `json:"title"`
	Prompt string            `json:"prompt"` // V1.4.1 完整题干
	Meta   map[string]string `json:"meta"`
}

type pkgJSON struct {
	Name      string         `json:"name"`
	Path      string         `json:"path"`
	Count     int            `json:"count"`
	Questions []questSummary `json:"questions"`
}

type bankJSON struct {
	Name    string `json:"name"`
	Dir     string `json:"dir"`
	Count   int    `json:"count"`
	Current bool   `json:"current"`
}

// relPath 题目相对题库目录的路径（斜杠分隔）
func relPath(b *Bank, q *Question) string {
	rel, err := filepath.Rel(b.Dir, q.Path)
	if err != nil {
		return q.File
	}
	return filepath.ToSlash(rel)
}

func summaryOf(b *Bank, q *Question) questSummary {
	return questSummary{
		ID:     q.ID(),
		File:   q.File,
		Title:  firstLine(q.Prompt),
		Prompt: q.Prompt,
		Meta:   q.Meta,
	}
}

// treeOf 把题库题目按顶层目录分组为题目包树
func treeOf(b *Bank) []pkgJSON {
	byPkg := map[string]*pkgJSON{}
	var order []string
	for _, q := range b.Questions {
		rel := relPath(b, q)
		name := "未分组"
		path := ""
		if idx := strings.Index(rel, "/"); idx >= 0 {
			name = rel[:idx]
			path = rel[:idx]
		}
		p, ok := byPkg[name]
		if !ok {
			p = &pkgJSON{Name: name, Path: path}
			byPkg[name] = p
			order = append(order, name)
		}
		p.Questions = append(p.Questions, summaryOf(b, q))
		p.Count++
	}
	sort.Strings(order)
	pkgs := make([]pkgJSON, 0, len(order))
	for _, n := range order {
		p := byPkg[n]
		sort.Slice(p.Questions, func(i, j int) bool { return p.Questions[i].ID < p.Questions[j].ID })
		pkgs = append(pkgs, *p)
	}
	return pkgs
}

// ---------- JSON API ----------

type questionJSON struct {
	ID      string            `json:"id"`
	Path    string            `json:"path"`
	File    string            `json:"file"`
	Meta    map[string]string `json:"meta"`
	Prompt  string            `json:"prompt"`
	Answer  string            `json:"answer,omitempty"`
	Explain string            `json:"explain,omitempty"`
	Note    string            `json:"note,omitempty"`
	Extra   map[string]string `json:"extra,omitempty"`
}

func toJSON(q *Question) questionJSON {
	return questionJSON{
		ID:      q.ID(),
		Path:    q.Path,
		File:    q.File,
		Meta:    q.Meta,
		Prompt:  q.Prompt,
		Answer:  q.Answer,
		Explain: q.Explain,
		Note:    q.Note,
		Extra:   q.Extra,
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}

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
		q, err := b.find(id)
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

func indexHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		indexTpl.Execute(w, indexData{CSS: indexCSS, AppJS: indexJS})
	}
}

// questionHandler 整页详情（兼容分享链接）
func questionHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "缺少参数 id（?id=M001）", http.StatusBadRequest)
			return
		}
		b, err := s.bankByDir(r.URL.Query().Get("bank"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		q, err := b.find(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		data := questDetailView{
			ID:      q.ID(),
			File:    q.File,
			Prompt:  q.Prompt,
			Meta:    q.Meta,
			Answer:  q.Answer,
			Explain: q.Explain,
			Note:    q.Note,
			Extra:   q.Extra,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		questionTpl.Execute(w, data)
	}
}

type questDetailView struct {
	ID      string
	File    string
	Prompt  string
	Meta    map[string]string
	Answer  string
	Explain string
	Note    string
	Extra   map[string]string
}

type indexData struct {
	CSS   template.CSS
	AppJS template.JS
}

var indexTpl = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html lang="zh">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>requiz</title>
<link rel="stylesheet" href="/katex/katex.min.css">
<style>{{.CSS}}</style>
</head>
<body>
<header id="topbar">
  <div class="brand">📚 requiz</div>
  <div class="bankbar">
    <button id="toggleSidebar" title="隐藏侧边栏">☷</button>
    <button id="toggleFilters" title="隐藏筛选栏">⌄</button>
    <select id="bankSel" title="选择题库"></select>
    <button id="reloadBtn" title="刷新题库（同步本地/网页改动）">⟳</button>
    <button id="settingsBtn" title="设置">⚙ 设置</button>
  </div>
</header>
<div id="filtersWrap">
  <div id="filtersBar">
    <div id="filters"></div>
    <button id="pinFilters" title="固定筛选栏（固定时拖拽只调高不隐藏）">📌</button>
  </div>
  <div id="filtersResizer" title="拖拽调整高度（拖到最矮隐藏）"></div>
</div>
<div id="body">
  <div id="sidebarWrap">
    <div id="sidebarCol">
      <div id="sidebarHead">
        <div id="sideTabs">
          <button class="stab active" data-tab="tree" title="题库题目导航">题库</button>
          <button class="stab" data-tab="fav" title="收藏题目导航">收藏</button>
          <button class="stab" data-tab="lists" title="自定义题目清单">清单</button>
        </div>
        <button id="pinSidebar" title="固定侧边栏（固定时拖拽只调宽不隐藏）">📌</button>
      </div>
      <aside id="sidebar"></aside>
    </div>
    <div id="resizer" title="拖拽调整宽度（拖到最窄隐藏）"></div>
  </div>
  <main id="main">
    <div id="mainToolbar">
      <button class="mode active" data-mode="list" title="列表浏览">📋 列表</button>
      <button class="mode" data-mode="split" title="双栏浏览">📑 双栏</button>
      <button class="mode" data-mode="card" title="卡片浏览">🃏 卡片</button>
      <span style="flex:1"></span>
      <button id="selectBtn" title="选择模式（勾选题目）">☑ 选择</button>
      <button id="saveListBtn" title="将选中题目存为清单" style="display:none">📋 存清单</button>
      <button id="exportBtn" title="导出选中题目" style="display:none">📤 导出</button>
      <button id="displayBtn" title="自定义显示字段">⚙ 字段</button>
      <button id="favFilterBtn" title="只看收藏的题目">☆ 收藏</button>
    </div>
    <div id="mainContent"><div class="empty">加载中…</div></div>
  </main>
</div>
<div id="modal" hidden>
  <div class="modal-box" style="width:600px;max-height:85vh;overflow-y:auto">
    <h3>⚙ 设置 <small style="color:var(--muted);font-weight:400">配置管理</small></h3>
    <div class="cfg-tabs">
      <button id="tabGlobal" class="cfg-tab active">🌐 全局配置</button>
      <button id="tabProject" class="cfg-tab">📁 题库配置</button>
    </div>
    <!-- 全局配置页签 -->
    <div id="cfgGlobal">
      <p class="tip">配置文件：<code id="cfgGlobalPath"></code></p>
      <h4>默认配置</h4>
      <div id="cfgDefaults"></div>
      <h4>题库列表（全部题库，当前打开的不可移除）<button id="linkOk" style="margin-left:8px;padding:2px 8px">＋ 添加</button></h4>
      <div id="cfgLinks"></div>
      <input id="linkInput" type="text" placeholder="输入题库目录路径后点链接" style="width:100%;margin:4px 0 8px">
      <h4>元数据字段定义</h4>
      <div id="cfgFields"></div>
    </div>
    <!-- 题库配置页签 -->
    <div id="cfgProject" hidden>
      <p class="tip">配置文件：<code id="cfgProjectPath"></code></p>
      <h4>题库信息</h4>
      <div id="cfgProjectInfo"></div>
      <h4>自定义属性字段</h4>
      <div id="cfgProjectFields"></div>
    </div>
    <div id="linkMsg" class="tip"></div>
    <div class="modal-actions">
      <button id="linkCancel">关闭</button>
    </div>
  </div>
</div>
<div id="displayModal" hidden>
  <div class="modal-box" style="width:360px">
    <h3>显示字段清单 <small style="color:var(--muted);font-weight:400">勾选要在题目上显示的字段</small></h3>
    <div id="displayList"></div>
    <div id="displayMsg" class="tip"></div>
    <div class="modal-actions">
      <button id="displayOk">保存</button>
      <button id="displayCancel">取消</button>
    </div>
  </div>
</div>
<div id="exportModal" hidden>
  <div class="modal-box" style="width:380px">
    <h3>导出题目 <span id="exportCount" class="meta"></span></h3>
    <label class="lbl">导出部分</label>
    <div id="exportParts">
      <label style="display:block;margin:3px 0"><input type="checkbox" value="prompt" checked> 题干</label>
      <label style="display:block;margin:3px 0"><input type="checkbox" value="answer"> 答案</label>
      <label style="display:block;margin:3px 0"><input type="checkbox" value="explain"> 解析</label>
      <label style="display:block;margin:3px 0"><input type="checkbox" value="note"> 备注</label>
    </div>
    <label class="lbl">导出格式</label>
    <select id="exportFormat" style="width:100%;padding:6px;margin:4px 0 8px;border:1px solid var(--border);border-radius:6px">
      <option value="json">JSON（自定义清单）</option>
      <option value="html">HTML（浏览器打印存 PDF / Word 可打开）</option>
    </select>
    <div id="exportMsg" class="tip"></div>
    <div class="modal-actions">
      <button id="exportOk">导出</button>
      <button id="exportCancel">取消</button>
    </div>
  </div>
</div>
<div id="editModal" hidden>
  <div class="modal-box">
    <h3>编辑题目 <span id="editId" class="meta"></span></h3>
    <label class="lbl">题目名（文件名）</label>
    <input id="editFile" type="text">
    <label class="lbl" id="metaFoldBtn" style="cursor:pointer;user-select:none">▾ 元数据</label>
    <div id="editMeta"></div>
    <div style="text-align:right;margin-bottom:8px"><button id="editAddField">＋ 添加字段</button></div>
    <label class="lbl">题干</label>
    <textarea id="editPrompt" rows="5"></textarea>
    <label class="lbl">答案</label>
    <textarea id="editAnswer" rows="3"></textarea>
    <label class="lbl">解析</label>
    <textarea id="editExplain" rows="3"></textarea>
    <label class="lbl">备注</label>
    <textarea id="editNote" rows="2"></textarea>
    <div id="editMsg" class="tip"></div>
    <div class="modal-actions">
      <button id="editSave">保存</button>
      <button id="editCancel">取消</button>
    </div>
  </div>
</div>
<script src="/katex/katex.min.js"></script>
<script src="/katex/contrib/auto-render.min.js"></script>
<script>{{.AppJS}}</script>
</body>
</html>`))

var questionTpl = template.Must(template.New("question").Parse(`<!DOCTYPE html>
<html lang="zh">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.ID}} — requiz</title>
<link rel="stylesheet" href="/katex/katex.min.css">
<style>
 body{font-family:-apple-system,"Segoe UI","Microsoft YaHei",sans-serif;max-width:800px;margin:0 auto;padding:24px;background:#f7f8fa;color:#24292f}
 .card{background:#fff;border:1px solid #e1e4e8;border-radius:8px;padding:16px;margin:12px 0}
 .meta{color:#586069;font-size:0.85em}
 .content{white-space:pre-wrap;background:#f6f8fa;padding:12px;border-radius:6px;overflow-x:auto;font-size:0.9em}
 .tag{display:inline-block;background:#ddf4ff;color:#0969da;border-radius:12px;padding:2px 10px;margin:2px;font-size:0.85em}
 a{color:#0969da;text-decoration:none}
</style>
</head>
<body>
<p><a href="/">← 返回题库</a></p>
<h1>{{.ID}} <small class="meta">{{.File}}</small></h1>
<div class="card">{{range $k, $v := .Meta}}<span class="tag">{{$k}}: {{$v}}</span>{{end}}</div>
{{if .Prompt}}<div class="card"><h2>题目</h2><div class="content">{{.Prompt}}</div></div>{{end}}
{{if .Answer}}<div class="card"><h2>答案</h2><div class="content">{{.Answer}}</div></div>{{end}}
{{if .Explain}}<div class="card"><h2>解析</h2><div class="content">{{.Explain}}</div></div>{{end}}
{{if .Note}}<div class="card"><h2>备注</h2><div class="content">{{.Note}}</div></div>{{end}}
{{range $k, $v := .Extra}}<div class="card"><h2>{{$k}}</h2><div class="content">{{$v}}</div></div>{{end}}
<script src="/katex/katex.min.js"></script>
<script src="/katex/contrib/auto-render.min.js"></script>
<script>
window.addEventListener("DOMContentLoaded", function(){
  if (window.renderMathInElement) {
    renderMathInElement(document.body, {
      delimiters: [
        {left: "$$", right: "$$", display: true},
        {left: "$", right: "$", display: false},
        {left: "\\\(", right: "\\\)", display: false},
        {left: "\\[", right: "\\]", display: true}
      ],
      throwOnError: false
    });
  }
});
</script>
</body>
</html>`))

// ---------- CSS / JS（以 template.JS 注入，避免模板引擎解析） ----------

const indexCSS = `
:root{--bg:#f7f8fa;--panel:#fff;--border:#e1e4e8;--text:#24292f;--muted:#586069;--accent:#0969da;--accent-bg:#ddf4ff;--sidebar:#f5f6f8;--hover:#eef2f5}
*{box-sizing:border-box}
html,body{margin:0;height:100%;overflow:hidden}
body{font-family:-apple-system,"Segoe UI","Microsoft YaHei",sans-serif;background:var(--bg);color:var(--text);display:flex;flex-direction:column}
#topbar{display:flex;align-items:center;justify-content:space-between;height:48px;padding:0 14px;background:var(--panel);border-bottom:1px solid var(--border);position:sticky;top:0;z-index:10}
.brand{font-size:16px;font-weight:600}
.brand small{color:var(--muted);font-weight:400}
.bankbar{display:flex;gap:8px;align-items:center}
select,input,button{font-family:inherit;font-size:13px}
#bankSel{width:220px;padding:6px 8px;border:1px solid var(--border);border-radius:6px;background:var(--panel)}
button{padding:6px 12px;border:1px solid var(--border);border-radius:6px;background:var(--panel);cursor:pointer}
button:hover{background:var(--hover)}
#filtersWrap{display:flex;flex-direction:column}
#filtersWrap.hidden{display:none}
#filtersBar{display:flex;align-items:center}
#filters{flex:1;display:flex;flex-wrap:wrap;gap:8px;padding:8px 14px;background:var(--panel);align-items:center;overflow-y:auto}
#filtersBar button{padding:4px 8px;margin:4px;font-size:12px}
#filtersResizer{height:5px;cursor:row-resize;flex-shrink:0;background:transparent}
#filtersResizer:hover,#filtersResizer.active{background:var(--accent-bg)}
#filters .f-item{display:flex;align-items:center;gap:4px;color:var(--muted);font-size:12px}
#filters select{max-width:150px;padding:4px 6px;border:1px solid var(--border);border-radius:6px}
#filters .clear{border:none;background:none;color:var(--accent);cursor:pointer;font-size:12px}
#body{flex:1;min-height:0;display:flex;overflow:hidden}
#sidebarWrap{display:flex}
#sidebarWrap.hidden{display:none}
#sidebarCol{display:flex;flex-direction:column;min-width:0;border-right:1px solid var(--border)}
#sidebarHead{display:flex;align-items:center;justify-content:space-between;padding:4px 8px;font-size:12px;color:var(--muted);background:var(--sidebar);border-bottom:1px solid var(--border)}
#sidebarHead button{padding:2px 6px;font-size:12px}
#sideTabs{display:flex;gap:2px}
.stab{border:1px solid transparent;background:none;color:var(--muted);cursor:pointer;font-size:12px;padding:3px 8px;border-radius:6px}
.stab.active{background:var(--accent-bg);color:var(--accent)}
.sel-cb{margin-right:6px;accent-color:var(--accent)}
#sidebar{flex:1;width:260px;min-width:120px;max-width:480px;background:var(--sidebar);overflow-y:auto;padding:8px 6px}
#resizer{width:5px;cursor:col-resize;flex-shrink:0;background:transparent}
#resizer:hover,#resizer.active{background:var(--accent-bg)}
button.pinned{background:var(--accent-bg);color:var(--accent);border-color:var(--accent)}
#main{flex:1;overflow-y:auto;padding:0}
#mainToolbar{display:flex;align-items:center;gap:6px;padding:8px 14px;background:var(--panel);border-bottom:1px solid var(--border);position:sticky;top:0;z-index:5}
#mainToolbar .mode{padding:4px 10px;font-size:12px}
#mainToolbar .mode.active{background:var(--accent-bg);color:var(--accent);border-color:var(--accent)}
#mainContent{padding:16px}
.qbox{background:var(--panel);border:1px solid var(--border);border-radius:8px;padding:12px 16px;margin-bottom:12px}
.qbox.active{border-color:var(--accent)}
.qbox-head{display:flex;align-items:center;gap:8px;margin-bottom:6px}
.qbox-id{font-weight:600;font-size:14px}
.qbox-tags{flex:1;display:flex;gap:4px;flex-wrap:wrap}
.qbox-actions{display:flex;gap:4px}
.fav-btn{border:none;background:none;font-size:16px;cursor:pointer;padding:2px}
.exp-btn{padding:3px 10px;font-size:12px}
.qbox-detail{margin-top:8px;border-top:1px dashed var(--border);padding-top:6px}
.sec-btn{display:block;width:100%;text-align:left;padding:6px 10px;margin:4px 0;background:var(--sidebar);border:1px solid var(--border);border-radius:6px;cursor:pointer;font-size:13px;font-weight:600}
.sec-body{padding:8px 10px}
.split-wrap{display:flex;gap:14px;height:calc(100vh - 140px)}
.split-left{width:42%;min-width:280px;overflow-y:auto;padding-right:6px}
.split-right{flex:1;overflow-y:auto;border:1px solid var(--border);border-radius:8px;padding:14px 18px;background:var(--panel)}
.card-nav{display:flex;align-items:center;justify-content:center;gap:12px;margin-top:12px}
.card-nav span{color:var(--muted);font-size:13px}
.split-right .qbox{margin-bottom:12px}
.pkg{font-size:13px}
.pkg-head{display:flex;align-items:center;gap:4px;padding:5px 8px;border-radius:6px;cursor:pointer;color:var(--text);font-weight:600}
.pkg-head:hover{background:var(--hover)}
.pkg-head .cnt{color:var(--muted);font-weight:400;font-size:12px}
.pkg-body{padding-left:18px}
.q-item{padding:4px 8px;border-radius:6px;cursor:pointer;font-size:13px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.q-item:hover{background:var(--hover)}
.q-item.active{background:var(--accent-bg)}
.card{background:var(--panel);border:1px solid var(--border);border-radius:8px;padding:14px 16px;margin-bottom:12px}
.card h3{margin:0 0 6px;font-size:15px}
.card .qmeta{color:var(--muted);font-size:12px;margin-bottom:8px}
.tag{display:inline-block;background:var(--accent-bg);color:var(--accent);border-radius:12px;padding:1px 9px;margin:2px;font-size:12px}
.content{white-space:pre-wrap;background:#f6f8fa;padding:12px;border-radius:6px;font-size:14px;line-height:1.6}
pre{white-space:pre-wrap;background:#f6f8fa;padding:12px;border-radius:6px;font-size:14px;line-height:1.6}
.detail{border-top:1px dashed var(--border);margin-top:10px;padding-top:8px}
.empty{color:var(--muted);text-align:center;padding:40px}
#modal{position:fixed;inset:0;background:rgba(0,0,0,.35);display:none;align-items:center;justify-content:center;z-index:100}
#modal.show{display:flex}
#editModal{position:fixed;inset:0;background:rgba(0,0,0,.35);display:none;align-items:center;justify-content:center;z-index:100}
#editModal.show{display:flex}
#displayModal{position:fixed;inset:0;background:rgba(0,0,0,.35);display:none;align-items:center;justify-content:center;z-index:100}
#displayModal.show{display:flex}
#exportModal{position:fixed;inset:0;background:rgba(0,0,0,.35);display:none;align-items:center;justify-content:center;z-index:100}
#exportModal.show{display:flex}
#displayList{display:flex;flex-direction:column;gap:4px;margin:10px 0}
#displayList label{display:flex;align-items:center;gap:6px;font-size:13px;cursor:pointer}
#editModal .modal-box{width:560px;max-width:92%;max-height:85vh;overflow-y:auto}
#editModal textarea{width:100%;font-family:inherit;font-size:13px;padding:6px;border:1px solid var(--border);border-radius:6px;margin:2px 0 8px;resize:vertical}
#editModal .lbl{font-size:12px;color:var(--muted)}
#editModal input[type=text]{width:100%;font-family:inherit;font-size:13px;padding:6px;border:1px solid var(--border);border-radius:6px;margin:2px 0 8px}
#editMeta{display:flex;flex-direction:column;gap:6px;margin-bottom:8px}
#editMeta.folded{display:none}
.meta-row{display:grid;grid-template-columns:80px 1fr 28px;gap:6px;align-items:center}
.meta-row .cust{grid-column:2;display:flex;gap:4px}
.meta-row .cust input{flex:1;padding:4px 6px;border:1px solid var(--border);border-radius:6px;font-size:13px}
.meta-row .cust .hint{font-size:11px;color:var(--muted);white-space:nowrap}
.meta-del{border:none;background:none;color:var(--muted);cursor:pointer;font-size:14px;padding:2px}
.meta-del:hover{color:#d1242f}
.q-actions{display:flex;gap:8px;margin-bottom:10px}
.q-actions button{font-size:12px;padding:4px 10px}
#reloadBtn{padding:6px 10px}
.modal-box{background:var(--panel);border-radius:10px;padding:20px;width:420px;max-width:90%}
#modal .modal-box h4{margin:10px 0 6px;font-size:13px;color:var(--text)}
.cfg-tabs{display:flex;gap:6px;margin-bottom:10px}
.cfg-tab{padding:5px 14px;border:1px solid var(--border);border-radius:6px;background:var(--panel);cursor:pointer;font-size:13px}
.cfg-tab.active{background:var(--accent-bg);color:var(--accent);border-color:var(--accent)}
.cfg-item{display:flex;align-items:center;justify-content:space-between;gap:8px;padding:4px 8px;border:1px solid var(--border);border-radius:6px;margin:4px 0;font-size:13px}
.cfg-item .path{color:var(--muted);font-size:12px;word-break:break-all}
.cfg-item .vals{color:var(--muted);font-size:12px}
.cfg-del{border:none;background:none;color:var(--muted);cursor:pointer;font-size:14px}
.cfg-del:hover{color:#d1242f}
.cfg-open{padding:2px 8px;font-size:12px}
#modal code{background:#f6f8fa;padding:1px 6px;border-radius:4px;font-size:12px;word-break:break-all}
.modal-box h3{margin:0 0 8px}
.tip{color:var(--muted);font-size:12px}
#linkInput{width:100%;padding:8px;margin:10px 0;border:1px solid var(--border);border-radius:6px}
.modal-actions{display:flex;gap:8px;justify-content:flex-end}
`

const indexJS = `
var state = { banks: [], bank: "", tree: [], filters: {}, expanded: {}, mode: "list", favOnly: false, cardIdx: 0, favs: {}, sideTab: "tree", selectMode: false, selected: {} };

function qs(s){ return document.querySelector(s); }
function esc(s){ var d=document.createElement("div"); d.textContent = (s==null?"":s); return d.innerHTML; }
function tagName(k){
  var names = {chapter:"章节",grade:"年级",difficulty:"难度",importance:"重要性",source:"来源",knowledge:"知识点",type:"题型"};
  return names[k] || k;
}

// KaTeX 公式渲染配置与函数（div.content 内 $...$/$$...$$ 自动渲染）
var katexDelims = [
  {left: "$$", right: "$$", display: true},
  {left: "$", right: "$", display: false},
  {left: "\\\(", right: "\\\)", display: false},
  {left: "\\[", right: "\\]", display: true}
];
function renderMath(el){
  if (window.renderMathInElement) {
    renderMathInElement(el, {delimiters: katexDelims, throwOnError: false});
  }
}

function init(){
  qs("#settingsBtn").onclick = openSettings;
  qs("#linkCancel").onclick = closeSettings;
  qs("#linkOk").onclick = doLink;
  qs("#linkInput").addEventListener("keydown", function(e){ if(e.key==="Enter") doLink(); });
  qs("#bankSel").onchange = function(){
    state.bank = this.value;
    state.filters = {};
    loadAll();
  };
  qs("#toggleSidebar").onclick = function(){
    var h = qs("#sidebarWrap").classList.toggle("hidden");
    this.textContent = h ? "☰" : "☷";
    this.title = h ? "显示侧边栏" : "隐藏侧边栏";
  };
  qs("#toggleFilters").onclick = function(){
    var h = qs("#filtersWrap").classList.toggle("hidden");
    this.textContent = h ? "⌃" : "⌄";
    this.title = h ? "显示筛选栏" : "隐藏筛选栏";
  };
  qs("#pinSidebar").onclick = function(){ this.classList.toggle("pinned"); };
  qs("#pinFilters").onclick = function(){ this.classList.toggle("pinned"); };
  qs("#reloadBtn").onclick = reloadBank;
  qs("#editSave").onclick = saveEdit;
  qs("#editCancel").onclick = closeEdit;
  qs("#editAddField").onclick = addField;
  qs("#favFilterBtn").onclick = function(){
    state.favOnly = !state.favOnly;
    this.textContent = state.favOnly ? "★ 收藏（仅看）" : "☆ 收藏";
    this.classList.toggle("active", state.favOnly);
    loadAll();
  };
  qs("#selectBtn").onclick = toggleSelectMode;
  qs("#exportBtn").onclick = openExport;
  qs("#saveListBtn").onclick = saveSelectedAsList;
  qs("#exportOk").onclick = doExport;
  qs("#exportCancel").onclick = closeExport;
  qs("#displayBtn").onclick = openDisplay;
  qs("#displayOk").onclick = saveDisplay;
  qs("#displayCancel").onclick = closeDisplay;
  document.querySelectorAll("#sideTabs .stab").forEach(function(b){
    b.onclick = function(){ switchSideTab(b.getAttribute("data-tab")); };
  });
  document.querySelectorAll("#mainToolbar .mode").forEach(function(b){
    b.onclick = function(){ setMode(b.getAttribute("data-mode")); };
  });
  // V1.4.3：键盘导航（双栏/卡片模式：上/左=上一题，下/右=下一题）
  document.addEventListener("keydown", function(e){
    if (state.mode !== "split" && state.mode !== "card") return;
    var t = e.target;
    if (t && (t.tagName === "INPUT" || t.tagName === "TEXTAREA" || t.tagName === "SELECT")) return;
    if (e.key === "ArrowUp" || e.key === "ArrowLeft") { navQuestion(-1); e.preventDefault(); }
    else if (e.key === "ArrowDown" || e.key === "ArrowRight") { navQuestion(1); e.preventDefault(); }
  });
  qs("#metaFoldBtn").onclick = function(){
    var folded = qs("#editMeta").classList.toggle("folded");
    this.textContent = folded ? "▸ 元数据" : "▾ 元数据";
  };
  loadBanks();
}

// 侧边栏拖拽调宽（拖到最窄自动隐藏）
(function(){
  var resizer = qs("#resizer"), sb = qs("#sidebar"), wrap = qs("#sidebarWrap");
  var sx = 0, sw = 0;
  resizer.addEventListener("mousedown", function(e){
    sx = e.clientX; sw = sb.offsetWidth;
    resizer.classList.add("active");
    document.addEventListener("mousemove", onMove);
    document.addEventListener("mouseup", onUp);
    e.preventDefault();
  });
  function onMove(e){
    var w = sw + (e.clientX - sx);
    if (w < 100) {
      if (qs("#pinSidebar").classList.contains("pinned")) {
        w = 100; // 固定时只调宽不隐藏
      } else {
        wrap.classList.add("hidden");
        qs("#toggleSidebar").textContent = "☰";
        qs("#toggleSidebar").title = "显示侧边栏";
        onUp();
        return;
      }
    }
    if (w > 480) w = 480;
    sb.style.width = w + "px";
  }
  function onUp(){
    resizer.classList.remove("active");
    document.removeEventListener("mousemove", onMove);
    document.removeEventListener("mouseup", onUp);
  }
})();

// 筛选栏拖拽调高（拖到最矮自动隐藏）
(function(){
  var resizer = qs("#filtersResizer"), f = qs("#filters"), wrap = qs("#filtersWrap");
  var sy = 0, sh = 0;
  resizer.addEventListener("mousedown", function(e){
    sy = e.clientY; sh = f.offsetHeight;
    resizer.classList.add("active");
    document.addEventListener("mousemove", onMove);
    document.addEventListener("mouseup", onUp);
    e.preventDefault();
  });
  function onMove(e){
    var h = sh + (e.clientY - sy);
    if (h < 40) {
      if (qs("#pinFilters").classList.contains("pinned")) {
        h = 40; // 固定时只调高不隐藏
      } else {
        wrap.classList.add("hidden");
        qs("#toggleFilters").textContent = "⌃";
        qs("#toggleFilters").title = "显示筛选栏";
        onUp();
        return;
      }
    }
    if (h > 160) h = 160;
    f.style.height = h + "px";
  }
  function onUp(){
    resizer.classList.remove("active");
    document.removeEventListener("mousemove", onMove);
    document.removeEventListener("mouseup", onUp);
  }
})();

function loadBanks(){
  fetch("/api/banks").then(function(r){ return r.json(); }).then(function(banks){
    state.banks = banks;
    var sel = qs("#bankSel");
    sel.innerHTML = "";
    banks.forEach(function(b){
      var o = document.createElement("option");
      o.value = b.dir; o.text = b.name + "（" + b.count + " 题）";
      if (b.current){ o.selected = true; state.bank = b.dir; }
      sel.appendChild(o);
    });
    loadAll();
  });
}

function loadAll(){
  fetch("/api/tree?bank=" + encodeURIComponent(state.bank)).then(function(r){ return r.json(); }).then(function(tree){
    state.tree = tree;
    renderSidebar();
    renderFilters();
    // 加载收藏状态
    fetch("/api/favorites").then(function(r){ return r.json(); }).then(function(favs){
      state.favs = {};
      favs.forEach(function(f){ state.favs[f.dir + "|" + f.id] = true; });
      renderMain();
    });
  });
}

// 模式切换
function setMode(m){
  state.mode = m;
  document.querySelectorAll("#mainToolbar .mode").forEach(function(b){
    b.classList.toggle("active", b.getAttribute("data-mode") === m);
  });
  if (m === "card" && state.cardIdx >= visibleQuestions().length) state.cardIdx = 0;
  renderMain();
}

function renderMain(){
  if (state.mode === "split") renderSplit();
  else if (state.mode === "card") renderCard();
  else renderList();
}

function isFav(q){
  return !!(state.favs[state.bank + "|" + q.id]);
}

function toggleFav(q, btn){
  fetch("/api/favorite", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({bank: state.bank, id: q.id})
  }).then(function(r){ return r.json(); }).then(function(d){
    if (d.ok) {
      if (d.favorited) state.favs[state.bank + "|" + q.id] = true;
      else delete state.favs[state.bank + "|" + q.id];
      if (btn) btn.textContent = d.favorited ? "⭐" : "☆";
      if (state.favOnly) renderMain();
    }
  });
}

// ---------- V1.5.0：侧边栏页签 / 选择 / 导出 ----------

// 切换侧边栏页签（题库 / 收藏 / 清单）
function switchSideTab(tab){
  state.sideTab = tab;
  document.querySelectorAll("#sideTabs .stab").forEach(function(b){
    b.classList.toggle("active", b.getAttribute("data-tab") === tab);
  });
  renderSidebar();
}

// 收藏导航
function renderFavSidebar(sb){
  fetch("/api/favorites?bank=" + encodeURIComponent(state.bank)).then(function(r){ return r.json(); }).then(function(favs){
    sb.innerHTML = "";
    if (favs.length === 0) { sb.innerHTML = '<div class="empty">暂无收藏（题目上点 ☆ 收藏）</div>'; return; }
    favs.forEach(function(f){
      var it = document.createElement("div");
      it.className = "q-item";
      it.textContent = f.id + " · " + (f.title || f.file);
      it.title = f.file;
      it.onclick = function(){ selectQuestion({id: f.id}); };
      sb.appendChild(it);
    });
  });
}

// 清单导航
function renderListsSidebar(sb){
  fetch("/api/lists?bank=" + encodeURIComponent(state.bank)).then(function(r){ return r.json(); }).then(function(lists){
    sb.innerHTML = "";
    if (lists.length === 0) { sb.innerHTML = '<div class="empty">暂无清单（选择题目后「📋 存清单」）</div>'; return; }
    lists.forEach(function(l){
      var head = document.createElement("div");
      head.className = "pkg-head";
      head.innerHTML = "▸ " + esc(l.name) + ' <span class="cnt">' + l.count + "</span>";
      head.onclick = function(){
        var body = head.nextElementSibling;
        if (body) {
          var open = body.style.display !== "none";
          body.style.display = open ? "none" : "block";
          head.textContent = (open ? "▸ " : "▾ ") + l.name + " (" + l.count + ")";
        }
      };
      sb.appendChild(head);
      var body = document.createElement("div");
      body.className = "pkg-body";
      body.style.display = "none";
      (l.questions || []).forEach(function(q){
        var it = document.createElement("div");
        it.className = "q-item";
        it.textContent = q.id + " · " + (q.title || q.file);
        it.onclick = function(){ selectQuestion({id: q.id}); };
        body.appendChild(it);
      });
      sb.appendChild(body);
    });
  });
}

// 选择模式切换
function toggleSelectMode(){
  state.selectMode = !state.selectMode;
  var btn = qs("#selectBtn");
  btn.classList.toggle("active", state.selectMode);
  btn.textContent = state.selectMode ? "☑ 选择中" : "☑ 选择";
  qs("#saveListBtn").style.display = state.selectMode ? "" : "none";
  qs("#exportBtn").style.display = state.selectMode ? "" : "none";
  renderMain();
}

function toggleSelect(q){
  var key = q.id;
  if (state.selected[key]) delete state.selected[key];
  else state.selected[key] = true;
}

// 选中题目存为清单（组卷）
function saveSelectedAsList(){
  var ids = Object.keys(state.selected);
  if (ids.length === 0) { alert("请先勾选题目"); return; }
  var name = prompt("清单名称（如：期中复习卷）");
  if (!name || !name.trim()) return;
  fetch("/api/lists/save", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({bank: state.bank, name: name.trim(), ids: ids})
  }).then(function(r){ return r.json(); }).then(function(d){
    if (d.ok) { alert("✅ 已保存清单「" + name.trim() + "」（" + ids.length + " 题）"); }
  });
}

// 导出弹窗
function openExport(){
  var ids = Object.keys(state.selected);
  if (ids.length === 0) { alert("请先勾选题目"); return; }
  qs("#exportCount").textContent = "（" + ids.length + " 道）";
  qs("#exportMsg").textContent = "";
  qs("#exportModal").classList.add("show");
}
function closeExport(){ qs("#exportModal").classList.remove("show"); }
function doExport(){
  var ids = Object.keys(state.selected);
  var parts = [];
  qs("#exportParts").querySelectorAll("input:checked").forEach(function(cb){ parts.push(cb.value); });
  var format = qs("#exportFormat").value;
  if (parts.length === 0) { qs("#exportMsg").textContent = "请至少勾选一个部分"; return; }
  fetch("/api/export", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({bank: state.bank, ids: ids, parts: parts, format: format})
  }).then(function(r){ return r.json(); }).then(function(d){
    if (d.ok) qs("#exportMsg").textContent = "✅ 已导出 " + d.count + " 题 → " + d.path;
    else qs("#exportMsg").textContent = "❌ " + (d.error || "导出失败");
  });
}

function renderSidebar(){
  var sb = qs("#sidebar");
  if (state.sideTab === "fav") { renderFavSidebar(sb); return; }
  if (state.sideTab === "lists") { renderListsSidebar(sb); return; }
  sb.innerHTML = "";
  state.tree.forEach(function(pkg, i){
    var head = document.createElement("div");
    head.className = "pkg-head";
    head.innerHTML = (state.expanded[pkg.name] ? "▾" : "▸") + " " + esc(pkg.name) + ' <span class="cnt">' + pkg.count + "</span>";
    head.onclick = function(){
      state.expanded[pkg.name] = !state.expanded[pkg.name];
      renderSidebar();
    };
    sb.appendChild(head);
    if (state.expanded[pkg.name]) {
      var body = document.createElement("div");
      body.className = "pkg-body";
      pkg.questions.forEach(function(q){
        var it = document.createElement("div");
        it.className = "q-item";
        it.textContent = q.id + " · " + (q.title ? q.title : q.file);
        it.title = q.file;
        it.onclick = function(){
          selectQuestion(q);
          document.querySelectorAll(".q-item").forEach(function(x){ x.classList.remove("active"); });
          it.classList.add("active");
        };
        body.appendChild(it);
      });
      sb.appendChild(body);
    }
  });
  if (state.tree.length === 0) {
    sb.innerHTML = '<div class="empty">题库为空</div>';
  }
}

function aggregateTags(){
  var map = {};
  state.tree.forEach(function(pkg){
    pkg.questions.forEach(function(q){
      var keys = ["chapter","grade","difficulty","importance","source","knowledge","type"];
      keys.forEach(function(k){
        var v = q.meta[k];
        if (v) {
          if (!map[k]) map[k] = {};
          map[k][v] = true;
        }
      });
    });
  });
  return map;
}

function renderFilters(){
  var bar = qs("#filters");
  bar.innerHTML = "";
  var map = aggregateTags();
  Object.keys(map).forEach(function(k){
    var f = document.createElement("span");
    f.className = "f-item";
    f.innerHTML = esc(tagName(k)) + " ";
    var sel = document.createElement("select");
    var opt = document.createElement("option");
    opt.value = ""; opt.text = "全部";
    sel.appendChild(opt);
    Object.keys(map[k]).sort().forEach(function(v){
      var o = document.createElement("option");
      o.value = v; o.text = v;
      if (state.filters[k] === v) o.selected = true;
      sel.appendChild(o);
    });
    f.appendChild(sel);
    sel.onchange = function(){
      var v = this.value;
      if (v) state.filters[k] = v; else delete state.filters[k];
      renderMain();
    };
    bar.appendChild(f);
  });
  if (Object.keys(map).length > 0) {
    var btn = document.createElement("button");
    btn.className = "clear";
    btn.textContent = "清空筛选";
    btn.onclick = function(){ state.filters = {}; renderFilters(); renderMain(); };
    bar.appendChild(btn);
  }
}

function visibleQuestions(){
  var out = [];
  state.tree.forEach(function(pkg){
    pkg.questions.forEach(function(q){
      if (state.favOnly && !isFav(q)) return;
      var hit = true;
      for (var k in state.filters) {
        if (q.meta[k] !== state.filters[k]) { hit = false; break; }
      }
      if (hit) out.push({ pkg: pkg.name, q: q });
    });
  });
  return out;
}

// ---------- V1.4.0：显示清单（自定义显示字段） ----------

function defaultDisplay(){ return ["type","difficulty","importance","source"]; }
function loadDisplay(){
  try {
    var d = JSON.parse(localStorage.getItem("reqDisplay"));
    if (d && d.length) return d;
  } catch(e){}
  return defaultDisplay();
}

// 打开显示清单面板（字段来自全局配置 meta_fields + 常见扩展）
function openDisplay(){
  fetch("/api/config/global").then(function(r){ return r.json(); }).then(function(g){
    var fields = (g.meta_fields || []).map(function(f){ return f.name; });
    ["chapter","grade","knowledge"].forEach(function(k){ if (fields.indexOf(k) < 0) fields.push(k); });
    var cur = loadDisplay();
    var box = qs("#displayList");
    box.innerHTML = "";
    fields.forEach(function(f){
      var lab = document.createElement("label");
      var cb = document.createElement("input");
      cb.type = "checkbox";
      cb.value = f;
      cb.checked = cur.indexOf(f) >= 0;
      lab.appendChild(cb);
      lab.appendChild(document.createTextNode(tagName(f)));
      box.appendChild(lab);
    });
    qs("#displayModal").classList.add("show");
  });
}
function closeDisplay(){ qs("#displayModal").classList.remove("show"); }
function saveDisplay(){
  var sel = [];
  qs("#displayList").querySelectorAll("input:checked").forEach(function(cb){ sel.push(cb.value); });
  localStorage.setItem("reqDisplay", JSON.stringify(sel));
  closeDisplay();
  renderMain();
}

function metaTagsOf(meta){
  var html = "";
  loadDisplay().forEach(function(k){
    var v = meta[k];
    if (v) html += '<span class="tag">' + esc(tagName(k)) + ": " + esc(v) + "</span>";
  });
  return html;
}

// ---------- V1.4.0：题目盒子 / 双栏 / 卡片 ----------

// 题目盒子：默认只显示题干 + 展开按钮；答案/解析/备注分段折叠
function buildBox(it, opts){
  opts = opts || {};
  var q = it.q;
  var box = document.createElement("div");
  box.className = "qbox" + (opts.active ? " active" : "");
  box.setAttribute("data-id", q.id);
  var meta = metaTagsOf(q.meta);
  var cbHtml = "";
  if (state.selectMode) cbHtml = '<input type="checkbox" class="sel-cb" ' + (state.selected[q.id] ? "checked" : "") + '>';
  box.innerHTML =
    '<div class="qbox-head">' + cbHtml + '<span class="qbox-id">' + esc(q.id) + '</span>' +
    '<span class="qbox-tags">' + meta + '</span>' +
    '<span class="qbox-actions"><button class="fav-btn" title="收藏">' + (isFav(q) ? "⭐" : "☆") + '</button>' +
    '<button class="exp-btn">▼ 展开</button></span></div>' +
    '<div class="qbox-prompt content">' + esc(q.prompt || q.title) + '</div>' +
    '<div class="qbox-detail" hidden></div>';
  box.querySelector(".fav-btn").onclick = function(e){
    e.stopPropagation();
    toggleFav(q, this);
  };
  var selCb = box.querySelector(".sel-cb");
  if (selCb) {
    selCb.onclick = function(e){ e.stopPropagation(); toggleSelect(q); };
  }
  box.querySelector(".exp-btn").onclick = function(e){
    e.stopPropagation();
    toggleExpand(box, q);
  };
  if (opts.onclick) box.onclick = function(){ opts.onclick(box, q); };
  renderMath(box);
  return box;
}

// 展开/收起题目详情（答案/解析/备注分段，各自点击显示）
function toggleExpand(box, q){
  var det = box.querySelector(".qbox-detail");
  var btn = box.querySelector(".exp-btn");
  if (det.hidden) {
    det.hidden = false;
    btn.textContent = "▲ 收起";
    if (det.getAttribute("data-loaded") !== "1") {
      det.innerHTML = '<span class="tip">加载中…</span>';
      fetch("/api/question?bank=" + encodeURIComponent(state.bank) + "&id=" + encodeURIComponent(q.id)).then(function(r){ return r.json(); }).then(function(d){
        var html = '<div class="q-actions"><button id="btnOpen">📂 打开本地</button><button id="btnEdit">✏️ 编辑</button></div>';
        var secs = [["答案", d.answer], ["解析", d.explain], ["备注", d.note]];
        secs.forEach(function(s){
          if (s[1]) html += '<div class="sec"><button class="sec-btn" data-title="' + esc(s[0]) + '">▸ ' + esc(s[0]) + '</button><div class="sec-body" hidden><div class="content">' + esc(s[1]) + "</div></div></div>";
        });
        if (!html) html = '<span class="tip">（无更多内容）</span>';
        det.innerHTML = html;
        det.querySelector("#btnOpen").onclick = function(){ openLocal(q); };
        det.querySelector("#btnEdit").onclick = function(){ openEdit(d, q); };
        det.querySelectorAll(".sec-btn").forEach(function(sb){
          sb.onclick = function(e){
            e.stopPropagation(); // 阻止冒泡到盒子（避免收起详情）
            var body = sb.parentElement.querySelector(".sec-body");
            body.hidden = !body.hidden;
            sb.textContent = (body.hidden ? "▸ " : "▾ ") + sb.getAttribute("data-title");
          };
        });
        renderMath(det);
        det.setAttribute("data-loaded", "1");
      });
    }
  } else {
    det.hidden = true;
    btn.textContent = "▼ 展开";
  }
}

// 列表模式：题目盒子流
function renderList(){
  var main = qs("#mainContent");
  main.innerHTML = "";
  var items = visibleQuestions();
  if (items.length === 0) {
    main.innerHTML = '<div class="empty">没有符合条件的题目' + (state.favOnly ? "（收藏中）" : "") + "</div>";
    return;
  }
  items.forEach(function(it){
    var box = buildBox(it, {onclick: function(b, q){ toggleExpand(b, q); }});
    main.appendChild(box);
  });
}

// 双栏模式：左列表 + 右详情
function renderSplit(){
  var main = qs("#mainContent");
  main.innerHTML = "";
  var items = visibleQuestions();
  if (items.length === 0) {
    main.innerHTML = '<div class="empty">没有符合条件的题目</div>';
    return;
  }
  var wrap = document.createElement("div");
  wrap.className = "split-wrap";
  var left = document.createElement("div");
  left.className = "split-left";
  var right = document.createElement("div");
  right.className = "split-right";
  right.innerHTML = '<span class="tip">← 点击左侧题目查看详情</span>';
  var selKey = state.splitSel;
  items.forEach(function(it, i){
    var q = it.q;
    var active = (selKey === (it.pkg + "|" + q.id));
    if (i === 0 && !selKey) active = true;
    var box = buildBox(it, {active: active});
    box.querySelector(".exp-btn").style.display = "none";
    box.onclick = function(){
      state.splitSel = it.pkg + "|" + q.id;
      document.querySelectorAll(".split-left .qbox").forEach(function(b){ b.classList.remove("active"); });
      box.classList.add("active");
      loadSplitDetail(right, q);
    };
    left.appendChild(box);
    if (active) {
      state.splitSel = it.pkg + "|" + q.id;
      loadSplitDetail(right, q);
    }
  });
  wrap.appendChild(left);
  wrap.appendChild(right);
  main.appendChild(wrap);
}

// 双栏右侧详情（完整展示，公式渲染）
function loadSplitDetail(right, q){
  fetch("/api/question?bank=" + encodeURIComponent(state.bank) + "&id=" + encodeURIComponent(q.id)).then(function(r){ return r.json(); }).then(function(d){
    var html = '<div class="q-actions"><button id="btnOpen">📂 打开本地</button><button id="btnEdit">✏️ 编辑</button></div>';
    html += '<div class="qbox-head"><span class="qbox-id">' + esc(q.id) + "</span><span style='flex:1'></span>";
    html += '<button class="fav-btn">' + (isFav(q) ? "⭐" : "☆") + "</button></div>";
    html += metaTagsOf(d.meta || {});
    if (d.prompt) html += "<div><b>题目</b><div class='content'>" + esc(d.prompt) + "</div></div>";
    if (d.answer) html += "<div><b>答案</b><div class='content'>" + esc(d.answer) + "</div></div>";
    if (d.explain) html += "<div><b>解析</b><div class='content'>" + esc(d.explain) + "</div></div>";
    if (d.note) html += "<div><b>备注</b><div class='content'>" + esc(d.note) + "</div></div>";
    right.innerHTML = html;
    var fb = right.querySelector(".fav-btn");
    if (fb) fb.onclick = function(){ toggleFav(q, this); };
    var bo = right.querySelector("#btnOpen");
    if (bo) bo.onclick = function(){ openLocal(q); };
    var be = right.querySelector("#btnEdit");
    if (be) be.onclick = function(){ openEdit(d, q); };
    renderMath(right);
  });
}

// 卡片模式：单题 + 前进后退
function renderCard(){
  var main = qs("#mainContent");
  main.innerHTML = "";
  var items = visibleQuestions();
  if (items.length === 0) {
    main.innerHTML = '<div class="empty">没有符合条件的题目</div>';
    return;
  }
  if (state.cardIdx >= items.length) state.cardIdx = 0;
  var it = items[state.cardIdx];
  var q = it.q;
  var box = buildBox(it);
  toggleExpand(box, q); // V1.4.2：卡片模式自动一级展开
  var nav = document.createElement("div");
  nav.className = "card-nav";
  var prev = document.createElement("button"); prev.textContent = "◀ 上一题";
  var cnt = document.createElement("span"); cnt.textContent = (state.cardIdx + 1) + " / " + items.length;
  var next = document.createElement("button"); next.textContent = "下一题 ▶";
  prev.onclick = function(){ state.cardIdx = (state.cardIdx - 1 + items.length) % items.length; renderCard(); };
  next.onclick = function(){ state.cardIdx = (state.cardIdx + 1) % items.length; renderCard(); };
  nav.appendChild(prev); nav.appendChild(cnt); nav.appendChild(next);
  main.appendChild(box);
  main.appendChild(nav);
}

// V1.4.3：键盘导航上一题/下一题
function navQuestion(delta){
  var items = visibleQuestions();
  if (items.length === 0) return;
  if (state.mode === "card") {
    state.cardIdx = (state.cardIdx + delta + items.length) % items.length;
    renderCard();
  } else if (state.mode === "split") {
    var idx = 0;
    for (var i = 0; i < items.length; i++) {
      if (state.splitSel === (items[i].pkg + "|" + items[i].q.id)) { idx = i; break; }
    }
    idx = (idx + delta + items.length) % items.length;
    state.splitSel = items[idx].pkg + "|" + items[idx].q.id;
    renderSplit();
  }
}

// 选中题目（侧边栏点击）：按模式处理
function selectQuestion(q){
  if (state.mode === "card") {
    var items = visibleQuestions();
    for (var i = 0; i < items.length; i++) {
      if (items[i].q.id === q.id) { state.cardIdx = i; break; }
    }
    renderCard();
  } else if (state.mode === "split") {
    state.splitSel = q.id;
    renderSplit();
  } else {
    var box = document.querySelector('.qbox[data-id="' + q.id + '"]');
    if (box) {
      box.scrollIntoView({block: "center"});
      toggleExpand(box, q);
    }
  }
}

function objectLen(obj){ var n = 0; for (var k in obj) n++; return n; }

function loadDetail(q, det, card){
  fetch("/api/question?bank=" + encodeURIComponent(state.bank) + "&id=" + encodeURIComponent(q.id)).then(function(r){ return r.json(); }).then(function(d){
    var html = '<div class="q-actions">';
    html += '<button id="btnOpen">📂 打开本地</button>';
    html += '<button id="btnEdit">✏️ 编辑</button>';
    html += '</div>';
    if (d.prompt) html += "<div><b>题目</b><div class=\"content\">" + esc(d.prompt) + "</div></div>";
    if (d.answer) html += "<div><b>答案</b><div class=\"content\">" + esc(d.answer) + "</div></div>";
    if (d.explain) html += "<div><b>解析</b><div class=\"content\">" + esc(d.explain) + "</div></div>";
    if (d.note) html += "<div><b>备注</b><div class=\"content\">" + esc(d.note) + "</div></div>";
    if (!html) html = '<span class="tip">（无更多内容）</span>';
    det.innerHTML = html;
    det.querySelector("#btnOpen").onclick = function(){ openLocal(q); };
    det.querySelector("#btnEdit").onclick = function(){ openEdit(d, q); };
    renderMath(det);
    card.classList.add("active");
  }).catch(function(){ det.innerHTML = '<span class="tip">加载失败</span>'; });
}

// 刷新题库：重新扫描目录（同步本地/网页改动）
function reloadBank(){
  fetch("/api/reload", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({bank: state.bank})
  }).then(function(r){ return r.json(); }).then(function(d){
    if (d.ok) loadAll();
  });
}

// 打开本地文件（资源管理器定位）
function openLocal(q){
  fetch("/api/open", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({bank: state.bank, id: q.id})
  }).then(function(r){ return r.json(); }).then(function(d){
    if (!d.ok) alert(d.error || "打开失败");
  });
}

// 编辑弹窗
var editing = null;
function openEdit(d, q){
  editing = q;
  qs("#editId").textContent = q.id;
  qs("#editFile").value = q.file;
  qs("#editPrompt").value = d.prompt || "";
  qs("#editAnswer").value = d.answer || "";
  qs("#editExplain").value = d.explain || "";
  qs("#editNote").value = d.note || "";
  qs("#editMsg").textContent = "";
  // 异步加载字段已知值后构建元数据行
  fetch("/api/meta-values?bank=" + encodeURIComponent(state.bank)).then(function(r){ return r.json(); }).then(function(values){
    buildMetaRows(d.meta || {}, values);
  });
  qs("#editModal").classList.add("show");
}
function buildMetaRows(meta, values){
  var metaBox = qs("#editMeta");
  metaBox.innerHTML = "";
  var keys = ["id","chapter","grade","difficulty","importance","source","knowledge","type"];
  var shown = {};
  keys.forEach(function(k){ if (meta[k]) { addMetaRow(metaBox, k, meta[k], (values||{})[k] || []); shown[k] = true; } });
  keys.forEach(function(k){ if (!shown[k]) addMetaRow(metaBox, k, "", (values||{})[k] || []); });
  Object.keys(meta).forEach(function(k){
    if (keys.indexOf(k) < 0 && ["app","bank","path"].indexOf(k) < 0) addMetaRow(metaBox, k, meta[k], (values||{})[k] || []);
  });
}
function addMetaRow(box, k, v, knownVals){
  var row = document.createElement("div");
  row.className = "meta-row";
  var lab = document.createElement("span");
  lab.className = "lbl";
  lab.textContent = tagName(k);
  lab.style.paddingTop = "5px";
  var sel = document.createElement("select");
  sel.setAttribute("data-key", k);
  var opt = document.createElement("option"); opt.value = ""; opt.text = "（空）";
  sel.appendChild(opt);
  (knownVals || []).forEach(function(vv){
    var o = document.createElement("option"); o.value = vv; o.text = vv;
    if (vv === v) o.selected = true;
    sel.appendChild(o);
  });
  // 当前值不在已知列表 → 显示为「值(自定义)」并选中
  if (v && (knownVals||[]).indexOf(v) < 0) {
    var co = document.createElement("option"); co.value = v; co.text = v + "(自定义)"; co.selected = true;
    sel.appendChild(co);
  }
  // 新增值 / 自定义 两个特殊选项
  var addOpt = document.createElement("option"); addOpt.value = "__add__"; addOpt.text = "➕ 新增值…";
  sel.appendChild(addOpt);
  var onceOpt = document.createElement("option"); onceOpt.value = "__once__"; onceOpt.text = "✍️ 自定义…";
  sel.appendChild(onceOpt);
  var cdiv = document.createElement("span");
  cdiv.className = "cust";
  cdiv.style.display = "none";
  var inp = document.createElement("input");
  var hint = document.createElement("span");
  hint.className = "hint";
  cdiv.appendChild(inp); cdiv.appendChild(hint);
  sel.onchange = function(){
    if (this.value === "__add__") {
      cdiv.style.display = "flex"; inp.value = ""; hint.textContent = "将加入题库配置";
    } else if (this.value === "__once__") {
      cdiv.style.display = "flex"; inp.value = ""; hint.textContent = "仅用于本题";
    } else {
      cdiv.style.display = "none"; inp.value = "";
    }
  };
  // 删除字段按钮
  var del = document.createElement("button");
  del.className = "meta-del"; del.textContent = "✕"; del.title = "删除该字段";
  del.onclick = function(){ row.remove(); };
  row.appendChild(lab); row.appendChild(sel); row.appendChild(del); row.appendChild(cdiv);
  box.appendChild(row);
}
function addField(){
  var box = qs("#editMeta");
  var existing = qs("#newFieldRow");
  if (existing) { existing.querySelector("input").focus(); return; }
  // 在编辑栏内展开输入行（不用浏览器 prompt 弹窗）
  var row = document.createElement("div");
  row.id = "newFieldRow";
  row.style.display = "flex";
  row.style.gap = "4px";
  row.style.alignItems = "center";
  var inp = document.createElement("input");
  inp.placeholder = "输入新字段名（如：出处）";
  var ok = document.createElement("button"); ok.textContent = "添加";
  var cancel = document.createElement("button"); cancel.textContent = "取消";
  row.appendChild(inp); row.appendChild(ok); row.appendChild(cancel);
  box.appendChild(row);
  inp.focus();
  function commit(){
    var name = inp.value.trim();
    if (!name) return;
    var exist = false;
    box.querySelectorAll("select[data-key]").forEach(function(s){ if (s.getAttribute("data-key") === name) exist = true; });
    if (exist) { alert("字段已存在"); return; }
    addMetaRow(box, name, "", []);
    row.remove();
  }
  ok.onclick = commit;
  cancel.onclick = function(){ row.remove(); };
  inp.addEventListener("keydown", function(e){ if (e.key === "Enter") commit(); if (e.key === "Escape") row.remove(); });
}
function closeEdit(){ qs("#editModal").classList.remove("show"); }
function saveEdit(){
  if (!editing) return;
  var meta = {};
  var addToConfig = [];
  qs("#editMeta").querySelectorAll("select[data-key]").forEach(function(sel){
    var k = sel.getAttribute("data-key");
    var val = "";
    if (sel.value === "__add__" || sel.value === "__once__") {
      var inp = sel.parentElement.querySelector("input");
      val = inp ? inp.value.trim() : "";
      if (val !== "" && sel.value === "__add__") addToConfig.push({k: k, v: val});
    } else {
      val = sel.value;
    }
    if (val !== "") meta[k] = val;
  });
  // 先写入配置（新增值），再保存题目
  var doSave = function(){
    var body = {
      bank: state.bank, id: editing.id,
      file: qs("#editFile").value.trim(),
      meta: meta,
      prompt: qs("#editPrompt").value,
      answer: qs("#editAnswer").value,
      explain: qs("#editExplain").value,
      note: qs("#editNote").value
    };
    fetch("/api/question/save", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify(body)
    }).then(function(r){ return r.json(); }).then(function(d){
      if (d.ok) {
        qs("#editMsg").textContent = "✅ 已保存";
        closeEdit();
        loadAll();
      } else {
        qs("#editMsg").textContent = "❌ " + (d.error || "保存失败");
      }
    });
  };
  if (addToConfig.length > 0) {
    var pending = addToConfig.length, failed = false;
    addToConfig.forEach(function(ac){
      fetch("/api/meta-value/add", {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({bank: state.bank, field: ac.k, value: ac.v})
      }).then(function(r){ return r.json(); }).then(function(d){
        if (!d.ok) failed = true;
        if (--pending === 0) {
          if (failed) qs("#editMsg").textContent = "⚠️ 部分新值写入配置失败";
          doSave();
        }
      });
    });
  } else {
    doSave();
  }
}

function openSettings(){
  qs("#modal").classList.add("show");
  qs("#linkMsg").textContent = "";
  qs("#linkInput").value = "";
  loadCfgGlobal();
  loadCfgProject();
}
function closeSettings(){ qs("#modal").classList.remove("show"); }

// 页签切换
qs("#tabGlobal").onclick = function(){
  qs("#tabGlobal").classList.add("active");
  qs("#tabProject").classList.remove("active");
  qs("#cfgGlobal").hidden = false;
  qs("#cfgProject").hidden = true;
};
qs("#tabProject").onclick = function(){
  qs("#tabProject").classList.add("active");
  qs("#tabGlobal").classList.remove("active");
  qs("#cfgGlobal").hidden = true;
  qs("#cfgProject").hidden = false;
};

// 加载全局配置
function loadCfgGlobal(){
  Promise.all([
    fetch("/api/config/global").then(function(r){ return r.json(); }),
    fetch("/api/banks").then(function(r){ return r.json(); })
  ]).then(function(res){
    var g = res[0], banks = res[1];
    qs("#cfgGlobalPath").textContent = g.path;
    // 题库目录 → 名称映射
    var nameMap = {};
    banks.forEach(function(b){ nameMap[b.dir] = b.name; });
    // 默认配置
    var dbox = qs("#cfgDefaults");
    dbox.innerHTML = "";
    Object.keys(g.defaults || {}).forEach(function(k){
      var it = document.createElement("div");
      it.className = "cfg-item";
      it.innerHTML = "<span>" + esc(k) + "</span><span class=\"vals\">" + esc(g.defaults[k]) + "</span>";
      dbox.appendChild(it);
    });
    // 题库列表（全部对等，当前打开的不可移除）
    var lbox = qs("#cfgLinks");
    lbox.innerHTML = "";
    (g.links || []).forEach(function(l, i){
      var isCur = (l === state.bank);
      var it = document.createElement("div");
      it.className = "cfg-item";
      it.innerHTML = "<span><b>" + esc(nameMap[l] || l) + "</b> " + (isCur ? "<b style=\"color:var(--accent)\">（当前）</b>" : "") + "<div class=\"path\">" + esc(l) + "</div></span>" + (isCur ? "<span class=\"vals\">不可移除</span>" : "<span style=\"display:flex;gap:4px\"><button class=\"cfg-open\" title=\"打开该题库\">打开</button><button class=\"cfg-del\" title=\"移除\">✕</button></span>");
      if (!isCur) {
        it.querySelector(".cfg-open").onclick = function(){ openBank(l); };
        it.querySelector(".cfg-del").onclick = function(){ removeLink(i); };
      }
      lbox.appendChild(it);
    });
    if (!g.links || g.links.length === 0) lbox.innerHTML = '<span class="tip">（暂无题库，请添加）</span>';
    // 字段定义
    var fbox = qs("#cfgFields");
    fbox.innerHTML = "";
    (g.meta_fields || []).forEach(function(f){
      var it = document.createElement("div");
      it.className = "cfg-item";
      it.innerHTML = "<span>" + esc(f.label || f.name) + " <small style=\"color:var(--muted)\">" + esc(f.name) + "</small></span><span class=\"vals\">" + esc((f.values||[]).join(" / ")) + "</span>";
      fbox.appendChild(it);
    });
  });
}

// 打开题库（切换下拉栏选中）
function openBank(dir){
  var sel = qs("#bankSel");
  sel.value = dir;
  state.bank = dir;
  state.filters = {};
  closeSettings();
  loadAll();
}

// 加载项目配置
function loadCfgProject(){
  fetch("/api/config/project?bank=" + encodeURIComponent(state.bank)).then(function(r){ return r.json(); }).then(function(p){
    qs("#cfgProjectPath").textContent = p.path;
    qs("#cfgProjectInfo").innerHTML =
      "<div class=\"cfg-item\"><span>题库名</span><span class=\"vals\">" + esc(p.bank) + "</span></div>" +
      "<div class=\"cfg-item\"><span>运行软件</span><span class=\"vals\">" + esc(p.app) + "</span></div>";
    var fbox = qs("#cfgProjectFields");
    fbox.innerHTML = "";
    (p.meta_fields || []).forEach(function(f){
      var it = document.createElement("div");
      it.className = "cfg-item";
      it.innerHTML = "<span>" + esc(f.label || f.name) + "</span><span class=\"vals\">" + esc((f.values||[]).join(" / ")) + "</span>";
      fbox.appendChild(it);
    });
    if (!p.meta_fields || p.meta_fields.length === 0) fbox.innerHTML = '<span class="tip">（暂无自定义字段）</span>';
  });
}

// 解除链接（写全局配置）
function removeLink(i){
  fetch("/api/config/global").then(function(r){ return r.json(); }).then(function(g){
    var links = g.links || [];
    links.splice(i, 1);
    return fetch("/api/config/global/save", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({links: links})
    });
  }).then(function(r){ return r.json(); }).then(function(d){
    if (d.ok) {
      loadCfgGlobal();
      loadBanks();
    }
  });
}

function doLink(){
  var dir = qs("#linkInput").value.trim();
  if (!dir) { qs("#linkMsg").textContent = "请输入题库目录路径"; return; }
  fetch("/api/link", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({dir: dir})
  }).then(function(r){
    return r.json().then(function(d){ return {ok: r.ok, d: d}; });
  }).then(function(res){
    if (res.ok) {
      qs("#linkMsg").textContent = "✅ 已链接";
      closeSettings();
      loadBanks();
    } else {
      qs("#linkMsg").textContent = "❌ " + (res.d.error || "链接失败");
    }
  }).catch(function(){ qs("#linkMsg").textContent = "❌ 请求失败"; });
}

init();
`

func firstLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		s = s[:idx]
	}
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) > 40 {
		s = string(runes[:40]) + "…"
	}
	return s
}