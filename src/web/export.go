// 由 server.go 拆分生成（V1.6.0 模块化，功能零变化）
package web

import (
	"requiz/src/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func apiExportHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "仅支持 POST", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Bank string   `json:"bank"`
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
		qs := []*model.Question{}
		for _, id := range body.IDs {
			if q, err := b.Find(id); err == nil {
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

// POST /api/export/open {path}：explorer 定位导出文件（仅限 output/ 目录内）

func apiExportOpenHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "仅支持 POST", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Path   string `json:"path"`
			Action string `json:"action"` // select=资源管理器定位 / open=默认程序打开
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil || body.Path == "" {
			http.Error(w, `{"error":"需要 path 字段"}`, http.StatusBadRequest)
			return
		}
		abs, err := filepath.Abs(body.Path)
		if err != nil {
			http.Error(w, `{"error":"路径解析失败"}`, http.StatusInternalServerError)
			return
		}
		outAbs, _ := filepath.Abs("output")
		rel, err := filepath.Rel(outAbs, abs)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			http.Error(w, `{"error":"只能操作 output 目录内的文件"}`, http.StatusForbidden)
			return
		}
		var cmd *exec.Cmd
		if body.Action == "open" {
			cmd = exec.Command("cmd", "/c", "start", "", abs) // 默认程序打开
		} else {
			cmd = exec.Command("explorer.exe", "/select,", abs) // 资源管理器定位
		}
		if err := cmd.Start(); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]bool{"ok": true})
	}
}

// exportJSON 导出题目为 JSON

func exportJSON(qs []*model.Question, parts []string) string {
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
// KaTeX JS 内嵌（file:// 打开时外部 JS 被安全策略禁止），公式自包含渲染

func exportHTML(qs []*model.Question, parts []string, bankName string) string {
	kd := katexDir()
	katexJS, _ := os.ReadFile(filepath.Join(kd, "katex.min.js"))
	autoJS, _ := os.ReadFile(filepath.Join(kd, "contrib", "auto-render.min.js"))
	partName := map[string]string{"prompt": "题目", "answer": "答案", "explain": "解析", "note": "备注"}
	var b strings.Builder
	b.WriteString("<!DOCTYPE html><html lang=\"zh\"><head><meta charset=\"utf-8\"><title>requiz 导出</title>")
	b.WriteString("<link rel=\"stylesheet\" href=\"../web/katex/katex.min.css\">")
	b.WriteString("<style>body{font-family:'Microsoft YaHei',sans-serif;max-width:900px;margin:0 auto;padding:30px;color:#222}.q{border:1px solid #ccc;border-radius:8px;padding:14px 18px;margin:14px 0;page-break-inside:avoid}h1{font-size:20px}h2{font-size:16px;margin:0 0 8px}.content{white-space:pre-wrap;margin:6px 0;line-height:1.8}.sec-title{font-weight:bold;margin-top:8px;color:#555}@media print{body{padding:10px}.q{border:none}}</style></head><body>")
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
				fmt.Fprintf(&b, "<div class=\"sec-title\">%s</div><div class=\"content\">%s</div>", partName[p], htmlEscape(content))
			}
		}
		b.WriteString("</div>")
	}
	// KaTeX JS 内嵌（file:// 下外部 JS 不加载），CSS 用相对引用
	if len(katexJS) > 0 {
		b.WriteString("<script>")
		b.Write(katexJS)
		b.WriteString("</script>\n")
	}
	if len(autoJS) > 0 {
		b.WriteString("<script>")
		b.Write(autoJS)
		b.WriteString("</script>\n")
	}
	b.WriteString("<script>window.addEventListener('DOMContentLoaded',function(){if(window.renderMathInElement){renderMathInElement(document.body,{delimiters:[{left:'$$',right:'$$',display:true},{left:'$',right:'$',display:false},{left:'\\\\\\(',right:'\\\\\\)',display:false},{left:'\\\\[',right:'\\\\]',display:true}],throwOnError:false});}});</script>")
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

