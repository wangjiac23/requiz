// V1.0.0：Localhost Web 服务（Go 标准库，零第三方依赖）
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"
)

// cmdServe 启动 localhost Web 服务
// 用法：requiz serve [题库目录] [-port 端口]
func cmdServe(args []string) error {
	port := "8080"
	dir := "."
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-port", "--port":
			if i+1 >= len(args) {
				return fmt.Errorf("-port 需要一个端口号")
			}
			port = args[i+1]
			i++
		default:
			if strings.HasPrefix(args[i], "-") {
				return fmt.Errorf("未知选项: %s", args[i])
			}
			if dir != "." {
				return fmt.Errorf("serve 参数过多（用法：requiz serve [题库目录] [-port 端口]）")
			}
			dir = args[i]
		}
	}

	bank, err := connectBank(dir)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler(bank))
	mux.HandleFunc("/question", questionHandler(bank))
	mux.HandleFunc("/api/questions", apiQuestionsHandler(bank))
	mux.HandleFunc("/api/question", apiQuestionHandler(bank))

	addr := "127.0.0.1:" + port
	fmt.Printf("requiz web   : http://%s/\n", addr)
	fmt.Printf("题库         : %s\n", bank.Dir)
	fmt.Printf("题目数       : %d\n", len(bank.Questions))
	fmt.Println("按 Ctrl+C 停止服务")
	return http.ListenAndServe(addr, mux)
}

// ---------- 视图模型 ----------

type questView struct {
	ID     string
	File   string
	Rel    string
	Title  string
	Prompt string
}

type bankView struct {
	BankName  string
	Dir       string
	Count     int
	App       string
	Version   string
	Questions []questView
}

type questDetailView struct {
	questView
	Meta    map[string]string
	Answer  string
	Explain string
	Note    string
	Extra   map[string]string
}

// ---------- 模板 ----------

var indexTpl = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html lang="zh">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>requiz — {{.BankName}}</title>
<style>
 body{font-family:-apple-system,"Segoe UI","Microsoft YaHei",sans-serif;max-width:800px;margin:0 auto;padding:24px;background:#f7f8fa;color:#24292f}
 h1 small{font-weight:normal;color:#888;font-size:0.6em}
 .card{background:#fff;border:1px solid #e1e4e8;border-radius:8px;padding:16px;margin:8px 0}
 .meta{color:#586069;font-size:0.9em}
 a{color:#0969da;text-decoration:none} a:hover{text-decoration:underline}
 li{margin:4px 0}
</style>
</head>
<body>
<h1>📚 requiz <small>v{{.Version}}</small></h1>
<div class="card meta">题库：<b>{{.BankName}}</b>（{{.Dir}}）· 题目 <b>{{.Count}}</b> 道 · 运行软件：{{.App}}</div>
{{if .Questions}}
<ul>
{{range .Questions}}<li><a href="/question?id={{.ID}}">{{.ID}}</a> — {{.File}}{{if .Prompt}} <span class="meta">· {{.Title}}</span>{{end}}</li>{{end}}
</ul>
{{else}}
<p>题库为空。</p>
{{end}}
</body>
</html>`))

var questionTpl = template.Must(template.New("question").Parse(`<!DOCTYPE html>
<html lang="zh">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.ID}} — requiz</title>
<style>
 body{font-family:-apple-system,"Segoe UI","Microsoft YaHei",sans-serif;max-width:800px;margin:0 auto;padding:24px;background:#f7f8fa;color:#24292f}
 .card{background:#fff;border:1px solid #e1e4e8;border-radius:8px;padding:16px;margin:12px 0}
 .meta{color:#586069;font-size:0.85em}
 pre{background:#f6f8fa;padding:12px;border-radius:6px;overflow-x:auto;font-size:0.9em}
 .tag{display:inline-block;background:#ddf4ff;color:#0969da;border-radius:12px;padding:2px 10px;margin:2px;font-size:0.85em}
 a{color:#0969da;text-decoration:none}
</style>
</head>
<body>
<p><a href="/">← 返回题库列表</a></p>
<h1>{{.ID}} <small class="meta">{{.File}}</small></h1>
<div class="card">
{{range $k, $v := .Meta}}<span class="tag">{{$k}}: {{$v}}</span>{{end}}
</div>
{{if .Prompt}}<div class="card"><h2>题目</h2><pre>{{.Prompt}}</pre></div>{{end}}
{{if .Answer}}<div class="card"><h2>答案</h2><pre>{{.Answer}}</pre></div>{{end}}
{{if .Explain}}<div class="card"><h2>解析</h2><pre>{{.Explain}}</pre></div>{{end}}
{{if .Note}}<div class="card"><h2>备注</h2><pre>{{.Note}}</pre></div>{{end}}
{{range $k, $v := .Extra}}<div class="card"><h2>{{$k}}</h2><pre>{{$v}}</pre></div>{{end}}
</body>
</html>`))

// ---------- Handlers ----------

func indexHandler(bank *Bank) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		qs := make([]questView, 0, len(bank.Questions))
		for _, q := range bank.Questions {
			qs = append(qs, questView{
				ID:     q.ID(),
				File:   q.File,
				Prompt: q.Prompt,
				Title:  firstLine(q.Prompt),
			})
		}
		sort.Slice(qs, func(i, j int) bool { return qs[i].ID < qs[j].ID })
		data := bankView{
			BankName:  bank.Name,
			Dir:       bank.Dir,
			Count:     len(qs),
			App:       bank.App,
			Questions: qs,
			Version:   version,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		indexTpl.Execute(w, data)
	}
}

func questionHandler(bank *Bank) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "缺少参数 id（?id=M001）", http.StatusBadRequest)
			return
		}
		q, err := bank.find(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		data := questDetailView{
			questView: questView{
				ID:     q.ID(),
				File:   q.File,
				Prompt: q.Prompt,
				Title:  firstLine(q.Prompt),
			},
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
	}
}

func apiQuestionsHandler(bank *Bank) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list := make([]questionJSON, 0, len(bank.Questions))
		for _, q := range bank.Questions {
			list = append(list, toJSON(q))
		}
		writeJSON(w, list)
	}
}

func apiQuestionHandler(bank *Bank) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, `{"error":"缺少参数 id"}`, http.StatusBadRequest)
			return
		}
		q, err := bank.find(id)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusNotFound)
			return
		}
		writeJSON(w, toJSON(q))
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}

func firstLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		s = s[:idx]
	}
	// 去掉行内 Markdown 标记，取前 40 字
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) > 40 {
		s = string(runes[:40]) + "…"
	}
	return s
}