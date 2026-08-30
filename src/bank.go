// 题库（bank）连接与扫描
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Bank 题库
type Bank struct {
	Dir       string // 题库目录
	Name      string // 题库名称（.requiz/config.yaml 的 bank 字段）
	App       string // 运行软件
	Questions []*Question
}

// connectBank 连接题库文件夹：
// 验证目录存在、读取 .requiz/config.yaml、扫描目录下所有题目 md 文件
func connectBank(dir string) (*Bank, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("题库目录不存在: %s（%w）", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("路径不是文件夹: %s", dir)
	}

	bank := &Bank{Dir: dir, App: "requiz"}

	// 读取 .requiz/config.yaml（容错：不存在也能连，只是没有配置信息）
	cfgPath := filepath.Join(dir, ".requiz", "config.yaml")
	if data, err := os.ReadFile(cfgPath); err == nil {
		for _, l := range strings.Split(string(data), "\n") {
			l = strings.TrimSpace(l)
			if l == "" || strings.HasPrefix(l, "#") {
				continue
			}
			if idx := strings.Index(l, ":"); idx > 0 {
				k := strings.TrimSpace(l[:idx])
				v := strings.TrimSpace(l[idx+1:])
				switch k {
				case "bank":
					bank.Name = v
				case "app":
					bank.App = v
				}
			}
		}
	}

	qs, err := scanQuestions(dir)
	if err != nil {
		return nil, err
	}
	bank.Questions = qs
	return bank, nil
}

// scanQuestions 递归扫描目录下所有 .md 题目文件（跳过 .requiz / .git）
func scanQuestions(root string) ([]*Question, error) {
	var qs []*Question
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".requiz" || d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ".md") {
			if q, err := parseQuestion(path); err == nil && q.isQuestion() {
				qs = append(qs, q)
			}
		}
		return nil
	})
	return qs, err
}

// find 按 id 或文件名查找题目；重复时提示指定题库目录
func (b *Bank) find(key string) (*Question, error) {
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