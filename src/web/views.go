// 由 server.go 拆分生成（V1.6.0 模块化，功能零变化）
package web

import (
	"requiz/src/model"
	"encoding/json"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
)

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

func relPath(b *model.Bank, q *model.Question) string {
	rel, err := filepath.Rel(b.Dir, q.Path)
	if err != nil {
		return q.File
	}
	return filepath.ToSlash(rel)
}


func summaryOf(b *model.Bank, q *model.Question) questSummary {
	return questSummary{
		ID:     q.ID(),
		File:   q.File,
		Title:  firstLine(q.Prompt),
		Prompt: q.Prompt,
		Meta:   q.Meta,
	}
}

// treeOf 把题库题目按顶层目录分组为题目包树

func treeOf(b *model.Bank) []pkgJSON {
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


func toJSON(q *model.Question) questionJSON {
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
