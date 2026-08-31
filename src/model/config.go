// model：纯数据模型
package model

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
	Favorites  []string // 收藏（"题库目录|题目ID"）
	Pi         PiConfig // V3.1.0：pi 配置
}

// PiConfig pi 集成配置（V3.1.0）
type PiConfig struct {
	Path  string // pi 可执行路径（默认 C:/nvm/v25.1.0/pi.cmd）
	Model string // pi 模型（可选）
}

// QuestionList 自定义题目清单（组卷）
type QuestionList struct {
	Name string
	IDs  []string
}

// ProjectConfig 项目（题库）配置（bank/app + 该题库自定义字段）
type ProjectConfig struct {
	App         string
	Bank        string
	Version     string
	QuestionDir string
	MetaFields  []FieldDef
	Favorites   []string // 收藏题目 id（V1.5.0）
	Lists       []QuestionList
}
