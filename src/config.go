// 配置系统（V1.3.0）：全局配置（用户/.requiz）+ 项目配置（题库/.requiz）
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FieldDef 元数据字段定义（含可选值列表）
type FieldDef struct {
	Name   string   `json:"name"`
	Label  string   `json:"label"`
	Values []string `json:"values"`
}

// GlobalConfig 全局配置（用户级：字段定义 + 链接题库 + 默认配置）
type GlobalConfig struct {
	Defaults   map[string]string
	MetaFields []FieldDef
	Links      []string
}

// ProjectConfig 项目（题库）配置（bank/app + 该题库自定义字段）
type ProjectConfig struct {
	App         string
	Bank        string
	Version     string
	QuestionDir string
	MetaFields  []FieldDef
}

// ---------- 路径 ----------

func globalConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".requiz"
	}
	return filepath.Join(home, ".requiz")
}

func globalConfigPath() string {
	return filepath.Join(globalConfigDir(), "config.yaml")
}

func projectConfigPath(dir string) string {
	return filepath.Join(dir, ".requiz", "config.yaml")
}

// ---------- 全局配置 ----------

// defaultGlobalConfig 返回全局配置默认模板（系统属性 + 可选属性字段定义）
func defaultGlobalConfig() GlobalConfig {
	return GlobalConfig{
		Defaults: map[string]string{
			"port": "8080",
		},
		MetaFields: []FieldDef{
			{Name: "chapter", Label: "章节", Values: []string{}},
			{Name: "grade", Label: "年级", Values: []string{"高一", "高二", "高三"}},
			{Name: "difficulty", Label: "难度", Values: []string{"★", "★★", "★★★"}},
			{Name: "importance", Label: "重要性", Values: []string{"了解", "必做经典题", "高频考点"}},
			{Name: "source", Label: "来源", Values: []string{"课本习题", "模考真题", "高考真题", "机构练习"}},
			{Name: "knowledge", Label: "知识点", Values: []string{}},
			{Name: "type", Label: "题型", Values: []string{"选择题", "填空题", "解答题"}},
		},
		Links: []string{},
	}
}

// readGlobalConfig 读取全局配置；不存在时创建默认模板
func readGlobalConfig() (GlobalConfig, error) {
	path := globalConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		// 不存在：创建默认模板
		g := defaultGlobalConfig()
		if err := writeGlobalConfig(g); err != nil {
			return g, err
		}
		return g, nil
	}
	return parseGlobalConfig(string(data)), nil
}

func parseGlobalConfig(data string) GlobalConfig {
	g := GlobalConfig{Defaults: map[string]string{}}
	section := ""
	inField := false
	inFieldValues := false
	curIdx := -1
	for _, line := range strings.Split(data, "\n") {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		trim := strings.TrimSpace(line)
		switch {
		case indent == 0:
			if idx := strings.Index(trim, ":"); idx > 0 {
				section = strings.TrimSpace(trim[:idx])
				inField = false
				inFieldValues = false
			}
		case indent == 2 && section == "links" && strings.HasPrefix(trim, "-"):
			g.Links = append(g.Links, strings.TrimSpace(trim[1:]))
		case indent == 2 && section == "defaults":
			if idx := strings.Index(trim, ":"); idx > 0 {
				g.Defaults[strings.TrimSpace(trim[:idx])] = strings.TrimSpace(trim[idx+1:])
			}
		case indent == 2 && section == "meta_fields":
			inFieldValues = false
			if strings.HasPrefix(trim, "- ") {
				rest := strings.TrimSpace(trim[2:])
				fd := FieldDef{}
				if idx := strings.Index(rest, ":"); idx > 0 {
					fd.Name = strings.TrimSpace(rest[idx+1:])
				}
				g.MetaFields = append(g.MetaFields, fd)
				curIdx = len(g.MetaFields) - 1
				inField = true
			}
		case indent == 4 && section == "meta_fields" && inField && curIdx >= 0:
			if idx := strings.Index(trim, ":"); idx > 0 {
				k := strings.TrimSpace(trim[:idx])
				v := strings.TrimSpace(trim[idx+1:])
				switch k {
				case "label":
					g.MetaFields[curIdx].Label = v
				case "values":
					inFieldValues = true
				}
			}
		case indent >= 4 && inFieldValues && curIdx >= 0 && strings.HasPrefix(trim, "-"):
			g.MetaFields[curIdx].Values = append(g.MetaFields[curIdx].Values, strings.TrimSpace(trim[1:]))
		}
	}
	return g
}

func (g GlobalConfig) serialize() string {
	var b strings.Builder
	b.WriteString("# requiz 全局配置（用户级）\n")
	b.WriteString("# 字段定义：系统属性 + 可选属性元数据字段\n\n")
	b.WriteString("defaults:\n")
	keys := []string{}
	for k := range g.Defaults {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&b, "  %s: %s\n", k, g.Defaults[k])
	}
	if len(g.MetaFields) > 0 {
		b.WriteString("\nmeta_fields:\n")
		for _, f := range g.MetaFields {
			fmt.Fprintf(&b, "  - name: %s\n", f.Name)
			if f.Label != "" {
				fmt.Fprintf(&b, "    label: %s\n", f.Label)
			}
			if len(f.Values) > 0 {
				b.WriteString("    values:\n")
				for _, v := range f.Values {
					fmt.Fprintf(&b, "      - %s\n", v)
				}
			}
		}
	}
	if len(g.Links) > 0 {
		b.WriteString("\nlinks:\n")
		for _, l := range g.Links {
			fmt.Fprintf(&b, "  - %s\n", l)
		}
	}
	return b.String()
}

func writeGlobalConfig(g GlobalConfig) error {
	dir := globalConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(globalConfigPath(), []byte(g.serialize()), 0644)
}

// ---------- 项目配置 ----------

// parseProjectConfig 解析题库 .requiz/config.yaml（bank/app/version/questions_dir/meta_fields）
func parseProjectConfig(data string) ProjectConfig {
	p := ProjectConfig{}
	section := ""
	inField := false
	inFieldValues := false
	curIdx := -1
	for _, line := range strings.Split(data, "\n") {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		trim := strings.TrimSpace(line)
		switch {
		case indent == 0:
			if idx := strings.Index(trim, ":"); idx > 0 {
				k := strings.TrimSpace(trim[:idx])
				v := strings.TrimSpace(trim[idx+1:])
				section = k
				inField = false
				inFieldValues = false
				switch k {
				case "app":
					p.App = v
				case "bank":
					p.Bank = v
				case "version":
					p.Version = v
				case "questions_dir":
					p.QuestionDir = v
				}
			}
		case indent == 2 && section == "meta_fields":
			inFieldValues = false
			if strings.HasPrefix(trim, "- ") {
				rest := strings.TrimSpace(trim[2:])
				fd := FieldDef{}
				if idx := strings.Index(rest, ":"); idx > 0 {
					fd.Name = strings.TrimSpace(rest[idx+1:])
				}
				p.MetaFields = append(p.MetaFields, fd)
				curIdx = len(p.MetaFields) - 1
				inField = true
			}
		case indent == 4 && section == "meta_fields" && inField && curIdx >= 0:
			if idx := strings.Index(trim, ":"); idx > 0 {
				k := strings.TrimSpace(trim[:idx])
				v := strings.TrimSpace(trim[idx+1:])
				switch k {
				case "label":
					p.MetaFields[curIdx].Label = v
				case "values":
					inFieldValues = true
				}
			}
		case indent >= 4 && inFieldValues && curIdx >= 0 && strings.HasPrefix(trim, "-"):
			p.MetaFields[curIdx].Values = append(p.MetaFields[curIdx].Values, strings.TrimSpace(trim[1:]))
		}
	}
	return p
}

func (p ProjectConfig) serialize() string {
	var b strings.Builder
	fmt.Fprintf(&b, "app: %s\n", p.App)
	fmt.Fprintf(&b, "bank: %s\n", p.Bank)
	if p.Version != "" {
		fmt.Fprintf(&b, "version: %s\n", p.Version)
	}
	if p.QuestionDir != "" {
		fmt.Fprintf(&b, "questions_dir: %s\n", p.QuestionDir)
	}
	if len(p.MetaFields) > 0 {
		b.WriteString("meta_fields:\n")
		for _, f := range p.MetaFields {
			fmt.Fprintf(&b, "  - name: %s\n", f.Name)
			if f.Label != "" {
				fmt.Fprintf(&b, "    label: %s\n", f.Label)
			}
			if len(f.Values) > 0 {
				b.WriteString("    values:\n")
				for _, v := range f.Values {
					fmt.Fprintf(&b, "      - %s\n", v)
				}
			}
		}
	}
	return b.String()
}

func writeProjectConfig(dir string, p ProjectConfig) error {
	cfgDir := filepath.Join(dir, ".requiz")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(projectConfigPath(dir), []byte(p.serialize()), 0644)
}

// readProjectConfig 读取题库项目配置（文件不存在时返回空）
func readProjectConfig(dir string) (ProjectConfig, error) {
	path := projectConfigPath(dir)
	data, err := os.ReadFile(path)
	if err != nil {
		return ProjectConfig{}, nil
	}
	return parseProjectConfig(string(data)), nil
}

// ---------- 字段定义合并 ----------

// mergeFieldDefs 合并全局字段与项目字段（同名取并集 values，项目 label 优先）
func mergeFieldDefs(global, project []FieldDef) []FieldDef {
	merged := []FieldDef{}
	byName := map[string]int{}
	for _, f := range global {
		merged = append(merged, f)
		byName[f.Name] = len(merged) - 1
	}
	for _, f := range project {
		if i, ok := byName[f.Name]; ok {
			// 合并 values 去重
			seen := map[string]bool{}
			for _, v := range merged[i].Values {
				seen[v] = true
			}
			for _, v := range f.Values {
				if !seen[v] {
					merged[i].Values = append(merged[i].Values, v)
					seen[v] = true
				}
			}
			if f.Label != "" {
				merged[i].Label = f.Label
			}
		} else {
			merged = append(merged, f)
			byName[f.Name] = len(merged) - 1
		}
	}
	return merged
}
