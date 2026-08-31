# requiz v1.1.1 演示

demo 演示区结构：从 `data/` 选取题目 → 组织成题库（如 `demo/题库A/`、`demo/题库B/`），题库目录内含 `.requiz/` 连接配置。data 只存原始题库，不加配置。

```
demo/
├── 题库A/            ← 主题库（可连接）
│   ├── .requiz/      连接配置（links 链接题库B）
│   └── 题目.md
└── 题库B/            ← 链接题库（函数专题，下拉切换演示）
    ├── .requiz/
    └── 函数专题/题目.md
```

```bash
cd requiz

# 1. 连接演示题库（demo/题库A 自带 .requiz 配置）
./dist/requiz.exe connect "demo/题库A"
# => connected to bank: demo/题库A

# 2. 列出题库中的题目
./dist/requiz.exe list demo/题库A
# => M001  交并运算.md

# 3. 读取单题全文
./dist/requiz.exe read M001 demo/题库A

# 4. 按需显示：只看看答案和解析
./dist/requiz.exe view M001 demo/题库A --a --e

# 5. 只看题干
./dist/requiz.exe view M001 demo/题库A

# 6. 启动 localhost 网页服务（v1.0.0 → v1.1.0）
./dist/requiz.exe serve demo/题库A -port 8080
# => requiz web   : http://127.0.0.1:8080/
# => 主题库 : demo/题库A（1 题）
# => 链接题库 : demo\题库B（1 题）

# 7. 网页操作（v1.1.0）
#   - 顶部栏：下拉框在 题库A / 题库B 间切换（仿 Obsidian 库选择）
#   - ⚙ 设置 → 输入题库目录 → 链接：写入 demo/题库A/.requiz/config.yaml 的 links 节
#   - 左侧边栏：题目包树（题库B 显示「函数专题」包，点击折叠/展开）
#   - 上边栏：自动聚合标签（难度/知识点/题型…），选中即筛选
#   - 主区域：点题目卡片加载题干/答案/解析/备注

# JSON API（v1.1.0）
#   http://127.0.0.1:8080/api/banks                    题库列表
#   http://127.0.0.1:8080/api/tree?bank=题库B          题目包树
#   http://127.0.0.1:8080/api/questions?tag=知识点&value=xxx   筛选
#   http://127.0.0.1:8080/api/question?id=M001         单题详情
```

预期输出示例（view M001 demo/题库A）：

```
# M001（交并运算）

## 题目
设集合 A = {x | x² - 1 = 0}，B = {x | x ∈ ℤ}，则 A ∩ B = ?

A. A = {-1, 1} 与 B 的交集是 {-1, 1}
```

详细用法见 README.md 与 docs/TECH.md。
