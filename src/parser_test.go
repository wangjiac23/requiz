// parser 单元测试
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sample = `---
app: requiz
bank: math
id: M001
chapter: 第一章
difficulty: ★★
自定义字段: 任意值
---

## 题目

题干第一行

题干第二行

### 答案

答案内容

### 解析

解析内容

### 备注

备注内容
`

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "题.md")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestParseFrontMatter(t *testing.T) {
	q, err := parseQuestion(writeTemp(t, sample))
	if err != nil {
		t.Fatal(err)
	}
	if q.Meta["app"] != "requiz" {
		t.Errorf("app = %q, want requiz", q.Meta["app"])
	}
	if q.Meta["bank"] != "math" {
		t.Errorf("bank = %q, want math", q.Meta["bank"])
	}
	if q.Meta["id"] != "M001" {
		t.Errorf("id = %q, want M001", q.Meta["id"])
	}
	if q.Meta["自定义字段"] != "任意值" {
		t.Errorf("自定义字段 = %q", q.Meta["自定义字段"])
	}
}

func TestParseSections(t *testing.T) {
	q, err := parseQuestion(writeTemp(t, sample))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(q.Prompt, "题干第一行") || !strings.Contains(q.Prompt, "题干第二行") {
		t.Errorf("Prompt = %q", q.Prompt)
	}
	if !strings.HasPrefix(q.Answer, "答案内容") {
		t.Errorf("Answer = %q", q.Answer)
	}
	if q.Explain != "解析内容" {
		t.Errorf("Explain = %q", q.Explain)
	}
	if q.Note != "备注内容" {
		t.Errorf("Note = %q", q.Note)
	}
}

func TestIDFallback(t *testing.T) {
	q, err := parseQuestion(writeTemp(t, "## 题目\n内容\n"))
	if err != nil {
		t.Fatal(err)
	}
	if q.Meta["id"] != "" {
		t.Errorf("meta id 应为空")
	}
	if q.ID() != "题" {
		t.Errorf("ID fallback = %q, want 题（文件名）", q.ID())
	}
}