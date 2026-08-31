// 题目文件解析器：YAML 元数据 + Markdown 二级/三级标题结构
package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"requiz/src/model"
)

// parseQuestion 解析一个题目 Markdown 文件
func ParseQuestion(path string) (*model.Question, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	q := &model.Question{
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

// serializeQuestion 将题目序列化为 YAML md 文件内容（供编辑保存写盘）
func SerializeQuestion(q *model.Question) string {
	var b strings.Builder
	b.WriteString("---\n")
	// 固定顺序输出常用元数据，其余按字母序
	order := []string{"app", "bank", "id", "path", "chapter", "grade", "difficulty", "importance", "source", "knowledge", "type"}
	written := map[string]bool{}
	for _, k := range order {
		if v, ok := q.Meta[k]; ok {
			fmt.Fprintf(&b, "%s: %s\n", k, v)
			written[k] = true
		}
	}
	keys := []string{}
	for k := range q.Meta {
		if !written[k] {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&b, "%s: %s\n", k, q.Meta[k])
	}
	b.WriteString("---\n\n")
	writeSection := func(title, content string) {
		if strings.TrimSpace(content) == "" {
			return
		}
		fmt.Fprintf(&b, "## %s\n\n%s\n\n", title, strings.TrimSpace(content))
	}
	writeSection("题目", q.Prompt)
	writeSection("答案", q.Answer)
	writeSection("解析", q.Explain)
	writeSection("备注", q.Note)
	// 其它自定义节（Extra）保持原样
	keys2 := []string{}
	for k := range q.Extra {
		keys2 = append(keys2, k)
	}
	sort.Strings(keys2)
	for _, k := range keys2 {
		writeSection(k, q.Extra[k])
	}
	return b.String()
}