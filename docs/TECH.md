# requiz 技术文档（v0.0.0）

## 1. 技术栈与运行环境

- 编程语言：Go
- 数据库：Obsidian Markdown 文件
- Agent：pi
- 运行环境：CLI

## 2. 核心功能

能够读取、解析题目文件，并且可以输出不同部分。

功能范围：
1. 连接题库文件夹（bank）
2. 打开并读取题目文件
3. 根据需要显示内容

验收标准：连接 bank 成功。

## 3. 软件结构

```
requiz/
├── src/                 Go 源码
├── data/                题库仓库（原始题库，仅存题目）
│   └── math bank/       数学题库（题目文件所在地）
├── test/                测试
├── dist/                构建产物
├── docs/                技术文档
├── demo/                演示区：从 data 选取题目组织成题库
│   └── 题库A/           演示题库（题目 + .requiz 配置）
│       └── .requiz/     题库A 的连接配置
├── output/              输出目录
├── .gitignore
├── README.md
├── LICENSE
└── .github/             CI 配置
```

配置文件位于**题库目录**下的 `.requiz/`（连接配置，如 `demo/题库A/.requiz/config.yaml`）。

每个可连接的题库自带一份 `.requiz/`；`demo/` 演示区可从 `data/` 选取题目组织成多个题库（题库A/题库B…），`data/` 仅存放原始题库不带配置。

## 4. 题目文件格式

一个题目 = 一个 Markdown 笔记：

```
题目文件
├── YAML 元数据
│   ├── 系统属性：app（运行软件）、bank（所属题库）、id、path
│   └── 可选属性：章节、年级、难度、重要性、来源、知识点、题型
├── ## 题目（二级标题，题干）
├── ### 答案（三级标题，可选）
├── ### 解析（三级标题，可选）
└── ### 备注（三级标题，可选）
```

系统属性预设 YAML 键（英文，程序解析用）：
`app` `bank` `id` `path`

可选属性 YAML 键：
`chapter`（章节）`grade`（年级）`difficulty`（难度）`importance`（重要性）
`source`（来源）`knowledge`（知识点）`type`（题型）

自定义属性：用户可自由扩展任意 YAML 键。

## 5. CLI 命令（v0.0.0）

| 命令 | 说明 |
|------|------|
| `requiz connect <bank>` | 连接题库文件夹，验证 / 生成 `.requiz/` 配置 |
| `requiz list [bank]` | 列出题库所有题目（id + 标题） |
| `requiz read <id> [bank]` | 读取单题全文（可附加元数据） |
| `requiz view <id> [bank] [--a --e --n --yaml]` | 按需显示：缺省题干，`--a` 答案 `--e` 解析 `--n` 备注 |
| `requiz serve [bank] [-port]` | 启动 localhost Web 服务（v1.0.0 新增） |

## 6. 难度与重要性约定

- 难度（5 星）：★ 简单题 ★★ 基础题 ★★★ 中档题 ★★★★ 较难题 ★★★★★ 很难题
- 重要性：必做经典题 / 变形提升题 / 原创扩展题
- 来源：课本习题 / 高考真题 / 模拟题 / 练习题

## 7. 路线图

- **v1.0.0** ✅：CLI + Localhost
- **v2.0.0**：引入 Agent

---

## V1.0.0（CLI + Localhost）

1. 技术栈：Go + Obsidian md 文件（同 v0.0.0，零第三方依赖）
2. 运行环境：CLI + Localhost
3. 新增：
   - `requiz serve [题库目录] [-port 端口]`：启动 Web 服务（默认 `127.0.0.1:8080`）
   - 首页 `/`：题库信息 + 题目列表（点击进入详情）
   - 详情 `/question?id=<id>`：YAML 元数据标签 + 题干/答案/解析/备注
   - JSON API：`/api/questions`（全部题目）、`/api/question?id=<id>`（单题），为组卷/Agent 预留