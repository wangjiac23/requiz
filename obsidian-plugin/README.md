# Requiz for Obsidian（v2.0.0）

在 Obsidian 新标签页中使用 requiz 题库系统（导航/组卷/测试/导出/图片全部功能）。

## 安装

1. 把本目录（`obsidian-plugin/`）复制为 Obsidian 库下的 `.obsidian/plugins/requiz-for-obsidian/`
2. Obsidian → 设置 → 第三方插件 → 启用「Requiz for Obsidian」
3. 功能区出现 📖 图标（或命令面板搜「打开 Requiz 题库」）

## 使用

1. **设置**（插件设置页）：
   - 端口：requiz 服务端口（默认 8099）
   - requiz.exe 路径：如 `D:\2.1 日常记忆\2026\2026.08\2026.08-6 题库系统\requiz\dist\requiz.exe`
   - 题库目录：可选（启动时打开的题库）
2. **启动服务**：设置页点「▶ 启动 requiz」；或手动运行 `requiz serve [题库] -port 8099`
3. **打开标签页**：功能区图标 / 命令 → Requiz 新标签页 iframe 加载 localhost 网页

> 说明：标签页顶部状态条显示连接状态；「⟳ 重新加载」可重载页面。
> 关闭标签页不会停止 requiz 服务（可手动 taskkill requiz.exe）。
