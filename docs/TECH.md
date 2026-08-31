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

- **v0.0.0** ✅：CLI 连接题库 + 读取/显示题目
- **v1.0.0** ✅：CLI + Localhost Web（列表/详情 + JSON API）
- **v1.1.0** ✅：多题库链接 + 三栏 Obsidian 风格 UI + 标签筛选
- **v1.1.1** ✅（修订）：题库A 扩至 10 题 · 设置弹窗修复 · 去除页面版本号
- **v1.1.2** ✅（修订）：侧边栏/上边栏可隐藏（按钮切换 + 拖拽调宽）
- **v1.1.3** ✅（修订）：拖拽到阈值自动隐藏（侧边栏 <100px / 筛选栏 <40px）· 筛选栏拖拽调高
- **v1.1.4** ✅（修订）：📌 固定按钮，固定时拖拽只调尺寸不隐藏
- **V1.2.0** ✅：公式渲染（KaTeX 本地化）+ 打开本地文件 + 网页编辑 + 刷新同步
- **v1.2.1** ✅（修订）：修复公式不渲染（auto-render 忽略 pre 标签 → div.content）
- **v1.2.2** ✅（修订）：网页端修改题目名（重命名文件）+ 元数据（留空删字段，path 自动维护）
- **v1.2.3** ✅（修订）：元数据下拉选值 + 新增字段 + 值可加入配置/仅本次
- **v1.2.4** ✅（修订）：元数据 UI 优化（按钮移位/下拉整合/删除/折叠/滚动）
- **V1.3.0** ✅：配置文件拆分（全局 `用户/.requiz` + 项目 `题库/.requiz`）+ 设置升级为配置管理入口 + 题库对等化
- **v1.4.0** ✅：题目盒子/展开/三模式浏览（列表/双栏/卡片）/自定义显示/收藏
- **v1.4.1** ✅（修订）：题干完整显示（tree 返回 prompt）；修复答案/解析/备注展开 bug（冒泡 + 标题截断）
- **v1.4.2** ✅（修订）：展开分级（一级=题目/二级=答案解析备注）；卡片模式自动一级展开
- **v1.4.3** ✅（修订）：双栏/卡片模式键盘导航（↑← 上一题、↓→ 下一题，输入框聚焦忽略）
- **v1.5.0** ✅：左侧导航三分区（题库/收藏/清单）+ 自定义清单组卷 + 筛选选题导出（JSON/HTML）
- **v1.5.1** ✅（修订）：修复收藏星标刷新不显示（favorites 带 bank 参数 + dir 字段）
- **v1.5.2** ✅（修订）：导出文件操作（📂 打开默认程序 / 📍 资源管理器定位，output 内校验）
- **v1.5.3** ✅（修订）：导出 HTML 内嵌 KaTeX JS，公式自包含渲染（file:// 安全限制修复）
- **v1.6.0** ✅：serve 代码模块化——JS 功能 Go 源文件组织（server.go 拆分为 8 模块，前端资源 assets.go）
- **v1.7.0** ✅：文件夹导航（任意层级树 + 目录筛选/Ctrl 多选/空白取消）+ 题组 + 在线测试模式（计分计时/自动评/报告）
- **v1.7.1** ✅（修订）：非测试模式点击题组名 → 主区域只查看该题组题目（提示条/高亮/空白取消）
- **v1.7.2** ✅（修订）：删除评分（用户自评）；做题区统一单一作答区（无选项/题型控件）
- **v1.8.0** ✅：图片及附件（题库/image/题目名/ 目录约定 + 渲染 + 上传）
- **v2.0.0** ✅：requiz for Obsidian（插件：新标签页 iframe 复用 localhost 全部功能）
- **v2.1.1** ✅（修订）：打开本地智能判断（库内 Obsidian / 库外资源管理器）+ 定位失败资源管理器兜底
- **v2.2.0** ✅：知识库、题库联动（题目链接笔记 `## 链接笔记` + Obsidian 反链）
- **v2.0.0** 🚧：requiz for Obsidian（插件：新标签页 iframe 复用 localhost 全部功能）
- **v2.1.0** 📝：分布式题目
- **v2.2.0** 📝：知识库、题库联动

---

## V1.0.0（CLI + Localhost）

1. 技术栈：Go + Obsidian md 文件（同 v0.0.0，零第三方依赖）
2. 运行环境：CLI + Localhost
3. 新增：
   - `requiz serve [题库目录] [-port 端口]`：启动 Web 服务（默认 `127.0.0.1:8080`）
   - 首页 `/`：题库信息 + 题目列表（点击进入详情）
   - 详情 `/question?id=<id>`：YAML 元数据标签 + 题干/答案/解析/备注
   - JSON API：`/api/questions`（全部题目）、`/api/question?id=<id>`（单题），为组卷/Agent 预留

---

## V1.1.0（多题库链接 + 三栏 UI + 标签筛选）

1. 技术栈：同 v1.0.0，仅 Go 标准库，零第三方依赖
2. 新增功能：
   - **多题库链接**：主题库通过 `.requiz/config.yaml` 的 `links` 节链接其它题库（相对/绝对路径）；无 config 的题库自动回退用目录名。
   - **三栏界面**（仿 Obsidian）：
     - 顶部栏：题库下拉切换 + ⚙ 设置按钮
     - 左侧边栏：题目包树（按顶层目录分组，可折叠）
     - 上边栏：元数据标签筛选（章节/年级/难度/重要性/来源/知识点/题型，自动聚合有值标签）
     - 主区域：题目卡片，点击加载题干/答案/解析/备注
   - **设置链接**：⚙ → 输入目录 → `POST /api/link`（持久化写入主题库 `links` 节，去重）
3. JSON API 扩展：
   - `GET /api/banks`：题库列表（name/dir/count/current）
   - `GET /api/tree?bank=<目录|名称>`：题目包树（含每题 meta）
   - `GET /api/questions?bank=&tag=&value=`：按元数据标签筛选
   - `GET /api/question?bank=&id=`：单题详情
4. 服务端重构：单 bank 闭包 → `Store`（主库 + 链接库缓存），按目录或名称切换；兼容 v1.0.0 全部端点。

### V1.1.1 修订
1. 题库A 题目增至 10 题（新增 M002~M010，分 4 个章节包：集合/不等式/函数/指数对数）
2. 设置弹窗修复：CSS 类型由 `template.JS` 改为 `template.CSS`（原 `ZgotmplZ` 导致样式全部未注入）；弹窗默认 `display:none`，点击 ⚙ 添加 `.show` 类弹出居中窗口
3. 移除页面版本号显示（标题 + 顶栏 brand）

### V1.1.2 修订
1. 左侧边栏 / 上边栏可隐藏：顶栏新增 `☷`/`⌄` 切换按钮（加 `hidden` 类）
2. 侧边栏右缘新增拖拽手柄（`#resizer`，mousedown 跟随调整宽度 120~480px），双击快速隐藏/显示

### V1.1.3 修订
1. 侧边栏拖拽：去掉双击快速隐藏，改为拖拽宽度 < 100px 自动收起（按钮 ☷ 恢复）
2. 筛选栏拖拽：新增底部手柄 `#filtersResizer` 拖拽调整高度（40~160px），拖到 < 40px 自动隐藏（按钮 ⌄ 恢复）

### V1.1.4 修订
1. 侧边栏顶部（`#sidebarHead`）与筛选栏右侧（`#filtersBar`）新增 📌 固定按钮
2. 固定后（`.pinned` 高亮）拖拽 clamp 到最小尺寸（侧边栏 100px / 筛选栏 40px），不触发自动隐藏；取消固定恢复正常隐藏逻辑

---

## V1.2.0（公式渲染 + 本地打开 + 网页编辑 + 刷新同步）

1. 技术栈：同 v1.1.x（Go 标准库 + 前端原生 JS），新增 KaTeX 本地静态资源（`web/katex/`，离线可用，无 CDN 依赖）
2. 新增功能：
   - **数学公式渲染**：引入 KaTeX + auto-render，题目详情中 `$...$`（行内）与 `$$...$$`（块级）公式自动渲染
   - **打开本地文件**：题目卡片「打开」按钮 → `POST /api/open` → `explorer.exe /select,<绝对路径>` 在资源管理器中定位该题 md 文件（仅限题库目录内，路径校验防越界）
   - **网页编辑**：题目卡片「编辑」按钮 → 弹窗表单（题干/答案/解析/备注）→ `POST /api/question/save` 序列化写回本地 md（YAML 元数据保持原样）
   - **刷新同步**：顶栏「刷新」按钮 → `POST /api/reload` 重新扫描题库目录；冲突处理：网页编辑即写盘、刷新即重扫，后保存/后修改的一方覆盖先前的（时间靠后者胜）
3. API 新增：
   - `POST /api/open`：body `{bank, id}`，explorer 定位文件
   - `POST /api/question/save`：body `{bank, id, prompt, answer, explain, note}`，写回 md
   - `POST /api/reload`：body `{bank}`，重新扫描题库
   - 静态：`GET /katex/*`（KaTeX 资源）
4. 结构变动：`web/katex/` 目录新增；src/server.go 增挂载与新 API；src/parser.go 增序列化函数

### V1.2.0 修订
1. v1.2.1 修复公式渲染：auto-render 默认忽略 `<pre>`/`<code>` 标签 → 题目内容改用 `<div class="content">`（CSS 保持 pre-wrap）；列表卡片、详情、整页兼容页统一渲染；提取统一 `katexDelims` 并修正 `\(` 分隔符
2. v1.2.2 网页端修改题目名与元数据：编辑弹窗新增「题目名」输入（同目录重命名，非法名/重名拒绝）与元数据键值编辑（留空删除，app/bank/path 系统字段保护，path 自动维护）
3. v1.2.3 元数据下拉选值：新增 `GET /api/meta-values` / `POST /api/meta-value/add`，字段已知值存 `.requiz/meta-values.yaml`；编辑弹窗字段行改为下拉（空/已知值/自定义），「＋ 添加字段」新增自定义键，新值可选「📥 加入配置」（写 yaml 去重，全局可用）或「✍️ 仅本次」（仅当前题目）
4. v1.2.4 元数据 UI 优化：添加字段按钮位于元数据区下方、在编辑栏内完成（点按钮展开字段名输入行，回车/添加按钮确认，取消/Esc 收起，无浏览器弹窗）；下拉栏整合「➕ 新增值/✍️ 自定义」两选项；当前值不在已知列表时显示「值(自定义)」；每条字段 ✕ 删除（后端 meta 改为整体替换生效）；编辑弹窗 max-height 可滚动；元数据区 ▾ 折叠

---

## V1.3.0（配置文件拆分 + 设置升级为配置管理入口）

1. 技术栈：同 v1.2.x（Go 标准库），新增全局配置读写（用户 home 目录）
2. 配置文件架构：
   - **全局配置** `用户/.requiz/config.yaml`：meta_fields（系统属性 + 可选属性元数据字段定义）、links（程序链接的所有题库地址）、defaults（默认配置：端口等）；首次启动自动创建模板
   - **项目配置** `题库A/.requiz/config.yaml`：bank/app 等系统标识 + 该题库自定义属性元数据字段；旧版 links 首次启动自动迁移至全局配置
3. 设置入口升级：⚙ 弹窗改为配置管理面板，含「全局配置」「题库配置」页签，显示两个配置文件真实路径；可编辑字段定义、管理链接题库（增删）、查看默认配置
4. 新 API：`GET/POST /api/config/global`、`GET/POST /api/config/project`（读写配置）；字段下拉数据源改为全局字段定义
5. 兼容：已有 .requiz/config.yaml 的 links 自动迁移；既有功能无回归
6. 题库对等化：connectBank 统一绝对路径；serve 打开的题库自动注册到全局；全局 links = 全部题库列表（下拉平等切换）；设置面板「题库管理」列出全部题库（含当前，不可移除）；POST /api/config/global/save 后端保护当前题库

---

## V1.4.0（题目盒子排版 + 双栏/卡片浏览 + 自定义显示 + 收藏）

1. 技术栈：同 v1.3.x（Go 标准库 + 前端原生 JS）
2. 新增功能：
   - **题目盒子**：每道题一个卡片盒子，默认只显示题干；「展开」显示完整内容，答案/解析/备注分段折叠（默认隐藏，点击显示）
   - **双栏浏览**：左侧题目列表 + 右侧详情面板（点列表题右侧实时显示），与单列模式可切换
   - **自定义显示**：显示清单（勾选显示字段：题型/难度/章节/来源等）+ 筛选器组合筛题（复用元数据标签）
   - **卡片浏览**：单题卡片模式 + ◀ 上一题 / 下一题 ▶ 前进后退
   - **收藏**：⭐ 收藏/取消；收藏列表单独查看；持久化
3. 收藏持久化：全局配置新增 `favorites`（用户级，跨题库），新 API：`POST /api/favorite`（收藏/取消）、`GET /api/favorites`（列表）、收藏筛选
4. 结构变动：主区域支持「列表/双栏/卡片」三种模式（顶部切换）；题目卡片组件化；显示清单设置存全局配置 display 节或前端偏好
5. 兼容：既有功能无回归

### V1.4.0 实现补充
1. 收藏持久化：全局配置新增 `favorites`（`题库目录|题目ID`），`POST /api/favorite` toggle、`GET /api/favorites` 列表；前端 ☆/⭐ 按钮 + 「仅看收藏」过滤
2. 三模式浏览：主区域工具栏（列表/双栏/卡片）+ 收藏过滤；题目盒子组件（题干常显 + 展开 + 答案/解析/备注分段折叠，保留打开本地/编辑按钮）
3. 显示清单：`⚙ 字段` 面板勾选显示字段（存 localStorage），盒子/详情标签按清单渲染

### V1.4.0 修订
1. v1.4.1：① 题干完整显示（questSummary 新增 `prompt` 字段，盒子渲染完整题干不再截断）② 修复展开 bug（sec-btn 点击 `stopPropagation` 防冒泡收起详情；标题改用 `data-title` 防 slice 截断错乱）
2. v1.4.2：展开分级（一级=展开题目、二级=答案/解析/备注）；卡片模式渲染后自动一级展开
3. v1.4.3：键盘导航（双栏/卡片模式：↑← 上一题、↓→ 下一题，input/textarea/select 聚焦时忽略）

---

## V1.5.0（左侧导航三分区 + 自定义清单组卷 + 筛选选题导出）

1. 技术栈：同 v1.4.x（Go 标准库 + 前端原生 JS）；导出 JSON 纯 Go 生成，PDF/Word 走可打印 HTML
2. 新增功能：
   - **左侧导航三分区**：侧边栏页签「题库 / 收藏 / 自定义清单」
   - **收藏元数据落盘**：收藏时在项目配置记录题目元数据（id + meta），收藏导航直接读配置
   - **自定义清单（组卷）**：选择题库题目组成清单，持久化项目配置 `lists`，导航可查看/打开
   - **筛选选题导出**：筛选后「全选」或逐个勾选 → 导出所选
   - **导出部分可选**：题干/答案/解析/备注 任意勾选组合
   - **多格式导出**：JSON / HTML（可打印存 PDF、Word 兼容）/ 图片（后续）
3. 项目配置新增：`favorites`（收藏 id + 元数据）、`lists`（清单：名称 + 题目 id）
4. 新 API：`GET/POST /api/lists`（清单 CRUD）、`POST /api/export`（bank/ids/parts/format → output/ 文件）；`/api/favorite` 改写项目配置
5. 结构变动：侧边栏三页签 UI；选择模式（勾选状态 + 工具栏导出按钮）；导出渲染模板
6. 兼容：全局配置 favorites 迁移至项目配置；既有功能无回归

### V1.5.0 实现补充
1. 收藏落盘：`POST /api/favorite` 改写项目配置 favorites（id 列表），`GET /api/favorites` 返回动态元数据；侧边栏「收藏」页签
2. 清单组卷：`GET /api/lists`、`POST /api/lists/save`（创建/更新/删除）；项目配置 `lists` 节；侧边栏「清单」页签可展开查看
3. 选择模式：☑ 选择按钮 → 盒子勾选（全选/单选）→ 📋 存清单 / 📤 导出
4. 导出：`POST /api/export`（bank/ids/parts/format）→ output/ 下 JSON（纯 Go）或 HTML（可打印 PDF / Word 兼容，KaTeX 不渲染保留 LaTeX）
5. 侧边栏三分区：题库树 / 收藏 / 清单页签切换（`state.sideTab`）

### V1.5.0 修订
1. v1.5.1 修复收藏星标刷新不显示：`GET /api/favorites` 前端带 `bank` 参数（切换题库后正确）；后端补回 `dir` 字段；前端收藏 key 统一用 `state.bank + "|" + id`
2. v1.5.2 导出文件操作：`POST /api/export/open` 支持 `action=open`（默认程序打开）/ `select`（explorer /select 定位）；仅限 `output/` 目录内（Rel 校验，越权 403）
3. v1.5.3 导出 HTML 公式渲染：KaTeX JS（katex.min.js + auto-render）内嵌进导出文件（file:// 下浏览器禁止外部 JS 加载）；CSS 相对引用 `../web/katex/`；实测 9 公式渲染

---

## V1.6.0（serve 代码模块化：JS 功能 Go 源文件组织）

1. 目标：将 server.go（1500+ 行大杂烩）按职责拆分为多个 Go 源文件，功能零变化
2. 拆分方案：
   - `store.go`：Store 结构、newStore、addLink、migrateProjectLinks、containsStr
   - `api.go`：核心查询 API（banks/tree/questions/question/link）
   - `config_api.go`：配置 API（config/global/project、meta-values）
   - `favorites.go`：收藏 + 清单 API
   - `export.go`：导出模块（JSON/HTML、打开/定位）
   - `web.go`：cmdServe/parseServeArgs/indexHandler/questionHandler + indexTpl/questionTpl
   - `views.go`：视图模型与转换（questSummary/treeOf/toJSON/firstLine/writeJSON）
   - `assets.go`：前端资源常量（indexJS/indexCSS）
3. 同包 main 共享类型/函数，无 import 变化；模板/CSS/JS 常量整体迁移
4. 验收：go vet/test/build 通过；页面渲染与交互零回归
5. 5 包深化：`model`（数据模型）/`parser`（解析）/`config`（配置）/`quiz`（题库连接）/`web`（HTTP 服务）；main 收口 CLI；依赖单向（main→quiz→parser→model、config→model、web→全部）

---

## V1.7.0（文件夹导航 + 题组 + 在线测试模式）

1. 技术栈：同 v1.6.x（5 包结构 + 前端原生 JS）
2. 新增功能：
   - **任意层级文件夹导航**：导航树严格按文件系统目录结构嵌套（多级展开收起）；点击目录主区域只显示该目录题目，Ctrl 多选合并，点空白取消
   - **题组**：原「清单」升级为题组（文件夹样式，含题目元数据引用，题目本体在题库）；题组旁「测试」按钮
   - **在线测试模式**：主区域只显示题组题目 + 答题区；可设计分/不计分、正计时/倒计时/不计时
   - **评分**：客观题（选择/填空）自动比对判分，主观题（解答）显示参考答案由用户评分
   - **测试报告**：每次测试生成报告（对错/得分/用时/总得分），可查看与导出
3. 结构变动：/api/tree 返回嵌套目录树；项目配置 lists 语义为题组；前端 state.selDirs（目录多选）+ state.testing（测试态）；评分与报告前端汇总
4. 兼容：既有功能无回归

### V1.7.0 实现补充（批次 A：文件夹导航）
1. `/api/tree` 改为嵌套目录树：`treeNode{name,path,dirs,questions}` 递归，`questSummary` 新增 `rel` 字段
2. 前端递归渲染导航树（文件夹行：箭头展开/收起 + 名称点击筛选），`state.selDirs` 目录多选（Ctrl/⌘），`visibleQuestions` 按目录过滤（含子目录）
3. 点主区域空白清除目录选择；根目录题目直接显示

### V1.7.0 实现补充（批次 B：题组 + 测试模式入口）
4. 清单升级「题组」：导航显示 📁 文件夹样式（名称/题数/▶ 测试按钮），题目元数据引用不复制本体
5. 测试模式：`state.testing` 状态机；主区域只显示题目 + 答题区（按题型：选择题 radio/选项提取、填空 input、解答 textarea）；「✕ 退出」返回
6. 完成测试收集答案（评分/计时/报告批次 C）

### V1.7.0 实现补充（批次 C：计分计时 + 评分 + 报告）
7. 测试设置弹窗：计分/不计分（每题 10 分）、正计时/倒计时（分钟可设）/不计时
8. 计时器：倒计时到 0 自动交卷；正计时显示用时
9. 客观题自动评：填空题/选择题归一化比对（容错空格/量词/LaTeX 符号，互相包含判对）；`/api/lists` 返回 `answer` 字段供评分
10. 主观题（解答）用户评分（0-10 分输入，更新总分）
11. 测试报告：对错/得分/用时/总得分 + 参考答案 + 我的答案

### V1.7.0 修订
1. v1.7.1 非测试模式查看题组：点击题组名（`state.selList`）→ `visibleQuestions` 按题组 id 过滤，主区域只显示题组内题目（提示条 + 题组行高亮，点空白取消，与目录筛选可叠加）
2. v1.7.2 删除评分（autoGrade/gradeMatch/normalizeAns 移除）：测试仅计时 + 作答，报告显示「我的答案 vs 参考答案」由用户自评；做题区统一单一 textarea（无选项提取/题型控件）；测试设置弹窗去掉计分选项

---

## V1.8.0（图片及附件）

1. 目录约定：`题库目录/image/<题目名>/`（题目名 = md 文件名去扩展名），存放该题图片与附件
2. 静态服务：`GET /image?bank=<题库>&file=<相对路径>` → 返回 `题库/image/` 内文件（Rel 校验防越界；Content-Type 按扩展名；非图片作为附件下载）
3. 上传：`POST /api/image/upload`（multipart：bank/id/file）→ 写入 `image/<题目名>/`（同名覆盖需确认）
4. 前端渲染：题目内容解析 Markdown 图片语法 `![alt](image/xxx)` → `<img src="/image?...">`（KaTeX 渲染后处理）；图片加载失败占位；附件渲染为下载链接
5. 兼容：无图片题目不受影响；既有功能无回归

---

## V2.0.0（requiz for Obsidian）

1. 目标：Obsidian 新标签页内实现 localhost requiz 全部功能（复用网页，服务端零改动）
2. 方案：Obsidian 插件 `obsidian-plugin/`：注册视图（`requiz-view`）→ 新标签页 iframe 加载 `http://127.0.0.1:<port>/`
3. 插件能力：
   - 功能区图标/命令打开 Requiz 标签页
   - 设置页：端口（默认 8099）、requiz.exe 路径、题库目录
   - 「启动 requiz」按钮（child_process 拉起 exe serve 题库目录 -port）；连接状态检测（fetch /api/banks）
4. 结构：`obsidian-plugin/manifest.json` + `main.js`（Plugin/ItemView/SettingTab）
5. CORS：requiz `withCORS` 中间件加 `Access-Control-Allow-Origin: *` + OPTIONS 预检，Obsidian 插件跨源 fetch /api/banks 状态检测可用

### V2.0.0 修订
1. v2.1.1 打开本地智能化：requiz 前端 iframe 环境 postMessage（`requiz-open` 带路径）；插件 `openInObsidian` 先判断路径是否在当前库内（`getBasePath`）——库内 `getAbstractFileByPath` 多级匹配 + Obsidian 新标签页打开；库外/定位失败用 `explorer.exe /select` 资源管理器兜底（不再报「找不到文件」）

---

## V2.2.0（知识库、题库联动：链接笔记）

1. 题目格式新增 `## 链接笔记` 章节（与题目/答案/解析/备注同级），内容为 Obsidian 双链 `[[笔记名]]` 或 `[[路径|别名]]`
2. parser：`Question` 新增 `Links` 字段，`## 链接笔记` 解析到 Links（serialize 保留输出）；questionJSON/toJSON 透传
3. 前端：详情（展开/双栏）渲染「链接笔记」区；编辑弹窗新增链接笔记 textarea（保存写回 md）
4. Obsidian 反链：题目 md 中 `[[笔记名]]` 由 Obsidian 建立双向链接，右边栏「反向链接」可见
5. demo：M001 链接 `[[集合的基本概念]]`，配套笔记 `demo/题库A/笔记/集合的基本概念.md`（普通笔记不带 requiz 元数据，不会被当题目）
5. 兼容：requiz 服务零改动；插件仅做 iframe 集成