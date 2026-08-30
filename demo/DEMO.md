# requiz v0.0.0 演示

demo 演示区结构：从 `data/` 选取题目 → 组织成题库（如 `demo/题库A/`），题库目录内含 `.requiz/` 连接配置。data 只存原始题库，不加配置。

```
demo/
└── 题库A/            ← 可连接的演示题库
    ├── .requiz/      连接配置
    └── 题目.md
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
```

预期输出示例（view M001 demo/题库A）：

```
# M001（交并运算）

## 题目
设集合 A = {x | x² - 1 = 0}，B = {x | x ∈ ℤ}，则 A ∩ B = ?

A. A = {-1, 1} 与 B 的交集是 {-1, 1}
```

详细用法见 README.md 与 docs/TECH.md。