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
	Rel    string            `json:"rel"`    // V1.7.0 相对路径（目录筛选）
	Meta   map[string]string `json:"meta"`
}

// treeNode 目录树节点（V1.7.0 任意层级文件夹导航）
type treeNode struct {
	Name      string         `json:"name"`
	Path      string         `json:"path"`
	Dirs      []*treeNode    `json:"dirs"`
	Questions []questSummary `json:"questions"`
}

type treeResp struct {
	Root *treeNode `json:"root"`
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

// treeOf 构建嵌套目录树（任意层级，按文件系统结构）
func treeOf(b *model.Bank) treeResp {
	root := &treeNode{Name: b.Name, Path: "", Dirs: []*treeNode{}, Questions: []questSummary{}}
	for _, q := range b.Questions {
		rel := relPath(b, q)
		parts := strings.Split(rel, "/")
		dirs := parts[:len(parts)-1]
		node := root
		cur := ""
		for _, d := range dirs {
			cur = joinPath(cur, d)
			node = getOrCreateDir(node, d, cur)
		}
		s := summaryOf(b, q)
		s.Rel = rel
		node.Questions = append(node.Questions, s)
	}
	sortTree(root)
	return treeResp{Root: root}
}

func joinPath(a, b string) string {
	if a == "" {
		return b
	}
	return a + "/" + b
}

func getOrCreateDir(node *treeNode, name, path string) *treeNode {
	for _, d := range node.Dirs {
		if d.Name == name {
			return d
		}
	}
	d := &treeNode{Name: name, Path: path, Dirs: []*treeNode{}, Questions: []questSummary{}}
	node.Dirs = append(node.Dirs, d)
	return d
}

func sortTree(n *treeNode) {
	sort.Slice(n.Dirs, func(i, j int) bool { return n.Dirs[i].Name < n.Dirs[j].Name })
	sort.Slice(n.Questions, func(i, j int) bool { return n.Questions[i].ID < n.Questions[j].ID })
	for _, d := range n.Dirs {
		sortTree(d)
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
	Links   string            `json:"links,omitempty"` // V2.2.0
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
		Links:   q.Links,
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
