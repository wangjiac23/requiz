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
)

// ---------- Store：多题库状态（主题库 + 链接题库） ----------

type Store struct {
	mu    sync.Mutex
	main  *Bank
	links []*Bank
}

func newStore(main *Bank) *Store {
	s := &Store{main: main}
	for _, linkDir := range main.Links {
		if b, err := connectBank(linkDir); err == nil {
			s.links = append(s.links, b)
		}
	}
	return s
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

// addLink 连接新题库并持久化到主题库 .requiz/config.yaml
func (s *Store) addLink(dir string) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	for _, b := range s.banks() {
		if b.Dir == abs {
			return nil // 已链接
		}
	}
	b, err := connectBank(abs)
	if err != nil {
		return err
	}
	if err := s.persistLink(abs); err != nil {
		return err
	}
	s.mu.Lock()
	s.links = append(s.links, b)
	s.mu.Unlock()
	return nil
}

// persistLink 把链接题库写入主题库配置的 links 节
func (s *Store) persistLink(dir string) error {
	cfg := filepath.Join(s.main.Dir, ".requiz", "config.yaml")
	text := ""
	if data, err := os.ReadFile(cfg); err == nil {
		text = string(data)
	}
	lines := strings.Split(text, "\n")
	entry := "  - " + dir
	// 去重
	for _, l := range lines {
		if strings.TrimSpace(l) == strings.TrimSpace(entry) {
			return nil
		}
	}
	linksIdx := -1
	for i, l := range lines {
		if strings.TrimSpace(l) == "links:" {
			linksIdx = i
			break
		}
	}
	if linksIdx < 0 {
		lines = append(lines, "", "links:", entry)
	} else {
		// 在 links 节之后、遇到下一非注释/非列表行前插入
		insertAt := len(lines)
		for i := linksIdx + 1; i < len(lines); i++ {
			t := strings.TrimSpace(lines[i])
			if t != "" && !strings.HasPrefix(t, "-") && !strings.HasPrefix(t, "#") {
				insertAt = i
				break
			}
		}
		newLines := append([]string{}, lines[:insertAt]...)
		newLines = append(newLines, entry)
		newLines = append(newLines, lines[insertAt:]...)
		lines = newLines
	}
	return os.WriteFile(cfg, []byte(strings.Join(lines, "\n")), 0644)
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
	mux.Handle("/katex/", http.StripPrefix("/katex/", http.FileServer(http.Dir(katexDir()))))

	addr := "127.0.0.1:" + port
	fmt.Printf("requiz web   : http://%s/\n", addr)
	fmt.Printf("主题库       : %s（%d 题）\n", bank.Dir, len(bank.Questions))
	for _, b := range store.banks()[1:] {
		fmt.Printf("链接题库     : %s（%d 题）\n", b.Dir, len(b.Questions))
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
			Bank    string `json:"bank"`
			ID      string `json:"id"`
			Prompt  string `json:"prompt"`
			Answer  string `json:"answer"`
			Explain string `json:"explain"`
			Note    string `json:"note"`
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

type questSummary struct {
	ID    string            `json:"id"`
	File  string            `json:"file"`
	Title string            `json:"title"`
	Meta  map[string]string `json:"meta"`
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
		ID:    q.ID(),
		File:  q.File,
		Title: firstLine(q.Prompt),
		Meta:  q.Meta,
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
      <div id="sidebarHead"><span>题目导航</span><button id="pinSidebar" title="固定侧边栏（固定时拖拽只调宽不隐藏）">📌</button></div>
      <aside id="sidebar"></aside>
    </div>
    <div id="resizer" title="拖拽调整宽度（拖到最窄隐藏）"></div>
  </div>
  <main id="main"><div class="empty">加载中…</div></main>
</div>
<div id="modal" hidden>
  <div class="modal-box">
    <h3>链接题库 🔗</h3>
    <p class="tip">输入题库目录路径（相对主题库或绝对路径），将写入主题库的 .requiz/config.yaml</p>
    <input id="linkInput" type="text" placeholder="例如 ../题库B 或 D:\xxx\题库B">
    <div class="modal-actions">
      <button id="linkOk">链接</button>
      <button id="linkCancel">取消</button>
    </div>
    <div id="linkMsg" class="tip"></div>
  </div>
</div>
<div id="editModal">
  <div class="modal-box">
    <h3>编辑题目 <span id="editId" class="meta"></span></h3>
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
html,body{margin:0;height:100%}
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
#body{flex:1;display:flex;overflow:hidden}
#sidebarWrap{display:flex}
#sidebarWrap.hidden{display:none}
#sidebarCol{display:flex;flex-direction:column;min-width:0;border-right:1px solid var(--border)}
#sidebarHead{display:flex;align-items:center;justify-content:space-between;padding:4px 8px;font-size:12px;color:var(--muted);background:var(--sidebar);border-bottom:1px solid var(--border)}
#sidebarHead button{padding:2px 6px;font-size:12px}
#sidebar{flex:1;width:260px;min-width:120px;max-width:480px;background:var(--sidebar);overflow-y:auto;padding:8px 6px}
#resizer{width:5px;cursor:col-resize;flex-shrink:0;background:transparent}
#resizer:hover,#resizer.active{background:var(--accent-bg)}
button.pinned{background:var(--accent-bg);color:var(--accent);border-color:var(--accent)}
#main{flex:1;overflow-y:auto;padding:16px}
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
#editModal .modal-box{width:560px;max-width:92%}
#editModal textarea{width:100%;font-family:inherit;font-size:13px;padding:6px;border:1px solid var(--border);border-radius:6px;margin:2px 0 8px;resize:vertical}
#editModal .lbl{font-size:12px;color:var(--muted)}
.q-actions{display:flex;gap:8px;margin-bottom:10px}
.q-actions button{font-size:12px;padding:4px 10px}
#reloadBtn{padding:6px 10px}
.modal-box{background:var(--panel);border-radius:10px;padding:20px;width:420px;max-width:90%}
.modal-box h3{margin:0 0 8px}
.tip{color:var(--muted);font-size:12px}
#linkInput{width:100%;padding:8px;margin:10px 0;border:1px solid var(--border);border-radius:6px}
.modal-actions{display:flex;gap:8px;justify-content:flex-end}
`

const indexJS = `
var state = { banks: [], bank: "", tree: [], filters: {}, expanded: {} };

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
    renderList();
  });
}

function renderSidebar(){
  var sb = qs("#sidebar");
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
          renderList(q);
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
      renderList();
    };
    bar.appendChild(f);
  });
  if (Object.keys(map).length > 0) {
    var btn = document.createElement("button");
    btn.className = "clear";
    btn.textContent = "清空筛选";
    btn.onclick = function(){ state.filters = {}; renderFilters(); renderList(); };
    bar.appendChild(btn);
  }
}

function visibleQuestions(){
  var out = [];
  state.tree.forEach(function(pkg){
    pkg.questions.forEach(function(q){
      var hit = true;
      for (var k in state.filters) {
        if (q.meta[k] !== state.filters[k]) { hit = false; break; }
      }
      if (hit) out.push({ pkg: pkg.name, q: q });
    });
  });
  return out;
}

function renderList(only){
  var main = qs("#main");
  main.innerHTML = "";
  var items = only ? [{pkg:"", q:only}] : visibleQuestions();
  if (items.length === 0) {
    main.innerHTML = '<div class="empty">没有符合条件的题目 (' + objectLen(state.filters) + ' 个筛选条件)</div>';
    return;
  }
  items.forEach(function(it){
    var q = it.q;
    var card = document.createElement("div");
    card.className = "card";
    var meta = "";
    var keys = ["type","difficulty","importance","source"];
    keys.forEach(function(k){
      var v = q.meta[k];
      if (v) meta += '<span class="tag">' + esc(tagName(k)) + ": " + esc(v) + "</span>";
    });
    card.innerHTML = "<h3>" + esc(q.id) + (it.pkg ? ' <small style="color:#586069">' + esc(it.pkg) + "</small>" : "") + "</h3><div class='qmeta'>" + meta + "</div><div class='content'>" + esc(q.title) + "</div>";
    var det = document.createElement("div");
    det.className = "detail";
    card.appendChild(det);
    det.innerHTML = "加载详情…";
    card.onclick = function(){ loadDetail(q, det, card); };
    main.appendChild(card);
    renderMath(card);
  });
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
  qs("#editPrompt").value = d.prompt || "";
  qs("#editAnswer").value = d.answer || "";
  qs("#editExplain").value = d.explain || "";
  qs("#editNote").value = d.note || "";
  qs("#editMsg").textContent = "";
  qs("#editModal").classList.add("show");
}
function closeEdit(){ qs("#editModal").classList.remove("show"); }
function saveEdit(){
  if (!editing) return;
  var body = {
    bank: state.bank, id: editing.id,
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
}

function openSettings(){
  qs("#modal").classList.add("show");
  qs("#linkMsg").textContent = "";
  qs("#linkInput").value = "";
  qs("#linkInput").focus();
}
function closeSettings(){ qs("#modal").classList.remove("show"); }

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