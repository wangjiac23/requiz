// 题库（bank）连接与扫描
package quiz

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"requiz/src/model"
	"requiz/src/parser"
)

// connectBank 连接题库文件夹：
// 验证目录存在、读取 .requiz/config.yaml、扫描目录下所有题目 md 文件
// 统一使用绝对路径（保证题库对等比较与配置一致性）
func ConnectBank(dir string) (*model.Bank, error) {
	return connectBankFiltered(dir, false)
}

// ConnectBankAppOnly 只收录 app: requiz 的题目（V2.1.0 Obsidian Vault 模式）
func ConnectBankAppOnly(dir string) (*model.Bank, error) {
	return connectBankFiltered(dir, true)
}

func connectBankFiltered(dir string, appOnly bool) (*model.Bank, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("题库目录解析失败: %s（%w）", dir, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("题库目录不存在: %s（%w）", abs, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("路径不是文件夹: %s", abs)
	}

	bank := &model.Bank{Dir: abs, App: "requiz"}

	// 读取 .requiz/config.yaml（容错：不存在也能连，只是没有配置信息）
	cfgPath := filepath.Join(abs, ".requiz", "config.yaml")
	if data, err := os.ReadFile(cfgPath); err == nil {
		parsingLinks := false
		for _, l := range strings.Split(string(data), "\n") {
			l = strings.TrimSpace(l)
			if l == "" {
				continue
			}
			if strings.HasPrefix(l, "#") {
				continue
			}
			// links 列表项：以 "-" 开头的行，属于上一个 "links:" 节
			if parsingLinks && strings.HasPrefix(l, "-") {
				p := strings.TrimSpace(strings.TrimPrefix(l, "-"))
				if p != "" {
					if !filepath.IsAbs(p) {
						p = filepath.Join(abs, p)
					}
					bank.Links = append(bank.Links, p)
				}
				continue
			}
			if idx := strings.Index(l, ":"); idx > 0 {
				k := strings.TrimSpace(l[:idx])
				v := strings.TrimSpace(l[idx+1:])
				if k == "links" {
					parsingLinks = true
					continue
				}
				parsingLinks = false
				if k == "bank" {
					bank.Name = v
				} else if k == "app" {
					bank.App = v
				}
			}
		}
	}

	if bank.Name == "" {
		// 无 config 时回退用目录名
		bank.Name = filepath.Base(abs)
	}

	qs, err := scanQuestions(abs, appOnly)
	if err != nil {
		return nil, err
	}
	bank.Questions = qs
	return bank, nil
}

// scanQuestions 递归扫描目录下所有 .md 题目文件（跳过 .requiz / .git）
// appOnly=true 时仅收录 app: requiz 的 md（Obsidian Vault 模式）
func scanQuestions(root string, appOnly bool) ([]*model.Question, error) {
	var qs []*model.Question
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
			if q, err := parser.ParseQuestion(path); err == nil {
				if appOnly {
					if q.Meta["app"] == "requiz" {
						qs = append(qs, q)
					}
				} else if q.IsQuestion() {
					qs = append(qs, q)
				}
			}
		}
		return nil
	})
	return qs, err
}

// CreateEmptyBank 创建空题库（V2.1.0：serve 无目录时，供插件注入）
func CreateEmptyBank(dir string) *model.Bank {
	abs, _ := filepath.Abs(dir)
	return &model.Bank{Dir: abs, Name: filepath.Base(abs), App: "requiz", Questions: []*model.Question{}}
}
