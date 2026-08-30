// CLI 命令实现：connect / list / read / view
package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// currentBank 按传入参数获取连接好的题库
func currentBank(args []string) (*Bank, error) {
	dir := "."
	if len(args) > 1 {
		return nil, fmt.Errorf("connect 只接受一个题库目录参数")
	}
	if len(args) == 1 {
		dir = args[0]
	}
	return connectBank(dir)
}

// cmdConnect 连接题库文件夹
func cmdConnect(args []string) error {
	bank, err := currentBank(args)
	if err != nil {
		return err
	}
	fmt.Printf("connected to bank: %s\n", bank.Dir)
	if bank.Name != "" {
		fmt.Printf("  bank name : %s\n", bank.Name)
	}
	fmt.Printf("  app       : %s\n", bank.App)
	fmt.Printf("  questions : %d\n", len(bank.Questions))
	for _, q := range bank.Questions {
		fmt.Printf("    - %s  %s\n", q.ID(), q.File)
	}
	return nil
}

// cmdList 列出题库题目（目录可选题库，缺省当前目录）
func cmdList(args []string) error {
	dir := "."
	if len(args) > 1 {
		return fmt.Errorf("list 最多接受一个题库目录参数")
	}
	if len(args) == 1 {
		dir = args[0]
	}
	bank, err := connectBank(dir)
	if err != nil {
		return err
	}
	if len(bank.Questions) == 0 {
		fmt.Println("题库为空（未扫描到 .md 题目文件）")
		return nil
	}
	fmt.Printf("共 %d 道题：\n", len(bank.Questions))
	for _, q := range bank.Questions {
		rel, _ := filepath.Rel(bank.Dir, q.Path)
		fmt.Printf("%-8s %s\n", q.ID(), rel)
	}
	return nil
}

// cmdRead 读取单题全文（用法：read <id> [题库目录]）
func cmdRead(args []string) error {
	if len(args) < 1 || len(args) > 2 {
		return fmt.Errorf("read 需要一个题目 id 或文件名，可选题库目录")
	}
	dir := "."
	if len(args) == 2 {
		dir = args[1]
	}
	bank, err := connectBank(dir)
	if err != nil {
		return err
	}
	q, err := bank.find(args[0])
	if err != nil {
		return err
	}
	fmt.Printf("文件: %s\n", q.Path)
	// 元数据
	fmt.Println("=== 元数据 ===")
	if len(q.Meta) == 0 {
		fmt.Println("（无）")
	}
	for k, v := range q.Meta {
		fmt.Printf("%s: %s\n", k, v)
	}
	// 各节
	printSection("题目", q.Prompt)
	printSection("答案", q.Answer)
	printSection("解析", q.Explain)
	printSection("备注", q.Note)
	for k, v := range q.Extra {
		printSection(k, v)
	}
	return nil
}

// cmdView 按需显示题目部分。
// 手动解析参数，支持选项出现在任意位置（如：view M001 --a --e 与 view --a M001 均可）
func cmdView(args []string) error {
	showA, showE, showN, showY := false, false, false, false
	key, dir := "", ""
	for _, a := range args {
		switch a {
		case "--a", "-a":
			showA = true
		case "--e", "-e":
			showE = true
		case "--n", "-n":
			showN = true
		case "--yaml", "-yaml", "-y":
			showY = true
		default:
			if strings.HasPrefix(a, "-") {
				return fmt.Errorf("未知选项: %s", a)
			}
			if key == "" {
				key = a
			} else if dir == "" {
				dir = a
			} else {
				return fmt.Errorf("view 参数过多（用法：view <id> [题库目录] [选项]）")
			}
		}
	}
	if key == "" {
		return fmt.Errorf("view 需要一个题目 id 或文件名")
	}
	if dir == "" {
		dir = "."
	}
	bank, err := connectBank(dir)
	if err != nil {
		return err
	}
	q, err := bank.find(key)
	if err != nil {
		return err
	}

	fmt.Printf("# %s（%s）\n", q.ID(), strings.TrimSuffix(q.File, ".md"))
	if showY {
		fmt.Println("```yaml")
		for k, v := range q.Meta {
			fmt.Printf("%s: %s\n", k, v)
		}
		fmt.Println("```")
	}
	if showA {
		printSection("答案", q.Answer)
	}
	if showE {
		printSection("解析", q.Explain)
	}
	if showN {
		printSection("备注", q.Note)
	}
	for k, v := range q.Extra {
		printSection(k, v)
	}
	printSection("题目", q.Prompt) // 题干缺省显示，放最后
	return nil
}

func printSection(name, content string) {
	if content == "" {
		return
	}
	fmt.Printf("\n## %s\n%s\n", name, content)
}