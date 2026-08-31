// model：纯数据模型
package model

import "strings"

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

// ID 返回题目编号（优先元数据 id，否则文件名去扩展名）
func (q *Question) ID() string {
	if id, ok := q.Meta["id"]; ok && id != "" {
		return id
	}
	return strings.TrimSuffix(q.File, ".md")
}

// IsQuestion 判断是否为有效题目文件。
// 判定：有 YAML 元数据，或含有真实的题目/答案/解析/备注正文（排除 README 等文档 md）
func (q *Question) IsQuestion() bool {
	if len(q.Meta) > 0 {
		return true
	}
	return q.Prompt != "" || q.Answer != "" || q.Explain != "" || q.Note != ""
}
