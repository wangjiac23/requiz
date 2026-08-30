// 题目文件解析器：YAML 元数据 + Markdown 二级/三级标题结构
package main

import (
	"os"
	"path/filepath"
	"strings"
)

// Question 一道题目
type Question struct {
	Path    string            // 文件完整路径
	File    string            // 文件名（含扩展名）
	Meta    map[string]string // YAML 元数据（系统属性 + 可选属性 + 自定义属性）
	Prompt  string            // ## 题目（题干）
	Answer  string            // ### 答案
	Explain string            // ### 解析
	Note    string            // ### 备注
	Extra   map[string]string // 未识别的其它节（保留）
}

// parseQuestion 解析一个题目 Markdown 文件
func parseQuestion(path string) (*Question, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	q := &Question{
		Path:  path,
		File:  filepath.Base(path),
		Meta:  map[string]string{},
		Extra: map[string]string{},
	}

	lines := strings.Split(string(data), "\n")

	// --- YAML front matter ---
	start := 0
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				for _, l := range lines[1:i] {
					l = strings.TrimSpace(l)
					if l == "" || strings.HasPrefix(l, "#") {
						continue // 空行 / 注释
					}
					if idx := strings.Index(l, ":"); idx > 0 {
						key := strings.TrimSpace(l[:idx])
						val := strings.TrimSpace(strings.Trim(l[idx+1:], `"'`))
						q.Meta[key] = val
					}
				}
				start = i + 1
				break
			}
		}
	}

	// --- Markdown 结构：## 题目、### 答案/解析/备注 ---
	// 跳过 fenced code block（```...```）内的内容
	sections := map[string]*[]string{
		"题目": {},
		"答案": {},
		"解析": {},
		"备注": {},
	}
	other := map[string]*[]string{}
	cur := ""
	inFence := false
	for _, l := range lines[start:] {
		if strings.HasPrefix(strings.TrimSpace(l), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		switch {
		case strings.HasPrefix(l, "### "):
			if s, ok := sections[strings.TrimSpace(l[4:])]; ok {
				cur = strings.TrimSpace(l[4:])
				_ = s
			} else {
				title := strings.TrimSpace(l[4:])
				if _, exists := other[title]; !exists {
					other[title] = &[]string{}
				}
				cur = title
			}
		case strings.HasPrefix(l, "## "):
			title := strings.TrimSpace(l[3:])
			if _, ok := sections[title]; !ok {
				if _, exists := other[title]; !exists {
					other[title] = &[]string{}
				}
			}
			cur = title
		default:
			if cur == "" {
				continue // 结构之前的散落文字忽略
			}
			if s, ok := sections[cur]; ok {
				*s = append(*s, l)
			} else if s, ok := other[cur]; ok {
				*s = append(*s, l)
			}
		}
	}

	join := func(ss []string) string { return strings.TrimSpace(strings.Join(ss, "\n")) }
	q.Prompt = join(*sections["题目"])
	q.Answer = join(*sections["答案"])
	q.Explain = join(*sections["解析"])
	q.Note = join(*sections["备注"])
	for k, v := range other {
		q.Extra[k] = join(*v)
	}
	return q, nil
}

// isQuestion 判断是否为有效题目文件。
// 判定：有 YAML 元数据，或含有真实的题目/答案/解析/备注正文（排除 README 等文档 md）
func (q *Question) isQuestion() bool {
	if len(q.Meta) > 0 {
		return true
	}
	return q.Prompt != "" || q.Answer != "" || q.Explain != "" || q.Note != ""
}

// ID 返回题目编号（优先元数据 id，否则文件名去扩展名）
func (q *Question) ID() string {
	if id, ok := q.Meta["id"]; ok && id != "" {
		return id
	}
	return strings.TrimSuffix(q.File, ".md")
}