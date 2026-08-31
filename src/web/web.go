// 由 server.go 拆分生成（V1.6.0 模块化，功能零变化）
package web

import (
	"requiz/src/quiz"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func Serve(args []string) error {
	dir, port, err := parseServeArgs(args)
	if err != nil {
		return err
	}
	bank, err := quiz.ConnectBank(dir)
	if err != nil {
		return err
	}
	store := newStore(bank)

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler(store))
	mux.HandleFunc("/question", questionHandler(store))
	mux.HandleFunc("/api/banks", apiBanksHandler(store))
	mux.HandleFunc("/api/tree", apiTreeHandler(store))
	mux.HandleFunc("/api/questions", apiQuestionsHandler(store))
	mux.HandleFunc("/api/question", apiQuestionHandler(store))
	mux.HandleFunc("/api/link", apiLinkHandler(store))
	mux.HandleFunc("/api/open", apiOpenHandler(store))
	mux.HandleFunc("/api/question/save", apiQuestionSaveHandler(store))
	mux.HandleFunc("/api/reload", apiReloadHandler(store))
	mux.HandleFunc("/api/meta-values", apiMetaValuesHandler(store))
	mux.HandleFunc("/api/meta-value/add", apiMetaValueAddHandler(store))
	mux.HandleFunc("/api/config/global", apiConfigGlobalHandler(store))
	mux.HandleFunc("/api/config/global/save", apiConfigGlobalSaveHandler(store))
	mux.HandleFunc("/api/config/project", apiConfigProjectHandler(store))
	mux.HandleFunc("/api/favorite", apiFavoriteHandler(store))
	mux.HandleFunc("/api/favorites", apiFavoritesHandler(store))
	mux.HandleFunc("/api/lists", apiListsHandler(store))
	mux.HandleFunc("/api/lists/save", apiListsSaveHandler(store))
	mux.HandleFunc("/api/export", apiExportHandler(store))
	mux.HandleFunc("/api/export/open", apiExportOpenHandler(store))
	mux.Handle("/katex/", http.StripPrefix("/katex/", http.FileServer(http.Dir(katexDir()))))

	addr := "127.0.0.1:" + port
	fmt.Printf("requiz web   : http://%s/\n", addr)
	fmt.Printf("当前题库     : %s（%d 题）\n", bank.Dir, len(bank.Questions))
	for _, b := range store.banks()[1:] {
		fmt.Printf("其它题库     : %s（%d 题）\n", b.Dir, len(b.Questions))
	}
	fmt.Println("按 Ctrl+C 停止服务")
	return http.ListenAndServe(addr, mux)
}


func parseServeArgs(args []string) (dir, port string, err error) {
	port = "8080"
	dir = "."
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-port", "--port":
			if i+1 >= len(args) {
				err = fmt.Errorf("-port 需要一个端口号")
				return
			}
			port = args[i+1]
			i++
		default:
			if strings.HasPrefix(args[i], "-") {
				err = fmt.Errorf("未知选项: %s", args[i])
				return
			}
			if dir != "." {
				err = fmt.Errorf("serve 参数过多（用法：requiz serve [题库目录] [-port 端口]）")
				return
			}
			dir = args[i]
		}
	}
	return
}

// katexDir 定位 KaTeX 静态资源目录（优先当前工作目录，其次可执行文件上级）

func katexDir() string {
	candidates := []string{"web/katex"}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "..", "web", "katex"))
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return "web/katex"
}

// ---------- V1.2.0 API：打开本地 / 编辑保存 / 刷新 ----------

// apiOpenHandler POST /api/open {bank, id}：explorer 定位本地题目文件

func indexHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		indexTpl.Execute(w, indexData{CSS: indexCSS, AppJS: indexJS})
	}
}

// questionHandler 整页详情（兼容分享链接）

func questionHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "缺少参数 id（?id=M001）", http.StatusBadRequest)
			return
		}
		b, err := s.bankByDir(r.URL.Query().Get("bank"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		q, err := b.Find(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		data := questDetailView{
			ID:      q.ID(),
			File:    q.File,
			Prompt:  q.Prompt,
			Meta:    q.Meta,
			Answer:  q.Answer,
			Explain: q.Explain,
			Note:    q.Note,
			Extra:   q.Extra,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		questionTpl.Execute(w, data)
	}
}


type questDetailView struct {
	ID      string
	File    string
	Prompt  string
	Meta    map[string]string
	Answer  string
	Explain string
	Note    string
	Extra   map[string]string
}


type indexData struct {
	CSS   template.CSS
	AppJS template.JS
}


var indexTpl = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html lang="zh">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>requiz</title>
<link rel="stylesheet" href="/katex/katex.min.css">
<style>{{.CSS}}</style>
</head>
<body>
<header id="topbar">
  <div class="brand">📚 requiz</div>
  <div class="bankbar">
    <button id="toggleSidebar" title="隐藏侧边栏">☷</button>
    <button id="toggleFilters" title="隐藏筛选栏">⌄</button>
    <select id="bankSel" title="选择题库"></select>
    <button id="reloadBtn" title="刷新题库（同步本地/网页改动）">⟳</button>
    <button id="settingsBtn" title="设置">⚙ 设置</button>
  </div>
</header>
<div id="filtersWrap">
  <div id="filtersBar">
    <div id="filters"></div>
    <button id="pinFilters" title="固定筛选栏（固定时拖拽只调高不隐藏）">📌</button>
  </div>
  <div id="filtersResizer" title="拖拽调整高度（拖到最矮隐藏）"></div>
</div>
<div id="body">
  <div id="sidebarWrap">
    <div id="sidebarCol">
      <div id="sidebarHead">
        <div id="sideTabs">
          <button class="stab active" data-tab="tree" title="题库题目导航">题库</button>
          <button class="stab" data-tab="fav" title="收藏题目导航">收藏</button>
          <button class="stab" data-tab="lists" title="自定义题目清单">清单</button>
        </div>
        <button id="pinSidebar" title="固定侧边栏（固定时拖拽只调宽不隐藏）">📌</button>
      </div>
      <aside id="sidebar"></aside>
    </div>
    <div id="resizer" title="拖拽调整宽度（拖到最窄隐藏）"></div>
  </div>
  <main id="main">
    <div id="mainToolbar">
      <button class="mode active" data-mode="list" title="列表浏览">📋 列表</button>
      <button class="mode" data-mode="split" title="双栏浏览">📑 双栏</button>
      <button class="mode" data-mode="card" title="卡片浏览">🃏 卡片</button>
      <span style="flex:1"></span>
      <button id="selectBtn" title="选择模式（勾选题目）">☑ 选择</button>
      <button id="saveListBtn" title="将选中题目存为清单" style="display:none">📋 存清单</button>
      <button id="exportBtn" title="导出选中题目" style="display:none">📤 导出</button>
      <button id="displayBtn" title="自定义显示字段">⚙ 字段</button>
      <button id="favFilterBtn" title="只看收藏的题目">☆ 收藏</button>
    </div>
    <div id="mainContent"><div class="empty">加载中…</div></div>
  </main>
</div>
<div id="modal" hidden>
  <div class="modal-box" style="width:600px;max-height:85vh;overflow-y:auto">
    <h3>⚙ 设置 <small style="color:var(--muted);font-weight:400">配置管理</small></h3>
    <div class="cfg-tabs">
      <button id="tabGlobal" class="cfg-tab active">🌐 全局配置</button>
      <button id="tabProject" class="cfg-tab">📁 题库配置</button>
    </div>
    <!-- 全局配置页签 -->
    <div id="cfgGlobal">
      <p class="tip">配置文件：<code id="cfgGlobalPath"></code></p>
      <h4>默认配置</h4>
      <div id="cfgDefaults"></div>
      <h4>题库列表（全部题库，当前打开的不可移除）<button id="linkOk" style="margin-left:8px;padding:2px 8px">＋ 添加</button></h4>
      <div id="cfgLinks"></div>
      <input id="linkInput" type="text" placeholder="输入题库目录路径后点链接" style="width:100%;margin:4px 0 8px">
      <h4>元数据字段定义</h4>
      <div id="cfgFields"></div>
    </div>
    <!-- 题库配置页签 -->
    <div id="cfgProject" hidden>
      <p class="tip">配置文件：<code id="cfgProjectPath"></code></p>
      <h4>题库信息</h4>
      <div id="cfgProjectInfo"></div>
      <h4>自定义属性字段</h4>
      <div id="cfgProjectFields"></div>
    </div>
    <div id="linkMsg" class="tip"></div>
    <div class="modal-actions">
      <button id="linkCancel">关闭</button>
    </div>
  </div>
</div>
<div id="displayModal" hidden>
  <div class="modal-box" style="width:360px">
    <h3>显示字段清单 <small style="color:var(--muted);font-weight:400">勾选要在题目上显示的字段</small></h3>
    <div id="displayList"></div>
    <div id="displayMsg" class="tip"></div>
    <div class="modal-actions">
      <button id="displayOk">保存</button>
      <button id="displayCancel">取消</button>
    </div>
  </div>
</div>
<div id="testSetupModal" hidden>
  <div class="modal-box" style="width:400px">
    <h3>🧪 测试设置 <span id="testSetupName" class="meta"></span></h3>
    <label class="lbl">计分</label>
    <select id="tsScored" style="width:100%;padding:6px;margin:4px 0 8px;border:1px solid var(--border);border-radius:6px">
      <option value="1">✅ 计分（每题 10 分）</option>
      <option value="0">🚫 不计分</option>
    </select>
    <label class="lbl">计时</label>
    <select id="tsTimer" style="width:100%;padding:6px;margin:4px 0 8px;border:1px solid var(--border);border-radius:6px">
      <option value="none">⏱ 不计时</option>
      <option value="countdown">⏳ 倒计时（分钟）</option>
      <option value="countup">⏲ 正计时</option>
    </select>
    <div id="tsMinWrap" hidden>
      <label class="lbl">倒计时时长（分钟）</label>
      <input id="tsMinutes" type="number" min="1" max="180" value="10" style="width:100%;padding:6px;margin:4px 0 8px;border:1px solid var(--border);border-radius:6px">
    </div>
    <div class="modal-actions">
      <button id="tsOk">开始测试</button>
      <button id="tsCancel">取消</button>
    </div>
  </div>
</div>
<div id="exportModal" hidden>
  <div class="modal-box" style="width:380px">
    <h3>导出题目 <span id="exportCount" class="meta"></span></h3>
    <label class="lbl">导出部分</label>
    <div id="exportParts">
      <label style="display:block;margin:3px 0"><input type="checkbox" value="prompt" checked> 题干</label>
      <label style="display:block;margin:3px 0"><input type="checkbox" value="answer"> 答案</label>
      <label style="display:block;margin:3px 0"><input type="checkbox" value="explain"> 解析</label>
      <label style="display:block;margin:3px 0"><input type="checkbox" value="note"> 备注</label>
    </div>
    <label class="lbl">导出格式</label>
    <select id="exportFormat" style="width:100%;padding:6px;margin:4px 0 8px;border:1px solid var(--border);border-radius:6px">
      <option value="json">JSON（自定义清单）</option>
      <option value="html">HTML（浏览器打印存 PDF / Word 可打开）</option>
    </select>
    <div id="exportMsg" class="tip"></div>
    <div class="modal-actions">
      <button id="exportOk">导出</button>
      <button id="exportCancel">取消</button>
    </div>
  </div>
</div>
<div id="editModal" hidden>
  <div class="modal-box">
    <h3>编辑题目 <span id="editId" class="meta"></span></h3>
    <label class="lbl">题目名（文件名）</label>
    <input id="editFile" type="text">
    <label class="lbl" id="metaFoldBtn" style="cursor:pointer;user-select:none">▾ 元数据</label>
    <div id="editMeta"></div>
    <div style="text-align:right;margin-bottom:8px"><button id="editAddField">＋ 添加字段</button></div>
    <label class="lbl">题干</label>
    <textarea id="editPrompt" rows="5"></textarea>
    <label class="lbl">答案</label>
    <textarea id="editAnswer" rows="3"></textarea>
    <label class="lbl">解析</label>
    <textarea id="editExplain" rows="3"></textarea>
    <label class="lbl">备注</label>
    <textarea id="editNote" rows="2"></textarea>
    <div id="editMsg" class="tip"></div>
    <div class="modal-actions">
      <button id="editSave">保存</button>
      <button id="editCancel">取消</button>
    </div>
  </div>
</div>
<script src="/katex/katex.min.js"></script>
<script src="/katex/contrib/auto-render.min.js"></script>
<script>{{.AppJS}}</script>
</body>
</html>`))


var questionTpl = template.Must(template.New("question").Parse(`<!DOCTYPE html>
<html lang="zh">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.ID}} — requiz</title>
<link rel="stylesheet" href="/katex/katex.min.css">
<style>
 body{font-family:-apple-system,"Segoe UI","Microsoft YaHei",sans-serif;max-width:800px;margin:0 auto;padding:24px;background:#f7f8fa;color:#24292f}
 .card{background:#fff;border:1px solid #e1e4e8;border-radius:8px;padding:16px;margin:12px 0}
 .meta{color:#586069;font-size:0.85em}
 .content{white-space:pre-wrap;background:#f6f8fa;padding:12px;border-radius:6px;overflow-x:auto;font-size:0.9em}
 .tag{display:inline-block;background:#ddf4ff;color:#0969da;border-radius:12px;padding:2px 10px;margin:2px;font-size:0.85em}
 a{color:#0969da;text-decoration:none}
</style>
</head>
<body>
<p><a href="/">← 返回题库</a></p>
<h1>{{.ID}} <small class="meta">{{.File}}</small></h1>
<div class="card">{{range $k, $v := .Meta}}<span class="tag">{{$k}}: {{$v}}</span>{{end}}</div>
{{if .Prompt}}<div class="card"><h2>题目</h2><div class="content">{{.Prompt}}</div></div>{{end}}
{{if .Answer}}<div class="card"><h2>答案</h2><div class="content">{{.Answer}}</div></div>{{end}}
{{if .Explain}}<div class="card"><h2>解析</h2><div class="content">{{.Explain}}</div></div>{{end}}
{{if .Note}}<div class="card"><h2>备注</h2><div class="content">{{.Note}}</div></div>{{end}}
{{range $k, $v := .Extra}}<div class="card"><h2>{{$k}}</h2><div class="content">{{$v}}</div></div>{{end}}
<script src="/katex/katex.min.js"></script>
<script src="/katex/contrib/auto-render.min.js"></script>
<script>
window.addEventListener("DOMContentLoaded", function(){
  if (window.renderMathInElement) {
    renderMathInElement(document.body, {
      delimiters: [
        {left: "$$", right: "$$", display: true},
        {left: "$", right: "$", display: false},
        {left: "\\\(", right: "\\\)", display: false},
        {left: "\\[", right: "\\]", display: true}
      ],
      throwOnError: false
    });
  }
});
</script>
</body>
</html>`))

// ---------- CSS / JS（以 template.JS 注入，避免模板引擎解析） ----------

