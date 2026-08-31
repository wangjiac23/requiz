// model：纯数据模型
package model

import (
	"fmt"
	"strings"
)

// Bank 题库
type Bank struct {
	Dir       string // 题库目录
	Name      string // 题库名称（.requiz/config.yaml 的 bank 字段）
	App       string // 运行软件
	Links     []string
	Questions []*Question
}

// Find 按 id 或文件名查找题目；重复时提示指定题库目录
func (b *Bank) Find(key string) (*Question, error) {
	var found []*Question
	for _, q := range b.Questions {
		if q.ID() == key || q.File == key || strings.TrimSuffix(q.File, ".md") == key {
			found = append(found, q)
		}
	}
	if len(found) == 0 {
		return nil, fmt.Errorf("找不到题目: %s", key)
	}
	if len(found) > 1 {
		return nil, fmt.Errorf("在多个题库目录找到题目 %s，请指定题库目录（如 requiz view %s demo）", key, key)
	}
	return found[0], nil
}
