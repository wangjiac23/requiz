// requiz for Obsidian (V2.0.0)
// 在 Obsidian 新标签页中 iframe 加载 localhost requiz 网页，复用全部功能
const { Plugin, PluginSettingTab, Setting, ItemView, Notice } = require("obsidian");

const VIEW_TYPE_REQUIZ = "requiz-view";

const DEFAULT_SETTINGS = {
  port: 8099,
  exePath: "",
  bankDir: "",
  vaultBank: "", // V2.1.0：空=不连；./=整个 Vault；./题目;./作业题=指定文件夹
};

class RequizView extends ItemView {
  constructor(leaf, settings) {
    super(leaf);
    this.settings = settings;
  }

  getViewType() {
    return VIEW_TYPE_REQUIZ;
  }

  getDisplayText() {
    return "Requiz 题库";
  }

  getIcon() {
    return "book-open";
  }

  async onOpen() {
    const container = this.containerEl.children[1];
    container.empty();
    // 修复：显式 flex 布局，保证 iframe 撑满剩余高度
    container.style.cssText =
      "display:flex;flex-direction:column;height:100%;overflow:hidden;padding:0";

    this.statusEl = container.createDiv({ cls: "requiz-status" });
    this.statusEl.style.cssText =
      "padding:8px 14px;font-size:13px;flex-shrink:0;background:var(--background-secondary);border-bottom:1px solid var(--background-modifier-border)";

    // iframe 容器（flex:1 撑满）
    const wrap = container.createDiv();
    wrap.style.cssText = "flex:1;min-height:0;position:relative;overflow:hidden";
    this.iframe = wrap.createEl("iframe");
    this.iframe.style.cssText = "width:100%;height:100%;border:none;display:block";
    this.loadFrame();

    // 重新加载按钮
    const reload = container.createEl("button", { text: "⟳ 重新加载" });
    reload.style.cssText =
      "position:absolute;top:46px;right:10px;z-index:10;opacity:0.6;font-size:12px";
    reload.onclick = () => this.loadFrame();
  }

  loadFrame() {
    this.iframe.src = "http://127.0.0.1:" + this.settings.port + "/";
    this.checkStatus();
  }

  async checkStatus() {
    this.statusEl.setText("⏳ 正在检测 requiz 服务…");
    try {
      const resp = await fetch(
        "http://127.0.0.1:" + this.settings.port + "/api/banks",
        { cache: "no-store" }
      );
      if (resp.ok) {
        const data = await resp.json();
        this.statusEl.setText(
          "✅ requiz 已连接（端口 " + this.settings.port + " · 题库 " + data.length + " 个）"
        );
      } else {
        this.showDisconnected();
      }
    } catch (e) {
      this.showDisconnected();
    }
  }

  showDisconnected() {
    this.statusEl.setText(
      "❌ requiz 服务未连接（端口 " +
        this.settings.port +
        "）——请在设置中配置 requiz.exe 与题库目录后点「启动 requiz」，或手动运行 requiz serve"
    );
  }

  async onClose() {
    // 仅关闭视图，不停止 requiz 服务
  }
}

class RequizPlugin extends Plugin {
  async onload() {
    await this.loadSettings();
    this.registerView(VIEW_TYPE_REQUIZ, (leaf) => new RequizView(leaf, this.settings));

    this.addRibbonIcon("book-open", "打开 Requiz 题库", () => this.activateView());
    this.addCommand({
      id: "open-requiz",
      name: "打开 Requiz 题库",
      callback: () => this.activateView(),
    });
    this.addSettingTab(new RequizSettingTab(this.app, this));

    // V2.1.0：接收 requiz 页面「打开本地」请求 → Obsidian 新标签页打开
    window.addEventListener("message", (e) => {
      if (e.data && e.data.type === "requiz-open" && e.data.path) {
        this.openInObsidian(e.data.path);
      }
    });

    // V2.1.0：Obsidian 文件变更 → 重新扫描注入（防抖）
    this.registerEvent(
      this.app.vault.on("modify", () => this.scheduleRescan())
    );
    this.registerEvent(
      this.app.vault.on("create", () => this.scheduleRescan())
    );
    this.registerEvent(
      this.app.vault.on("delete", () => this.scheduleRescan())
    );
  }

  // V2.1.0：打开 Obsidian 文件（绝对路径 → Vault 相对路径，统一正斜杠，多级匹配）
  async openInObsidian(absPath) {
    const base = this.app.vault.adapter.getBasePath();
    const baseN = (base || "").replace(/\\/g, "/").replace(/\/+$/, "");
    const absN = (absPath || "").replace(/\\/g, "/");
    let rel = absN;
    if (baseN && absN.startsWith(baseN)) {
      rel = absN.slice(baseN.length + 1);
    }
    // ① 直接按相对路径定位
    let file = this.app.vault.getAbstractFileByPath(rel);
    // ② 兑底：全文件匹配（相对路径/文件名）
    if (!file && absN) {
      const name = absN.split("/").pop();
      file =
        this.app.vault.getFiles().find((f) => f.path === rel) ||
        this.app.vault.getFiles().find((f) => f.path.endsWith("/" + rel)) ||
        this.app.vault.getFiles().find((f) => f.name === name && absN.endsWith("/" + f.path));
    }
    if (file) {
      await this.app.workspace.getLeaf("tab").openFile(file);
    } else {
      new Notice("❌ 找不到文件：" + rel + "\nbase:" + baseN + "\nabs:" + absN);
    }
  }

  // V2.1.0：Vault 扫描注入（复用 Obsidian API，防抖 1s）
  scheduleRescan() {
    clearTimeout(this._scanTimer);
    this._scanTimer = setTimeout(() => this.scanAndInject(), 1000);
  }

  async scanAndInject() {
    if (!this.settings.vaultBank) return; // 未启用 Vault 模式
    const { vault, metadataCache } = this.app;
    const base = vault.adapter.getBasePath();
    // 解析允许的目录列表（./=全部；./a;./b=指定）
    const allowDirs = this.resolveVaultDirs();
    const files = vault.getMarkdownFiles();
    const qs = [];
    let scanned = 0, matched = 0;
    for (const f of files) {
      // 目录过滤
      if (allowDirs) {
        const rel = f.path;
        const ok = allowDirs.some((d) => rel === d || rel.startsWith(d + "/"));
        if (!ok) continue;
      }
      scanned++;
      // 先读正文（metadataCache 预筛在启动初期可能未就绪，直接读最可靠）
      let content = "";
      try {
        content = await vault.adapter.read(f.path);
      } catch (e) {
        continue;
      }
      // 仅检查 frontmatter 元数据块（--- 开头到 --- 结尾），避免正文误匹配
      const fmMatch = content.match(/^---\n([\s\S]*?)\n---/);
      if (!fmMatch) continue;
      if (!/app\s*:\s*requiz/.test(fmMatch[1])) continue;
      matched++;
      qs.push({ path: base + "\\" + f.path.split("/").join("\\"), content: content });
    }
    // 注入 requiz
    try {
      const resp = await fetch("http://127.0.0.1:" + this.settings.port + "/api/external/banks", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ bank: { name: "Obsidian Vault", dir: base, questions: qs } }),
      });
      const d = await resp.json();
      new Notice("📚 requiz 已同步 Vault：扫描 " + scanned + " 文件，收录 " + (d.count || 0) + " 题（目录：" + base + "）");
      this.refreshViews();
    } catch (e) {
      new Notice("⚠️ requiz 同步失败：" + e.message + "（服务是否已启动？）");
    }
  }

  // 解析 Vault 目录列表：空=null（全部）；./a;./b → [a, b]
  resolveVaultDirs() {
    const v = this.settings.vaultBank;
    if (!v) return null;
    if (v === "./") return null; // 全部
    return v
      .split(";")
      .map((p) => p.trim())
      .filter(Boolean)
      .map((p) => (p.startsWith("./") ? p.slice(2) : p.replace(/^[\\\\/]+/, "")));
  }

  onunload() {}

  async activateView() {
    const { workspace } = this.app;
    let leaf = workspace.getLeavesOfType(VIEW_TYPE_REQUIZ)[0];
    if (!leaf) {
      leaf = workspace.getLeaf("tab");
      await leaf.setViewState({ type: VIEW_TYPE_REQUIZ, active: true });
    }
    workspace.revealLeaf(leaf);
  }

  async loadSettings() {
    this.settings = Object.assign({}, DEFAULT_SETTINGS, await this.loadData());
  }

  async saveSettings() {
    await this.saveData(this.settings);
  }

  // 启动 requiz 服务：先检测是否已运行，未运行再启动；结果 Notice 弹窗提示
  startRequiz() {
    // ① 先检测端口是否已有 requiz 服务
    fetch("http://127.0.0.1:" + this.settings.port + "/api/banks", { cache: "no-store" })
      .then((r) => {
        if (r.ok) {
          new Notice("✅ requiz 已在运行（端口 " + this.settings.port + "），直接打开 Requiz 标签页即可");
          this.refreshViews();
          return;
        }
        this.launchRequiz();
      })
      .catch(() => this.launchRequiz());
  }

  launchRequiz() {
    if (!this.settings.exePath) {
      new Notice("❌ 请先在设置中配置 requiz.exe 路径");
      return;
    }
    // V2.1.0：Vault 模式不指定目录（空库启动，插件扫描注入）；普通模式用 bankDir
    const dirArg = this.settings.vaultBank ? "" : this.settings.bankDir;
    const parts = [
      '"' + this.settings.exePath + '"',
      "serve",
      dirArg ? '"' + dirArg + '"' : null,
      "-port",
      String(this.settings.port),
    ].filter(Boolean);
    const cmd = parts.join(" ");
    try {
      const { exec } = require("child_process");
      exec(cmd, (err, stdout, stderr) => {
        if (err && !err.killed) {
          if (stderr && stderr.indexOf("bind") >= 0) {
            new Notice(
              "ℹ️ 端口 " + this.settings.port + " 已被占用（requiz 可能已在运行），直接打开 Requiz 标签页即可"
            );
            this.afterServiceReady();
          } else {
            new Notice("⚠️ requiz 启动失败：" + (stderr || err.message));
          }
          return;
        }
      });
      new Notice("⏳ 正在启动 requiz（端口 " + this.settings.port + "）…");
      this.waitForService(this.settings.port, (ok) => {
        if (ok) {
          new Notice("✅ requiz 启动成功，已连接（端口 " + this.settings.port + "）");
          this.afterServiceReady();
        } else {
          new Notice(
            "❌ 启动后未检测到服务。请检查 exe 路径是否正确，或手动运行：requiz serve [题库] -port " +
              this.settings.port
          );
        }
      });
    } catch (e) {
      new Notice(
        "⚠️ 插件环境无法直接启动服务，请手动运行：requiz serve -port " + this.settings.port
      );
    }
  }

  // 服务就绪后：Vault 模式触发扫描注入，普通模式刷新视图
  afterServiceReady() {
    if (this.settings.vaultBank) this.scanAndInject();
    this.refreshViews();
  }

  waitForService(port, cb) {
    const deadline = Date.now() + 10000;
    const tick = () => {
      fetch("http://127.0.0.1:" + port + "/api/banks", { cache: "no-store" })
        .then((r) => (r.ok ? cb(true) : retry()))
        .catch(() => retry());
    };
    const retry = () => {
      if (Date.now() > deadline) {
        cb(false);
        return;
      }
      setTimeout(tick, 800);
    };
    tick();
  }

  refreshViews() {
    this.app.workspace.getLeavesOfType(VIEW_TYPE_REQUIZ).forEach((leaf) => {
      if (leaf.view && leaf.view.checkStatus) leaf.view.checkStatus();
    });
  }
}

class RequizSettingTab extends PluginSettingTab {
  constructor(app, plugin) {
    super(app, plugin);
    this.plugin = plugin;
  }

  display() {
    const { containerEl } = this;
    containerEl.empty();

    containerEl.createEl("h2", { text: "Requiz for Obsidian" });

    new Setting(containerEl)
      .setName("端口")
      .setDesc("requiz 服务端口（默认 8099）")
      .addText((text) =>
        text
          .setPlaceholder("8099")
          .setValue(String(this.plugin.settings.port))
          .onChange(async (value) => {
            this.plugin.settings.port = parseInt(value, 10) || 8099;
            await this.plugin.saveSettings();
          })
      );

    new Setting(containerEl)
      .setName("requiz.exe 路径")
      .setDesc("requiz 可执行文件完整路径（如 D:\\requiz\\dist\\requiz.exe）")
      .addText((text) =>
        text
          .setPlaceholder("D:\\requiz\\dist\\requiz.exe")
          .setValue(this.plugin.settings.exePath)
          .onChange(async (value) => {
            this.plugin.settings.exePath = value.trim();
            await this.plugin.saveSettings();
          })
      );

    new Setting(containerEl)
      .setName("题库目录")
      .setDesc("启动时打开的题库目录（可选，留空用默认）")
      .addText((text) =>
        text
          .setPlaceholder("D:\\requiz\\demo\\题库A")
          .setValue(this.plugin.settings.bankDir)
          .onChange(async (value) => {
            this.plugin.settings.bankDir = value.trim();
            await this.plugin.saveSettings();
          })
      );

    new Setting(containerEl)
      .setName("Vault 题库路径（分布式）")
      .setDesc("空 = 不连接；./ = 整个库；./题目;./作业题 = 只读取指定文件夹（相对/绝对、分号分隔）")
      .addText((text) =>
        text
          .setPlaceholder("./题目;./作业题")
          .setValue(this.plugin.settings.vaultBank)
          .onChange(async (value) => {
            this.plugin.settings.vaultBank = value.trim();
            await this.plugin.saveSettings();
          })
      );

    new Setting(containerEl)
      .setName("启动 requiz 服务")
      .setDesc("使用上面配置的 exe 路径与题库目录启动服务（结果以弹窗提示）")
      .addButton((btn) =>
        btn.setButtonText("▶ 启动 requiz").setCta().onClick(() => {
          this.plugin.startRequiz();
        })
      );
  }
}

module.exports = RequizPlugin;
