// 由 server.go 拆分生成（V1.6.0 模块化，功能零变化）
package web

const indexCSS = `
:root{--bg:#f7f8fa;--panel:#fff;--border:#e1e4e8;--text:#24292f;--muted:#586069;--accent:#0969da;--accent-bg:#ddf4ff;--sidebar:#f5f6f8;--hover:#eef2f5}
*{box-sizing:border-box}
html,body{margin:0;height:100%;overflow:hidden}
body{font-family:-apple-system,"Segoe UI","Microsoft YaHei",sans-serif;background:var(--bg);color:var(--text);display:flex;flex-direction:column}
#topbar{display:flex;align-items:center;justify-content:space-between;height:48px;padding:0 14px;background:var(--panel);border-bottom:1px solid var(--border);position:sticky;top:0;z-index:10}
.brand{font-size:16px;font-weight:600}
.brand small{color:var(--muted);font-weight:400}
.bankbar{display:flex;gap:8px;align-items:center}
select,input,button{font-family:inherit;font-size:13px}
#bankSel{width:220px;padding:6px 8px;border:1px solid var(--border);border-radius:6px;background:var(--panel)}
button{padding:6px 12px;border:1px solid var(--border);border-radius:6px;background:var(--panel);cursor:pointer}
button:hover{background:var(--hover)}
#filtersWrap{display:flex;flex-direction:column}
#filtersWrap.hidden{display:none}
#filtersBar{display:flex;align-items:center}
#filters{flex:1;display:flex;flex-wrap:wrap;gap:8px;padding:8px 14px;background:var(--panel);align-items:center;overflow-y:auto}
#filtersBar button{padding:4px 8px;margin:4px;font-size:12px}
#filtersResizer{height:5px;cursor:row-resize;flex-shrink:0;background:transparent}
#filtersResizer:hover,#filtersResizer.active{background:var(--accent-bg)}
#filters .f-item{display:flex;align-items:center;gap:4px;color:var(--muted);font-size:12px}
#filters select{max-width:150px;padding:4px 6px;border:1px solid var(--border);border-radius:6px}
#filters .clear{border:none;background:none;color:var(--accent);cursor:pointer;font-size:12px}
#body{flex:1;min-height:0;display:flex;overflow:hidden}
#sidebarWrap{display:flex}
#sidebarWrap.hidden{display:none}
#sidebarCol{display:flex;flex-direction:column;min-width:0;border-right:1px solid var(--border)}
#sidebarHead{display:flex;align-items:center;justify-content:space-between;padding:4px 8px;font-size:12px;color:var(--muted);background:var(--sidebar);border-bottom:1px solid var(--border)}
#sidebarHead button{padding:2px 6px;font-size:12px}
#sideTabs{display:flex;gap:2px}
.stab{border:1px solid transparent;background:none;color:var(--muted);cursor:pointer;font-size:12px;padding:3px 8px;border-radius:6px}
.stab.active{background:var(--accent-bg);color:var(--accent)}
.sel-cb{margin-right:6px;accent-color:var(--accent)}
#sidebar{flex:1;width:260px;min-width:120px;max-width:480px;background:var(--sidebar);overflow-y:auto;padding:8px 6px}
#resizer{width:5px;cursor:col-resize;flex-shrink:0;background:transparent}
#resizer:hover,#resizer.active{background:var(--accent-bg)}
button.pinned{background:var(--accent-bg);color:var(--accent);border-color:var(--accent)}
#main{flex:1;overflow-y:auto;padding:0}
#mainToolbar{display:flex;align-items:center;gap:6px;padding:8px 14px;background:var(--panel);border-bottom:1px solid var(--border);position:sticky;top:0;z-index:5}
#mainToolbar .mode{padding:4px 10px;font-size:12px}
#mainToolbar .mode.active{background:var(--accent-bg);color:var(--accent);border-color:var(--accent)}
/* V3.0.0：pi 聊天面板 */
#piPanel{width:0;min-width:0;display:flex;flex-direction:column;background:var(--panel);border-left:1px solid var(--border);transition:width .2s;overflow:hidden}
#piPanel.open{width:340px;min-width:340px}
#piPanelHead{display:flex;align-items:center;gap:8px;padding:8px 12px;font-size:13px;border-bottom:1px solid var(--border);flex-shrink:0}
#piPanelHead button{margin-left:auto;padding:2px 8px}
#piMsgs{flex:1;overflow-y:auto;padding:10px;display:flex;flex-direction:column;gap:8px;font-size:13px}
.pi-msg{max-width:92%;padding:8px 10px;border-radius:8px;white-space:pre-wrap;word-break:break-word}
.pi-msg.user{align-self:flex-end;background:var(--accent-bg);color:var(--accent)}
.pi-msg.pi{align-self:flex-start;background:var(--sidebar);border:1px solid var(--border)}
.pi-msg.waiting{color:var(--muted);font-style:italic}
#piInputBar{display:flex;gap:6px;padding:8px;border-top:1px solid var(--border);flex-shrink:0}
#piInput{flex:1;font-family:inherit;font-size:13px;padding:6px;border:1px solid var(--border);border-radius:6px;resize:none}
#piSend{padding:6px 14px;align-self:flex-end}
#mainContent{padding:16px}
.qbox{background:var(--panel);border:1px solid var(--border);border-radius:8px;padding:12px 16px;margin-bottom:12px}
.qbox.active{border-color:var(--accent)}
.qbox-head{display:flex;align-items:center;gap:8px;margin-bottom:6px}
.qbox-id{font-weight:600;font-size:14px}
.qbox-tags{flex:1;display:flex;gap:4px;flex-wrap:wrap}
.qbox-actions{display:flex;gap:4px}
.fav-btn{border:none;background:none;font-size:16px;cursor:pointer;padding:2px}
.exp-btn{padding:3px 10px;font-size:12px}
.qbox-detail{margin-top:8px;border-top:1px dashed var(--border);padding-top:6px}
.sec-btn{display:block;width:100%;text-align:left;padding:6px 10px;margin:4px 0;background:var(--sidebar);border:1px solid var(--border);border-radius:6px;cursor:pointer;font-size:13px;font-weight:600}
.sec-body{padding:8px 10px}
.split-wrap{display:flex;gap:14px;height:calc(100vh - 140px)}
.split-left{width:42%;min-width:280px;overflow-y:auto;padding-right:6px}
.split-right{flex:1;overflow-y:auto;border:1px solid var(--border);border-radius:8px;padding:14px 18px;background:var(--panel)}
.card-nav{display:flex;align-items:center;justify-content:center;gap:12px;margin-top:12px}
.card-nav span{color:var(--muted);font-size:13px}
.split-right .qbox{margin-bottom:12px}
.pkg{font-size:13px}
.pkg-head{display:flex;align-items:center;gap:4px;padding:5px 8px;border-radius:6px;cursor:pointer;color:var(--text);font-weight:600}
.pkg-head:hover{background:var(--hover)}
.pkg-head.sel{background:var(--accent-bg);color:var(--accent)}
.dir-exp{cursor:pointer;display:inline-block;width:14px;flex-shrink:0}
.dir-name{cursor:pointer;flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.pkg-head .cnt{color:var(--muted);font-weight:400;font-size:12px}
.pkg-body{padding-left:18px}
.q-item{padding:4px 8px;border-radius:6px;cursor:pointer;font-size:13px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.q-item:hover{background:var(--hover)}
.q-item.active{background:var(--accent-bg)}
.card{background:var(--panel);border:1px solid var(--border);border-radius:8px;padding:14px 16px;margin-bottom:12px}
.card h3{margin:0 0 6px;font-size:15px}
.card .qmeta{color:var(--muted);font-size:12px;margin-bottom:8px}
.tag{display:inline-block;background:var(--accent-bg);color:var(--accent);border-radius:12px;padding:1px 9px;margin:2px;font-size:12px}
.content{white-space:pre-wrap;background:#f6f8fa;padding:12px;border-radius:6px;font-size:14px;line-height:1.6}
pre{white-space:pre-wrap;background:#f6f8fa;padding:12px;border-radius:6px;font-size:14px;line-height:1.6}
.detail{border-top:1px dashed var(--border);margin-top:10px;padding-top:8px}
.empty{color:var(--muted);text-align:center;padding:40px}
#modal{position:fixed;inset:0;background:rgba(0,0,0,.35);display:none;align-items:center;justify-content:center;z-index:100}
#modal.show{display:flex}
#editModal{position:fixed;inset:0;background:rgba(0,0,0,.35);display:none;align-items:center;justify-content:center;z-index:100}
#editModal.show{display:flex}
#displayModal{position:fixed;inset:0;background:rgba(0,0,0,.35);display:none;align-items:center;justify-content:center;z-index:100}
#displayModal.show{display:flex}
#exportModal{position:fixed;inset:0;background:rgba(0,0,0,.35);display:none;align-items:center;justify-content:center;z-index:100}
#exportModal.show{display:flex}
#testSetupModal{position:fixed;inset:0;background:rgba(0,0,0,.35);display:none;align-items:center;justify-content:center;z-index:100}
#testSetupModal.show{display:flex}
#imgModal{position:fixed;inset:0;background:rgba(0,0,0,.35);display:none;align-items:center;justify-content:center;z-index:100}
#imgModal.show{display:flex}
.md-img{max-width:100%;border-radius:8px;margin:8px 0;display:block}
#imgModal code{background:#f6f8fa;padding:1px 6px;border-radius:4px;font-size:12px;word-break:break-all}
#displayList{display:flex;flex-direction:column;gap:4px;margin:10px 0}
#displayList label{display:flex;align-items:center;gap:6px;font-size:13px;cursor:pointer}
#editModal .modal-box{width:560px;max-width:92%;max-height:85vh;overflow-y:auto}
#editModal textarea{width:100%;font-family:inherit;font-size:13px;padding:6px;border:1px solid var(--border);border-radius:6px;margin:2px 0 8px;resize:vertical}
#editModal .lbl{font-size:12px;color:var(--muted)}
#editModal input[type=text]{width:100%;font-family:inherit;font-size:13px;padding:6px;border:1px solid var(--border);border-radius:6px;margin:2px 0 8px}
#editMeta{display:flex;flex-direction:column;gap:6px;margin-bottom:8px}
#editMeta.folded{display:none}
.meta-row{display:grid;grid-template-columns:80px 1fr 28px;gap:6px;align-items:center}
.meta-row .cust{grid-column:2;display:flex;gap:4px}
.meta-row .cust input{flex:1;padding:4px 6px;border:1px solid var(--border);border-radius:6px;font-size:13px}
.meta-row .cust .hint{font-size:11px;color:var(--muted);white-space:nowrap}
.meta-del{border:none;background:none;color:var(--muted);cursor:pointer;font-size:14px;padding:2px}
.meta-del:hover{color:#d1242f}
.q-actions{display:flex;gap:8px;margin-bottom:10px}
.q-actions button{font-size:12px;padding:4px 10px}
#reloadBtn{padding:6px 10px}
.test-btn{padding:2px 8px;font-size:11px;margin-left:4px;background:var(--accent-bg);color:var(--accent);border-color:var(--accent)}
.test-bar{display:flex;align-items:center;gap:10px;padding:10px 14px;background:var(--panel);border:1px solid var(--border);border-radius:8px;margin-bottom:12px}
.test-bar .cnt{color:var(--muted)}
.test-bar button{margin-left:auto}
.test-q .qbox-head{border-bottom:1px dashed var(--border);padding-bottom:6px}
.test-answer{margin-top:8px;padding-top:6px;border-top:1px dashed var(--border)}
.test-answer b{font-size:13px;color:var(--muted)}
.opt-group{display:flex;flex-direction:column;gap:4px;margin-top:6px}
.opt{display:flex;align-items:center;gap:6px;font-size:14px;padding:4px 8px;border:1px solid var(--border);border-radius:6px;cursor:pointer}
.opt:hover{background:var(--hover)}
.ans-input{width:100%;max-width:400px;padding:8px;margin-top:6px;border:1px solid var(--border);border-radius:6px;font-size:14px}
.ans-textarea{width:100%;padding:8px;margin-top:6px;border:1px solid var(--border);border-radius:6px;font-size:14px;resize:vertical;font-family:inherit}
.test-done{margin:16px auto;display:block;padding:10px 30px;font-size:14px;background:var(--accent);color:#fff;border:none;border-radius:8px;cursor:pointer}
.test-timer{font-weight:bold;color:var(--accent)}
.sel-tip{display:flex;align-items:center;gap:6px;padding:8px 14px;background:var(--accent-bg);color:var(--accent);border-radius:8px;margin-bottom:12px;font-size:13px}
.grade{font-size:13px;font-weight:bold;margin-left:8px}
.grade.right{color:#1a7f37}
.grade.wrong{color:#d1242f}
.report-ans{margin-top:8px;font-size:13px;color:var(--muted)}
.report-ref{margin-top:6px;font-size:13px}
.report-ref .content{background:var(--sidebar);padding:8px;border-radius:6px;margin-top:4px}
.report-grade{margin-top:8px;font-size:13px}
.subj-score{width:70px;padding:4px 6px;border:1px solid var(--border);border-radius:6px;margin:0 6px}
.subj-ok{padding:4px 12px;font-size:12px}
.modal-box{background:var(--panel);border-radius:10px;padding:20px;width:420px;max-width:90%}
#modal .modal-box h4{margin:10px 0 6px;font-size:13px;color:var(--text)}
.cfg-tabs{display:flex;gap:6px;margin-bottom:10px}
.cfg-tab{padding:5px 14px;border:1px solid var(--border);border-radius:6px;background:var(--panel);cursor:pointer;font-size:13px}
.cfg-tab.active{background:var(--accent-bg);color:var(--accent);border-color:var(--accent)}
.cfg-item{display:flex;align-items:center;justify-content:space-between;gap:8px;padding:4px 8px;border:1px solid var(--border);border-radius:6px;margin:4px 0;font-size:13px}
.cfg-item .path{color:var(--muted);font-size:12px;word-break:break-all}
.cfg-item .vals{color:var(--muted);font-size:12px}
.cfg-del{border:none;background:none;color:var(--muted);cursor:pointer;font-size:14px}
.cfg-del:hover{color:#d1242f}
.cfg-open{padding:2px 8px;font-size:12px}
#modal code{background:#f6f8fa;padding:1px 6px;border-radius:4px;font-size:12px;word-break:break-all}
.modal-box h3{margin:0 0 8px}
.tip{color:var(--muted);font-size:12px}
#linkInput{width:100%;padding:8px;margin:10px 0;border:1px solid var(--border);border-radius:6px}
.modal-actions{display:flex;gap:8px;justify-content:flex-end}
`


const indexJS = `
var state = { banks: [], bank: "", tree: [], filters: {}, expanded: {}, mode: "list", favOnly: false, cardIdx: 0, favs: {}, sideTab: "tree", selectMode: false, selected: {}, selDirs: [], selList: null, testing: null, pendingTest: null, timerInt: null };

function qs(s){ return document.querySelector(s); }
function esc(s){ var d=document.createElement("div"); d.textContent = (s==null?"":s); return d.innerHTML; }
function tagName(k){
  var names = {chapter:"章节",grade:"年级",difficulty:"难度",importance:"重要性",source:"来源",knowledge:"知识点",type:"题型"};
  return names[k] || k;
}

// KaTeX 公式渲染配置与函数（div.content 内 $...$/$$...$$ 自动渲染）
var katexDelims = [
  {left: "$$", right: "$$", display: true},
  {left: "$", right: "$", display: false},
  {left: "\\\(", right: "\\\)", display: false},
  {left: "\\[", right: "\\]", display: true}
];
function renderMath(el){
  if (window.renderMathInElement) {
    renderMathInElement(el, {delimiters: katexDelims, throwOnError: false});
  }
}

function init(){
  qs("#settingsBtn").onclick = openSettings;
  qs("#linkCancel").onclick = closeSettings;
  qs("#linkOk").onclick = doLink;
  qs("#linkInput").addEventListener("keydown", function(e){ if(e.key==="Enter") doLink(); });
  qs("#bankSel").onchange = function(){
    state.bank = this.value;
    state.filters = {};
    loadAll();
  };
  qs("#toggleSidebar").onclick = function(){
    var h = qs("#sidebarWrap").classList.toggle("hidden");
    this.textContent = h ? "☰" : "☷";
    this.title = h ? "显示侧边栏" : "隐藏侧边栏";
  };
  qs("#toggleFilters").onclick = function(){
    var h = qs("#filtersWrap").classList.toggle("hidden");
    this.textContent = h ? "⌃" : "⌄";
    this.title = h ? "显示筛选栏" : "隐藏筛选栏";
  };
  qs("#pinSidebar").onclick = function(){ this.classList.toggle("pinned"); };
  qs("#pinFilters").onclick = function(){ this.classList.toggle("pinned"); };
  qs("#reloadBtn").onclick = reloadBank;
  qs("#editSave").onclick = saveEdit;
  qs("#editCancel").onclick = closeEdit;
  qs("#editAddField").onclick = addField;
  qs("#favFilterBtn").onclick = function(){
    state.favOnly = !state.favOnly;
    this.textContent = state.favOnly ? "★ 收藏（仅看）" : "☆ 收藏";
    this.classList.toggle("active", state.favOnly);
    loadAll();
  };
  qs("#selectBtn").onclick = toggleSelectMode;
  qs("#exportBtn").onclick = openExport;
  qs("#saveListBtn").onclick = saveSelectedAsList;
  qs("#exportOk").onclick = doExport;
  qs("#tsOk").onclick = confirmTestSetup;
  qs("#imgOk").onclick = doImgUpload;
  qs("#imgCancel").onclick = closeImgUpload;
  qs("#tsCancel").onclick = function(){ qs("#testSetupModal").classList.remove("show"); state.pendingTest = null; };
  qs("#tsTimer").onchange = function(){ qs("#tsMinWrap").hidden = (this.value !== "countdown"); };
  qs("#exportCancel").onclick = closeExport;
  qs("#displayBtn").onclick = openDisplay;
  qs("#displayOk").onclick = saveDisplay;
  qs("#displayCancel").onclick = closeDisplay;
  // V3.0.0：pi 聊天面板
  qs("#piChatBtn").onclick = function(){ qs("#piPanel").classList.add("open"); qs("#piInput").focus(); };
  qs("#piPanelClose").onclick = function(){ qs("#piPanel").classList.remove("open"); };
  qs("#piSend").onclick = sendPiMsg;
  qs("#piInput").addEventListener("keydown", function(e){ if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); sendPiMsg(); } });
  // V1.1.7.0：点主区域空白取消目录/题组选择
  qs("#mainContent").addEventListener("click", function(e){
    if ((e.target === this || e.target.classList.contains("empty")) && (state.selDirs.length > 0 || state.selList)) {
      state.selDirs = [];
      state.selList = null;
      renderSidebar();
      renderMain();
    }
  });
  document.querySelectorAll("#sideTabs .stab").forEach(function(b){
    b.onclick = function(){ switchSideTab(b.getAttribute("data-tab")); };
  });
  document.querySelectorAll("#mainToolbar .mode").forEach(function(b){
    b.onclick = function(){ setMode(b.getAttribute("data-mode")); };
  });
  // V1.4.3：键盘导航（双栏/卡片模式：上/左=上一题，下/右=下一题）
  document.addEventListener("keydown", function(e){
    if (state.mode !== "split" && state.mode !== "card") return;
    var t = e.target;
    if (t && (t.tagName === "INPUT" || t.tagName === "TEXTAREA" || t.tagName === "SELECT")) return;
    if (e.key === "ArrowUp" || e.key === "ArrowLeft") { navQuestion(-1); e.preventDefault(); }
    else if (e.key === "ArrowDown" || e.key === "ArrowRight") { navQuestion(1); e.preventDefault(); }
  });
  qs("#metaFoldBtn").onclick = function(){
    var folded = qs("#editMeta").classList.toggle("folded");
    this.textContent = folded ? "▸ 元数据" : "▾ 元数据";
  };
  loadBanks();
}

// 侧边栏拖拽调宽（拖到最窄自动隐藏）
(function(){
  var resizer = qs("#resizer"), sb = qs("#sidebar"), wrap = qs("#sidebarWrap");
  var sx = 0, sw = 0;
  resizer.addEventListener("mousedown", function(e){
    sx = e.clientX; sw = sb.offsetWidth;
    resizer.classList.add("active");
    document.addEventListener("mousemove", onMove);
    document.addEventListener("mouseup", onUp);
    e.preventDefault();
  });
  function onMove(e){
    var w = sw + (e.clientX - sx);
    if (w < 100) {
      if (qs("#pinSidebar").classList.contains("pinned")) {
        w = 100; // 固定时只调宽不隐藏
      } else {
        wrap.classList.add("hidden");
        qs("#toggleSidebar").textContent = "☰";
        qs("#toggleSidebar").title = "显示侧边栏";
        onUp();
        return;
      }
    }
    if (w > 480) w = 480;
    sb.style.width = w + "px";
  }
  function onUp(){
    resizer.classList.remove("active");
    document.removeEventListener("mousemove", onMove);
    document.removeEventListener("mouseup", onUp);
  }
})();

// 筛选栏拖拽调高（拖到最矮自动隐藏）
(function(){
  var resizer = qs("#filtersResizer"), f = qs("#filters"), wrap = qs("#filtersWrap");
  var sy = 0, sh = 0;
  resizer.addEventListener("mousedown", function(e){
    sy = e.clientY; sh = f.offsetHeight;
    resizer.classList.add("active");
    document.addEventListener("mousemove", onMove);
    document.addEventListener("mouseup", onUp);
    e.preventDefault();
  });
  function onMove(e){
    var h = sh + (e.clientY - sy);
    if (h < 40) {
      if (qs("#pinFilters").classList.contains("pinned")) {
        h = 40; // 固定时只调高不隐藏
      } else {
        wrap.classList.add("hidden");
        qs("#toggleFilters").textContent = "⌃";
        qs("#toggleFilters").title = "显示筛选栏";
        onUp();
        return;
      }
    }
    if (h > 160) h = 160;
    f.style.height = h + "px";
  }
  function onUp(){
    resizer.classList.remove("active");
    document.removeEventListener("mousemove", onMove);
    document.removeEventListener("mouseup", onUp);
  }
})();

function loadBanks(){
  fetch("/api/banks").then(function(r){ return r.json(); }).then(function(banks){
    state.banks = banks;
    var sel = qs("#bankSel");
    sel.innerHTML = "";
    banks.forEach(function(b){
      var o = document.createElement("option");
      o.value = b.dir; o.text = b.name + "（" + b.count + " 题）";
      if (b.current){ o.selected = true; state.bank = b.dir; }
      sel.appendChild(o);
    });
    loadAll();
  });
}

function loadAll(){
  fetch("/api/tree?bank=" + encodeURIComponent(state.bank)).then(function(r){ return r.json(); }).then(function(tree){
    state.tree = tree;
    renderSidebar();
    renderFilters();
    // 加载收藏状态（V1.5.1：带 bank 参数 + 用 state.bank 建 key）
    fetch("/api/favorites?bank=" + encodeURIComponent(state.bank)).then(function(r){ return r.json(); }).then(function(favs){
      state.favs = {};
      favs.forEach(function(f){ state.favs[state.bank + "|" + f.id] = true; });
      renderMain();
    });
  });
}

// 模式切换
function setMode(m){
  state.mode = m;
  document.querySelectorAll("#mainToolbar .mode").forEach(function(b){
    b.classList.toggle("active", b.getAttribute("data-mode") === m);
  });
  if (m === "card" && state.cardIdx >= visibleQuestions().length) state.cardIdx = 0;
  renderMain();
}

// V3.0.0：与 pi 对话（题库为工作目录）
function sendPiMsg(){
  var msg = qs("#piInput").value.trim();
  if (!msg) return;
  qs("#piInput").value = "";
  var box = qs("#piMsgs");
  box.innerHTML += '<div class="pi-msg user">' + esc(msg) + '</div>';
  var wait = document.createElement("div");
  wait.className = "pi-msg pi waiting";
  wait.textContent = "⏳ pi 思考中…";
  box.appendChild(wait);
  box.scrollTop = box.scrollHeight;
  fetch("/api/pi/chat", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({message: msg})
  }).then(function(r){ return r.json(); }).then(function(d){
    wait.className = "pi-msg pi";
    if (d.ok) wait.textContent = d.reply;
    else wait.textContent = "❌ " + (d.error || "调用失败");
    box.scrollTop = box.scrollHeight;
  }).catch(function(){
    wait.className = "pi-msg pi";
    wait.textContent = "❌ 请求失败";
  });
}

function renderMain(){
  if (state.testing) { renderTest(); return; }
  if (state.mode === "split") renderSplit();
  else if (state.mode === "card") renderCard();
  else renderList();
}

function isFav(q){
  return !!(state.favs[state.bank + "|" + q.id]);
}

function toggleFav(q, btn){
  fetch("/api/favorite", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({bank: state.bank, id: q.id})
  }).then(function(r){ return r.json(); }).then(function(d){
    if (d.ok) {
      if (d.favorited) state.favs[state.bank + "|" + q.id] = true;
      else delete state.favs[state.bank + "|" + q.id];
      if (btn) btn.textContent = d.favorited ? "⭐" : "☆";
      if (state.favOnly) renderMain();
    }
  });
}

// ---------- V1.5.0：侧边栏页签 / 选择 / 导出 ----------

// 切换侧边栏页签（题库 / 收藏 / 清单）
function switchSideTab(tab){
  state.sideTab = tab;
  document.querySelectorAll("#sideTabs .stab").forEach(function(b){
    b.classList.toggle("active", b.getAttribute("data-tab") === tab);
  });
  renderSidebar();
}

// 收藏导航
function renderFavSidebar(sb){
  fetch("/api/favorites?bank=" + encodeURIComponent(state.bank)).then(function(r){ return r.json(); }).then(function(favs){
    sb.innerHTML = "";
    if (favs.length === 0) { sb.innerHTML = '<div class="empty">暂无收藏（题目上点 ☆ 收藏）</div>'; return; }
    favs.forEach(function(f){
      var it = document.createElement("div");
      it.className = "q-item";
      it.textContent = f.id + " · " + (f.title || f.file);
      it.title = f.file;
      it.onclick = function(){ selectQuestion({id: f.id}); };
      sb.appendChild(it);
    });
  });
}

// 清单导航
function renderListsSidebar(sb){
  fetch("/api/lists?bank=" + encodeURIComponent(state.bank)).then(function(r){ return r.json(); }).then(function(lists){
    sb.innerHTML = "";
    if (lists.length === 0) { sb.innerHTML = '<div class="empty">暂无题组（选择题目后「📋 存题组」）</div>'; return; }
    lists.forEach(function(l){
      var head = document.createElement("div");
      head.className = "pkg-head";
      var exp = document.createElement("span");
      exp.className = "dir-exp";
      exp.textContent = "▸";
      exp.onclick = function(e){
        e.stopPropagation();
        var body = head.nextElementSibling;
        if (body) {
          var open = body.style.display !== "none";
          body.style.display = open ? "none" : "block";
          exp.textContent = open ? "▸" : "▾";
        }
      };
      var nm = document.createElement("span");
      nm.className = "dir-name";
      nm.textContent = "📁 " + l.name;
      // V1.7.1：点击题组名 → 主区域只查看该题组题目（非测试模式）
      nm.onclick = function(e){
        e.stopPropagation();
        if (state.selList && state.selList.name === l.name) state.selList = null;
        else state.selList = {name: l.name, ids: l.ids};
        renderSidebar();
        renderMain();
      };
      head.classList.toggle("sel", !!(state.selList && state.selList.name === l.name));
      var cnt = document.createElement("span");
      cnt.className = "cnt";
      cnt.textContent = l.count;
      var testBtn = document.createElement("button");
      testBtn.className = "test-btn";
      testBtn.textContent = "▶ 测试";
      testBtn.onclick = function(e){
        e.stopPropagation();
        startTest(l);
      };
      head.appendChild(exp); head.appendChild(nm); head.appendChild(cnt); head.appendChild(testBtn);
      sb.appendChild(head);
      var body = document.createElement("div");
      body.className = "pkg-body";
      body.style.display = "none";
      (l.questions || []).forEach(function(q){
        var it = document.createElement("div");
        it.className = "q-item";
        it.textContent = q.id + " · " + (q.title || q.file);
        it.onclick = function(){ selectQuestion({id: q.id}); };
        body.appendChild(it);
      });
      sb.appendChild(body);
    });
  });
}

// ---------- V1.7.0：在线测试模式（批次 B：答题基础） ----------

// 开始测试：先弹测试设置（计分/计时）
function startTest(list){
  state.pendingTest = list;
  qs("#testSetupName").textContent = "（" + list.name + " · " + (list.questions||[]).length + " 题）";
  qs("#testSetupModal").classList.add("show");
}

// 确认设置并开始
function confirmTestSetup(){
  var list = state.pendingTest;
  var timerType = qs("#tsTimer").value;
  var minutes = parseInt(qs("#tsMinutes").value, 10) || 10;
  state.testing = {
    name: list.name, qs: (list.questions || []), answers: {},
    scored: false, timerType: timerType, minutes: minutes,
    elapsed: 0, remainMs: (timerType === "countdown" ? minutes * 60000 : 0), report: null
  };
  state.sideTab = "tree";
  qs("#testSetupModal").classList.remove("show");
  if (state.timerInt) clearInterval(state.timerInt);
  state.timerInt = setInterval(tickTimer, 1000);
  renderTest();
}

// 计时器每秒
function tickTimer(){
  if (!state.testing || state.testing.report) return;
  if (state.testing.timerType === "countup") {
    state.testing.elapsed++;
    var el = qs("#testTimer");
    if (el) el.textContent = "⏲ " + fmtTime(state.testing.elapsed);
  } else if (state.testing.timerType === "countdown") {
    state.testing.remainMs -= 1000;
    var el2 = qs("#testTimer");
    if (el2) el2.textContent = "⏳ " + fmtTime(Math.ceil(state.testing.remainMs / 1000));
    if (state.testing.remainMs <= 0) finishTest();
  }
}

function fmtTime(sec){
  var m = Math.floor(sec / 60), s = sec % 60;
  return m + ":" + (s < 10 ? "0" : "") + s;
}

// 渲染测试（主区域只显示题目 + 答题区）
function renderTest(){
  var main = qs("#mainContent");
  main.innerHTML = "";
  var t = state.testing;
  if (!t) return;
  var bar = document.createElement("div");
  bar.className = "test-bar";
  bar.innerHTML = "<b>🧪 测试：" + esc(t.name) + "</b> <span class='cnt'>" + t.qs.length + " 题</span>";
  if (t.timerType !== "none") bar.innerHTML += ' <span id="testTimer" class="test-timer">' + (t.timerType === "countdown" ? "⏳ " + fmtTime(Math.ceil(t.remainMs/1000)) : "⏲ 0:00") + "</span>";
  var exitBtn = document.createElement("button");
  exitBtn.textContent = "✕ 退出";
  exitBtn.onclick = function(){ state.testing = null; clearInterval(state.timerInt); renderMain(); };
  bar.appendChild(exitBtn);
  main.appendChild(bar);
  if (t.qs.length === 0) { main.innerHTML += '<div class="empty">题组为空</div>'; return; }
  t.qs.forEach(function(q, i){
    var card = document.createElement("div");
    card.className = "qbox test-q";
    card.setAttribute("data-id", q.id);
    card.innerHTML =
      '<div class="qbox-head"><span class="qbox-id">' + (i + 1) + ". " + esc(q.id) + '</span>' +
      '<span class="qbox-tags">' + metaTagsOf(q.meta) + '</span></div>' +
      '<div class="qbox-prompt content">' + mdImages(esc(q.prompt || q.title), state.bank) + '</div>' +
      answerArea(q);
    main.appendChild(card);
    renderMath(card);
  });
  var done = document.createElement("button");
  done.className = "test-done";
  done.textContent = "✅ 完成测试";
  done.onclick = finishTest;
  main.appendChild(done);
}

// 完成测试：收集答案 → 展示报告（V1.7.2：用户自评，无自动评分）
function finishTest(){
  if (!state.testing || state.testing.report) return;
  collectAnswers();
  var t = state.testing;
  if (t.timerType === "countup") t.elapsed = Math.round(t.elapsed);
  else if (t.timerType === "countdown") t.elapsed = Math.round((t.minutes * 60000 - t.remainMs) / 1000);
  clearInterval(state.timerInt);
  t.report = true;
  renderReport();
}

// 渲染报告
// 渲染报告（V1.7.2：我的答案 vs 参考答案，用户自评）
function renderReport(){
  var main = qs("#mainContent");
  main.innerHTML = "";
  var t = state.testing;
  var bar = document.createElement("div");
  bar.className = "test-bar report-bar";
  bar.innerHTML = "<b>📊 测试报告：" + esc(t.name) + "</b> <span class='cnt'>" + t.qs.length + " 题" +
    (t.timerType !== "none" ? " · 用时 " + fmtTime(t.elapsed) : "") + "</span>";
  var backBtn = document.createElement("button");
  backBtn.textContent = "← 返回";
  backBtn.onclick = function(){ state.testing = null; clearInterval(state.timerInt); renderMain(); };
  bar.appendChild(backBtn);
  main.appendChild(bar);
  t.qs.forEach(function(q, i){
    var card = document.createElement("div");
    card.className = "qbox";
    var myAns = t.answers[q.id] ? esc(t.answers[q.id]) : "（未作答）";
    card.innerHTML =
      '<div class="qbox-head"><span class="qbox-id">' + (i + 1) + ". " + esc(q.id) + '</span></div>' +
      '<div class="qbox-prompt content">' + esc(q.prompt || q.title) + "</div>" +
      '<div class="report-ans"><b>我的答案：</b><div class="content">' + myAns + "</div></div>" +
      '<div class="report-ref"><b>参考答案：</b><div class="content">' + esc(q.answer || "（无）") + "</div></div>";
    main.appendChild(card);
    renderMath(card);
  });
}

// 答题区（V1.7.2：统一单一作答区，用户自评）
function answerArea(q){
  return '<div class="test-answer"><b>作答</b> <textarea class="ans-textarea" data-id="' + esc(q.id) + '" rows="4" placeholder="作答"></textarea></div>';
}


// 收集答案
function collectAnswers(){
  if (!state.testing) return;
  state.testing.answers = {};
  document.querySelectorAll("#mainContent .test-q").forEach(function(card){
    var id = card.getAttribute("data-id");
    var checked = card.querySelector("input[type=radio]:checked");
    if (checked) { state.testing.answers[id] = checked.value; return; }
    var inp = card.querySelector(".ans-input");
    if (inp && inp.value.trim()) { state.testing.answers[id] = inp.value.trim(); return; }
    var ta = card.querySelector(".ans-textarea");
    if (ta && ta.value.trim()) { state.testing.answers[id] = ta.value.trim(); }
  });
}

// 选择模式切换
function toggleSelectMode(){
  state.selectMode = !state.selectMode;
  var btn = qs("#selectBtn");
  btn.classList.toggle("active", state.selectMode);
  btn.textContent = state.selectMode ? "☑ 选择中" : "☑ 选择";
  qs("#saveListBtn").style.display = state.selectMode ? "" : "none";
  qs("#exportBtn").style.display = state.selectMode ? "" : "none";
  renderMain();
}

function toggleSelect(q){
  var key = q.id;
  if (state.selected[key]) delete state.selected[key];
  else state.selected[key] = true;
}

// 选中题目存为清单（组卷）
function saveSelectedAsList(){
  var ids = Object.keys(state.selected);
  if (ids.length === 0) { alert("请先勾选题目"); return; }
  var name = prompt("题组名称（如：期中复习卷）");
  if (!name || !name.trim()) return;
  fetch("/api/lists/save", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({bank: state.bank, name: name.trim(), ids: ids})
  }).then(function(r){ return r.json(); }).then(function(d){
    if (d.ok) { alert("✅ 已保存题组「" + name.trim() + "」（" + ids.length + " 题）"); }
  });
}

// 操作导出文件（open=默认程序打开 / select=资源管理器定位）
function openExportFile(p, action){
  fetch("/api/export/open", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({path: p, action: action || "select"})
  }).then(function(r){ return r.json(); }).then(function(d){
    if (!d.ok) alert(d.error || "操作失败");
  });
}

// V1.8.0：图片上传
var imgQ = null;
function openImgUpload(q){
  imgQ = q;
  qs("#imgQ").textContent = "（" + q.id + "）";
  qs("#imgFile").value = "";
  qs("#imgMsg").textContent = "";
  qs("#imgModal").classList.add("show");
}
function closeImgUpload(){ qs("#imgModal").classList.remove("show"); }
function doImgUpload(){
  var f = qs("#imgFile").files[0];
  if (!f || !imgQ) { qs("#imgMsg").textContent = "请选择文件"; return; }
  var fd = new FormData();
  fd.append("bank", state.bank);
  fd.append("id", imgQ.id);
  fd.append("file", f);
  qs("#imgMsg").textContent = "上传中…";
  fetch("/api/image/upload", {method: "POST", body: fd})
    .then(function(r){ return r.json(); })
    .then(function(d){
      if (d.ok) {
        qs("#imgMsg").innerHTML = "✅ 已上传<br>引用语法：<code>" + esc(d.markdown) + "</code><br><span style='color:var(--muted)'>编辑题目粘贴即可显示</span>";
      } else qs("#imgMsg").textContent = "❌ " + (d.error || "上传失败");
    }).catch(function(){ qs("#imgMsg").textContent = "❌ 请求失败"; });
}

// 导出弹窗
function openExport(){
  var ids = Object.keys(state.selected);
  if (ids.length === 0) { alert("请先勾选题目"); return; }
  qs("#exportCount").textContent = "（" + ids.length + " 道）";
  qs("#exportMsg").textContent = "";
  qs("#exportModal").classList.add("show");
}
function closeExport(){ qs("#exportModal").classList.remove("show"); }
function doExport(){
  var ids = Object.keys(state.selected);
  var parts = [];
  qs("#exportParts").querySelectorAll("input:checked").forEach(function(cb){ parts.push(cb.value); });
  var format = qs("#exportFormat").value;
  if (parts.length === 0) { qs("#exportMsg").textContent = "请至少勾选一个部分"; return; }
  fetch("/api/export", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({bank: state.bank, ids: ids, parts: parts, format: format})
  }).then(function(r){ return r.json(); }).then(function(d){
    if (d.ok) {
      qs("#exportMsg").innerHTML = "✅ 已导出 " + d.count + " 题 → " + esc(d.path) + ' <button id="openExportFile">📂 打开</button> <button id="locateExportFile">📍 定位</button>';
      var ob = qs("#openExportFile");
      if (ob) ob.onclick = function(){ openExportFile(d.path, "open"); };
      var lb = qs("#locateExportFile");
      if (lb) lb.onclick = function(){ openExportFile(d.path, "select"); };
    } else qs("#exportMsg").textContent = "❌ " + (d.error || "导出失败");
  });
}

function renderSidebar(){
  var sb = qs("#sidebar");
  if (state.sideTab === "fav") { renderFavSidebar(sb); return; }
  if (state.sideTab === "lists") { renderListsSidebar(sb); return; }
  sb.innerHTML = "";
  if (!state.tree || !state.tree.root) { sb.innerHTML = '<div class="empty">题库为空</div>'; return; }
  var root = state.tree.root;
  // 根目录题目（直接放在题库根）
  root.questions.forEach(function(q){ sb.appendChild(buildQItem(q)); });
  // 递归渲染子目录
  root.dirs.forEach(function(d){ sb.appendChild(renderDirNode(d, 0)); });
  if (root.dirs.length === 0 && root.questions.length === 0) {
    sb.innerHTML = '<div class="empty">题库为空</div>';
  }
}

// 目录节点：名称点击=筛选（Ctrl 多选），箭头=展开/收起
function renderDirNode(dir, depth){
  var wrap = document.createElement("div");
  var row = document.createElement("div");
  row.className = "pkg-head" + (state.selDirs.indexOf(dir.path) >= 0 ? " sel" : "");
  row.style.paddingLeft = (6 + depth * 14) + "px";
  var exp = document.createElement("span");
  exp.className = "dir-exp";
  exp.textContent = state.expanded[dir.path] ? "▾" : "▸";
  exp.onclick = function(e){
    e.stopPropagation();
    state.expanded[dir.path] = !state.expanded[dir.path];
    renderSidebar();
  };
  var nm = document.createElement("span");
  nm.className = "dir-name";
  nm.textContent = "📁 " + dir.name;
  var cnt = document.createElement("span");
  cnt.className = "cnt";
  cnt.textContent = (dir.questions.length + countSub(dir));
  row.appendChild(exp); row.appendChild(nm); row.appendChild(cnt);
  // 点击名称 → 筛选目录（Ctrl 多选）
  row.onclick = function(e){
    var idx = state.selDirs.indexOf(dir.path);
    if (e.ctrlKey || e.metaKey) {
      if (idx >= 0) state.selDirs.splice(idx, 1); else state.selDirs.push(dir.path);
    } else {
      state.selDirs = (idx >= 0 && state.selDirs.length === 1) ? [] : [dir.path];
    }
    renderSidebar();
    renderMain();
  };
  wrap.appendChild(row);
  if (state.expanded[dir.path]) {
    dir.dirs.forEach(function(d){ wrap.appendChild(renderDirNode(d, depth + 1)); });
    var body = document.createElement("div");
    body.className = "pkg-body";
    body.style.paddingLeft = (6 + (depth + 1) * 14) + "px";
    dir.questions.forEach(function(q){ body.appendChild(buildQItem(q)); });
    if (dir.questions.length) wrap.appendChild(body);
  }
  return wrap;
}

function countSub(dir){
  var n = dir.questions.length;
  dir.dirs.forEach(function(d){ n += countSub(d); });
  return n;
}

// 题目叶节点
function buildQItem(q){
  var it = document.createElement("div");
  it.className = "q-item";
  it.textContent = q.id + " · " + (q.title ? q.title : q.file);
  it.title = q.file;
  it.onclick = function(){
    selectQuestion(q);
    document.querySelectorAll(".q-item").forEach(function(x){ x.classList.remove("active"); });
    it.classList.add("active");
  };
  return it;
}

function aggregateTags(){
  var map = {};
  walkTree(function(q){
    var keys = ["chapter","grade","difficulty","importance","source","knowledge","type"];
    keys.forEach(function(k){
      var v = q.meta[k];
      if (v) {
        if (!map[k]) map[k] = {};
        map[k][v] = true;
      }
    });
  });
  return map;
}

// 递归遍历树中所有题目
function walkTree(fn){
  if (!state.tree || !state.tree.root) return;
  var walk = function(n){
    n.questions.forEach(fn);
    n.dirs.forEach(walk);
  };
  walk(state.tree.root);
}

function renderFilters(){
  var bar = qs("#filters");
  bar.innerHTML = "";
  var map = aggregateTags();
  Object.keys(map).forEach(function(k){
    var f = document.createElement("span");
    f.className = "f-item";
    f.innerHTML = esc(tagName(k)) + " ";
    var sel = document.createElement("select");
    var opt = document.createElement("option");
    opt.value = ""; opt.text = "全部";
    sel.appendChild(opt);
    Object.keys(map[k]).sort().forEach(function(v){
      var o = document.createElement("option");
      o.value = v; o.text = v;
      if (state.filters[k] === v) o.selected = true;
      sel.appendChild(o);
    });
    f.appendChild(sel);
    sel.onchange = function(){
      var v = this.value;
      if (v) state.filters[k] = v; else delete state.filters[k];
      renderMain();
    };
    bar.appendChild(f);
  });
  if (Object.keys(map).length > 0) {
    var btn = document.createElement("button");
    btn.className = "clear";
    btn.textContent = "清空筛选";
    btn.onclick = function(){ state.filters = {}; renderFilters(); renderMain(); };
    bar.appendChild(btn);
  }
}

function visibleQuestions(){
  var out = [];
  walkTree(function(q){
    if (state.favOnly && !isFav(q)) return;
    if (state.selDirs.length > 0 && !inSelDirs(q)) return;
    // V1.7.1：题组筛选（非测试模式查看题组内题目）
    if (state.selList) {
      var inList = false;
      state.selList.ids.forEach(function(id){ if (id === q.id) inList = true; });
      if (!inList) return;
    }
    var hit = true;
    for (var k in state.filters) {
      if (q.meta[k] !== state.filters[k]) { hit = false; break; }
    }
    if (hit) out.push({ pkg: dirOf(q), q: q });
  });
  return out;
}

// V1.7.0：目录筛选（selDirs 多选，含子目录）
function dirOf(q){
  var rel = q.rel || "";
  var idx = rel.lastIndexOf("/");
  return idx >= 0 ? rel.slice(0, idx) : "";
}
function inSelDirs(q){
  var d = dirOf(q);
  if (d === "") return false; // 根目录题目不在任何选中文件夹
  for (var i = 0; i < state.selDirs.length; i++) {
    var s = state.selDirs[i];
    if (d === s || d.indexOf(s + "/") === 0) return true;
  }
  return false;
}

// ---------- V1.4.0：显示清单（自定义显示字段） ----------

function defaultDisplay(){ return ["type","difficulty","importance","source"]; }
function loadDisplay(){
  try {
    var d = JSON.parse(localStorage.getItem("reqDisplay"));
    if (d && d.length) return d;
  } catch(e){}
  return defaultDisplay();
}

// 打开显示清单面板（字段来自全局配置 meta_fields + 常见扩展）
function openDisplay(){
  fetch("/api/config/global").then(function(r){ return r.json(); }).then(function(g){
    var fields = (g.meta_fields || []).map(function(f){ return f.name; });
    ["chapter","grade","knowledge"].forEach(function(k){ if (fields.indexOf(k) < 0) fields.push(k); });
    var cur = loadDisplay();
    var box = qs("#displayList");
    box.innerHTML = "";
    fields.forEach(function(f){
      var lab = document.createElement("label");
      var cb = document.createElement("input");
      cb.type = "checkbox";
      cb.value = f;
      cb.checked = cur.indexOf(f) >= 0;
      lab.appendChild(cb);
      lab.appendChild(document.createTextNode(tagName(f)));
      box.appendChild(lab);
    });
    qs("#displayModal").classList.add("show");
  });
}
function closeDisplay(){ qs("#displayModal").classList.remove("show"); }
function saveDisplay(){
  var sel = [];
  qs("#displayList").querySelectorAll("input:checked").forEach(function(cb){ sel.push(cb.value); });
  localStorage.setItem("reqDisplay", JSON.stringify(sel));
  closeDisplay();
  renderMain();
}

// V1.8.0：解析 Markdown 图片语法 ![alt](image/xxx) → img 标签
function mdImages(text, bank){
  return text.replace(/!\[([^\]]*)\]\(([^)]+)\)/g, function(m, alt, src){
    var f = src;
    if (f.indexOf("./") === 0) f = f.slice(2);
    if (f.indexOf("image/") !== 0) return m;
    var url = "/image?bank=" + encodeURIComponent(bank || state.bank) + "&file=" + encodeURIComponent(f);
    return '<img src="' + url + '" alt="' + esc(alt) + '" class="md-img" onerror="this.remove()">';
  });
}

function metaTagsOf(meta){
  var html = "";
  loadDisplay().forEach(function(k){
    var v = meta[k];
    if (v) html += '<span class="tag">' + esc(tagName(k)) + ": " + esc(v) + "</span>";
  });
  return html;
}

// ---------- V1.4.0：题目盒子 / 双栏 / 卡片 ----------

// 题目盒子：默认只显示题干 + 展开按钮；答案/解析/备注分段折叠
function buildBox(it, opts){
  opts = opts || {};
  var q = it.q;
  var box = document.createElement("div");
  box.className = "qbox" + (opts.active ? " active" : "");
  box.setAttribute("data-id", q.id);
  var meta = metaTagsOf(q.meta);
  var cbHtml = "";
  if (state.selectMode) cbHtml = '<input type="checkbox" class="sel-cb" ' + (state.selected[q.id] ? "checked" : "") + '>';
  box.innerHTML =
    '<div class="qbox-head">' + cbHtml + '<span class="qbox-id">' + esc(q.id) + '</span>' +
    '<span class="qbox-tags">' + meta + '</span>' +
    '<span class="qbox-actions"><button class="fav-btn" title="收藏">' + (isFav(q) ? "⭐" : "☆") + '</button>' +
    '<button class="exp-btn">▼ 展开</button></span></div>' +
    '<div class="qbox-prompt content">' + mdImages(esc(q.prompt || q.title), state.bank) + '</div>' +
    '<div class="qbox-detail" hidden></div>';
  box.querySelector(".fav-btn").onclick = function(e){
    e.stopPropagation();
    toggleFav(q, this);
  };
  var selCb = box.querySelector(".sel-cb");
  if (selCb) {
    selCb.onclick = function(e){ e.stopPropagation(); toggleSelect(q); };
  }
  box.querySelector(".exp-btn").onclick = function(e){
    e.stopPropagation();
    toggleExpand(box, q);
  };
  if (opts.onclick) box.onclick = function(){ opts.onclick(box, q); };
  renderMath(box);
  return box;
}

// 展开/收起题目详情（答案/解析/备注分段，各自点击显示）
function toggleExpand(box, q){
  var det = box.querySelector(".qbox-detail");
  var btn = box.querySelector(".exp-btn");
  if (det.hidden) {
    det.hidden = false;
    btn.textContent = "▲ 收起";
    if (det.getAttribute("data-loaded") !== "1") {
      det.innerHTML = '<span class="tip">加载中…</span>';
      fetch("/api/question?bank=" + encodeURIComponent(state.bank) + "&id=" + encodeURIComponent(q.id)).then(function(r){ return r.json(); }).then(function(d){
        var html = '<div class="q-actions"><button id="btnOpen">📂 打开本地</button><button id="btnEdit">✏️ 编辑</button><button id="btnImg">📷 图片</button></div>';
        var secs = [["答案", d.answer], ["解析", d.explain], ["备注", d.note], ["链接笔记", d.links]];
        secs.forEach(function(s){
          if (s[1]) html += '<div class="sec"><button class="sec-btn" data-title="' + esc(s[0]) + '">▸ ' + esc(s[0]) + '</button><div class="sec-body" hidden><div class="content">' + esc(s[1]) + "</div></div></div>";
        });
        if (!html) html = '<span class="tip">（无更多内容）</span>';
        det.innerHTML = html;
        det.querySelector("#btnOpen").onclick = function(){ openLocal(q); };
        det.querySelector("#btnEdit").onclick = function(){ openEdit(d, q); };
        det.querySelector("#btnImg").onclick = function(){ openImgUpload(q); };
        det.querySelectorAll(".sec-btn").forEach(function(sb){
          sb.onclick = function(e){
            e.stopPropagation(); // 阻止冒泡到盒子（避免收起详情）
            var body = sb.parentElement.querySelector(".sec-body");
            body.hidden = !body.hidden;
            sb.textContent = (body.hidden ? "▸ " : "▾ ") + sb.getAttribute("data-title");
          };
        });
        renderMath(det);
        det.setAttribute("data-loaded", "1");
      });
    }
  } else {
    det.hidden = true;
    btn.textContent = "▼ 展开";
  }
}

// 列表模式：题目盒子流
function renderList(){
  var main = qs("#mainContent");
  main.innerHTML = "";
  var items = visibleQuestions();
  // V1.7.1：题组查看提示条
  if (state.selList) {
    var tip = document.createElement("div");
    tip.className = "sel-tip";
    tip.innerHTML = "📁 查看题组：<b>" + esc(state.selList.name) + "</b>（" + items.length + " 题，点空白取消）";
    main.appendChild(tip);
  }
  if (items.length === 0) {
    main.innerHTML += '<div class="empty">没有符合条件的题目' + (state.favOnly ? "（收藏中）" : "") + "</div>";
    return;
  }
  items.forEach(function(it){
    var box = buildBox(it, {onclick: function(b, q){ toggleExpand(b, q); }});
    main.appendChild(box);
  });
}

// 双栏模式：左列表 + 右详情
function renderSplit(){
  var main = qs("#mainContent");
  main.innerHTML = "";
  var items = visibleQuestions();
  if (items.length === 0) {
    main.innerHTML = '<div class="empty">没有符合条件的题目</div>';
    return;
  }
  var wrap = document.createElement("div");
  wrap.className = "split-wrap";
  var left = document.createElement("div");
  left.className = "split-left";
  var right = document.createElement("div");
  right.className = "split-right";
  right.innerHTML = '<span class="tip">← 点击左侧题目查看详情</span>';
  var selKey = state.splitSel;
  items.forEach(function(it, i){
    var q = it.q;
    var active = (selKey === (it.pkg + "|" + q.id));
    if (i === 0 && !selKey) active = true;
    var box = buildBox(it, {active: active});
    box.querySelector(".exp-btn").style.display = "none";
    box.onclick = function(){
      state.splitSel = it.pkg + "|" + q.id;
      document.querySelectorAll(".split-left .qbox").forEach(function(b){ b.classList.remove("active"); });
      box.classList.add("active");
      loadSplitDetail(right, q);
    };
    left.appendChild(box);
    if (active) {
      state.splitSel = it.pkg + "|" + q.id;
      loadSplitDetail(right, q);
    }
  });
  wrap.appendChild(left);
  wrap.appendChild(right);
  main.appendChild(wrap);
}

// 双栏右侧详情（完整展示，公式渲染）
function loadSplitDetail(right, q){
  fetch("/api/question?bank=" + encodeURIComponent(state.bank) + "&id=" + encodeURIComponent(q.id)).then(function(r){ return r.json(); }).then(function(d){
    var html = '<div class="q-actions"><button id="btnOpen">📂 打开本地</button><button id="btnEdit">✏️ 编辑</button><button id="btnImg">📷 图片</button></div>';
    html += '<div class="qbox-head"><span class="qbox-id">' + esc(q.id) + "</span><span style='flex:1'></span>";
    html += '<button class="fav-btn">' + (isFav(q) ? "⭐" : "☆") + "</button></div>";
    html += metaTagsOf(d.meta || {});
    if (d.prompt) html += "<div><b>题目</b><div class='content'>" + mdImages(esc(d.prompt), state.bank) + "</div></div>";
    if (d.answer) html += "<div><b>答案</b><div class='content'>" + mdImages(esc(d.answer), state.bank) + "</div></div>";
    if (d.explain) html += "<div><b>解析</b><div class='content'>" + mdImages(esc(d.explain), state.bank) + "</div></div>";
    if (d.note) html += "<div><b>备注</b><div class='content'>" + mdImages(esc(d.note), state.bank) + "</div></div>";
    if (d.links) html += "<div><b>链接笔记</b><div class='content'>" + mdImages(esc(d.links), state.bank) + "</div></div>";
    right.innerHTML = html;
    var fb = right.querySelector(".fav-btn");
    if (fb) fb.onclick = function(){ toggleFav(q, this); };
    var bo = right.querySelector("#btnOpen");
    if (bo) bo.onclick = function(){ openLocal(q); };
    var be = right.querySelector("#btnEdit");
    if (be) be.onclick = function(){ openEdit(d, q); };
    var bi = right.querySelector("#btnImg");
    if (bi) bi.onclick = function(){ openImgUpload(q); };
    renderMath(right);
  });
}

// 卡片模式：单题 + 前进后退
function renderCard(){
  var main = qs("#mainContent");
  main.innerHTML = "";
  var items = visibleQuestions();
  if (items.length === 0) {
    main.innerHTML = '<div class="empty">没有符合条件的题目</div>';
    return;
  }
  if (state.cardIdx >= items.length) state.cardIdx = 0;
  var it = items[state.cardIdx];
  var q = it.q;
  var box = buildBox(it);
  toggleExpand(box, q); // V1.4.2：卡片模式自动一级展开
  var nav = document.createElement("div");
  nav.className = "card-nav";
  var prev = document.createElement("button"); prev.textContent = "◀ 上一题";
  var cnt = document.createElement("span"); cnt.textContent = (state.cardIdx + 1) + " / " + items.length;
  var next = document.createElement("button"); next.textContent = "下一题 ▶";
  prev.onclick = function(){ state.cardIdx = (state.cardIdx - 1 + items.length) % items.length; renderCard(); };
  next.onclick = function(){ state.cardIdx = (state.cardIdx + 1) % items.length; renderCard(); };
  nav.appendChild(prev); nav.appendChild(cnt); nav.appendChild(next);
  main.appendChild(box);
  main.appendChild(nav);
}

// V1.4.3：键盘导航上一题/下一题
function navQuestion(delta){
  var items = visibleQuestions();
  if (items.length === 0) return;
  if (state.mode === "card") {
    state.cardIdx = (state.cardIdx + delta + items.length) % items.length;
    renderCard();
  } else if (state.mode === "split") {
    var idx = 0;
    for (var i = 0; i < items.length; i++) {
      if (state.splitSel === (items[i].pkg + "|" + items[i].q.id)) { idx = i; break; }
    }
    idx = (idx + delta + items.length) % items.length;
    state.splitSel = items[idx].pkg + "|" + items[idx].q.id;
    renderSplit();
  }
}

// 选中题目（侧边栏点击）：按模式处理
function selectQuestion(q){
  if (state.mode === "card") {
    var items = visibleQuestions();
    for (var i = 0; i < items.length; i++) {
      if (items[i].q.id === q.id) { state.cardIdx = i; break; }
    }
    renderCard();
  } else if (state.mode === "split") {
    state.splitSel = q.id;
    renderSplit();
  } else {
    var box = document.querySelector('.qbox[data-id="' + q.id + '"]');
    if (box) {
      box.scrollIntoView({block: "center"});
      toggleExpand(box, q);
    }
  }
}

function objectLen(obj){ var n = 0; for (var k in obj) n++; return n; }

function loadDetail(q, det, card){
  fetch("/api/question?bank=" + encodeURIComponent(state.bank) + "&id=" + encodeURIComponent(q.id)).then(function(r){ return r.json(); }).then(function(d){
    var html = '<div class="q-actions">';
    html += '<button id="btnOpen">📂 打开本地</button>';
    html += '<button id="btnEdit">✏️ 编辑</button>';
    html += '</div>';
    if (d.prompt) html += "<div><b>题目</b><div class=\"content\">" + mdImages(esc(d.prompt), state.bank) + "</div></div>";
    if (d.answer) html += "<div><b>答案</b><div class=\"content\">" + mdImages(esc(d.answer), state.bank) + "</div></div>";
    if (d.explain) html += "<div><b>解析</b><div class=\"content\">" + mdImages(esc(d.explain), state.bank) + "</div></div>";
    if (d.note) html += "<div><b>备注</b><div class=\"content\">" + mdImages(esc(d.note), state.bank) + "</div></div>";
    if (!html) html = '<span class="tip">（无更多内容）</span>';
    det.innerHTML = html;
    det.querySelector("#btnOpen").onclick = function(){ openLocal(q); };
    det.querySelector("#btnEdit").onclick = function(){ openEdit(d, q); };
    renderMath(det);
    card.classList.add("active");
  }).catch(function(){ det.innerHTML = '<span class="tip">加载失败</span>'; });
}

// 刷新题库：重新扫描目录（同步本地/网页改动）
function reloadBank(){
  fetch("/api/reload", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({bank: state.bank})
  }).then(function(r){ return r.json(); }).then(function(d){
    if (d.ok) loadAll();
  });
}

// 打开本地文件（资源管理器定位）
// 打开本地：Obsidian iframe 环境 → postMessage 通知插件（插件判断库内外）；否则 explorer
function openLocal(q){
  if (window.parent !== window) {
    // Obsidian：通知插件打开（库内新标签页 / 库外资源管理器）
    fetch("/api/question?bank=" + encodeURIComponent(state.bank) + "&id=" + encodeURIComponent(q.id))
      .then(function(r){ return r.json(); })
      .then(function(d){
        window.parent.postMessage({type: "requiz-open", path: d.path}, "*");
      });
    return;
  }
  fetch("/api/open", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({bank: state.bank, id: q.id})
  }).then(function(r){ return r.json(); }).then(function(d){
    if (!d.ok) alert(d.error || "打开失败");
  });
}

// 编辑弹窗
var editing = null;
function openEdit(d, q){
  editing = q;
  qs("#editId").textContent = q.id;
  qs("#editFile").value = q.file;
  qs("#editPrompt").value = d.prompt || "";
  qs("#editAnswer").value = d.answer || "";
  qs("#editExplain").value = d.explain || "";
  qs("#editNote").value = d.note || "";
  qs("#editLinks").value = d.links || "";
  qs("#editMsg").textContent = "";
  // 异步加载字段已知值后构建元数据行
  fetch("/api/meta-values?bank=" + encodeURIComponent(state.bank)).then(function(r){ return r.json(); }).then(function(values){
    buildMetaRows(d.meta || {}, values);
  });
  qs("#editModal").classList.add("show");
}
function buildMetaRows(meta, values){
  var metaBox = qs("#editMeta");
  metaBox.innerHTML = "";
  var keys = ["id","chapter","grade","difficulty","importance","source","knowledge","type"];
  var shown = {};
  keys.forEach(function(k){ if (meta[k]) { addMetaRow(metaBox, k, meta[k], (values||{})[k] || []); shown[k] = true; } });
  keys.forEach(function(k){ if (!shown[k]) addMetaRow(metaBox, k, "", (values||{})[k] || []); });
  Object.keys(meta).forEach(function(k){
    if (keys.indexOf(k) < 0 && ["app","bank","path"].indexOf(k) < 0) addMetaRow(metaBox, k, meta[k], (values||{})[k] || []);
  });
}
function addMetaRow(box, k, v, knownVals){
  var row = document.createElement("div");
  row.className = "meta-row";
  var lab = document.createElement("span");
  lab.className = "lbl";
  lab.textContent = tagName(k);
  lab.style.paddingTop = "5px";
  var sel = document.createElement("select");
  sel.setAttribute("data-key", k);
  var opt = document.createElement("option"); opt.value = ""; opt.text = "（空）";
  sel.appendChild(opt);
  (knownVals || []).forEach(function(vv){
    var o = document.createElement("option"); o.value = vv; o.text = vv;
    if (vv === v) o.selected = true;
    sel.appendChild(o);
  });
  // 当前值不在已知列表 → 显示为「值(自定义)」并选中
  if (v && (knownVals||[]).indexOf(v) < 0) {
    var co = document.createElement("option"); co.value = v; co.text = v + "(自定义)"; co.selected = true;
    sel.appendChild(co);
  }
  // 新增值 / 自定义 两个特殊选项
  var addOpt = document.createElement("option"); addOpt.value = "__add__"; addOpt.text = "➕ 新增值…";
  sel.appendChild(addOpt);
  var onceOpt = document.createElement("option"); onceOpt.value = "__once__"; onceOpt.text = "✍️ 自定义…";
  sel.appendChild(onceOpt);
  var cdiv = document.createElement("span");
  cdiv.className = "cust";
  cdiv.style.display = "none";
  var inp = document.createElement("input");
  var hint = document.createElement("span");
  hint.className = "hint";
  cdiv.appendChild(inp); cdiv.appendChild(hint);
  sel.onchange = function(){
    if (this.value === "__add__") {
      cdiv.style.display = "flex"; inp.value = ""; hint.textContent = "将加入题库配置";
    } else if (this.value === "__once__") {
      cdiv.style.display = "flex"; inp.value = ""; hint.textContent = "仅用于本题";
    } else {
      cdiv.style.display = "none"; inp.value = "";
    }
  };
  // 删除字段按钮
  var del = document.createElement("button");
  del.className = "meta-del"; del.textContent = "✕"; del.title = "删除该字段";
  del.onclick = function(){ row.remove(); };
  row.appendChild(lab); row.appendChild(sel); row.appendChild(del); row.appendChild(cdiv);
  box.appendChild(row);
}
function addField(){
  var box = qs("#editMeta");
  var existing = qs("#newFieldRow");
  if (existing) { existing.querySelector("input").focus(); return; }
  // 在编辑栏内展开输入行（不用浏览器 prompt 弹窗）
  var row = document.createElement("div");
  row.id = "newFieldRow";
  row.style.display = "flex";
  row.style.gap = "4px";
  row.style.alignItems = "center";
  var inp = document.createElement("input");
  inp.placeholder = "输入新字段名（如：出处）";
  var ok = document.createElement("button"); ok.textContent = "添加";
  var cancel = document.createElement("button"); cancel.textContent = "取消";
  row.appendChild(inp); row.appendChild(ok); row.appendChild(cancel);
  box.appendChild(row);
  inp.focus();
  function commit(){
    var name = inp.value.trim();
    if (!name) return;
    var exist = false;
    box.querySelectorAll("select[data-key]").forEach(function(s){ if (s.getAttribute("data-key") === name) exist = true; });
    if (exist) { alert("字段已存在"); return; }
    addMetaRow(box, name, "", []);
    row.remove();
  }
  ok.onclick = commit;
  cancel.onclick = function(){ row.remove(); };
  inp.addEventListener("keydown", function(e){ if (e.key === "Enter") commit(); if (e.key === "Escape") row.remove(); });
}
function closeEdit(){ qs("#editModal").classList.remove("show"); }
function saveEdit(){
  if (!editing) return;
  var meta = {};
  var addToConfig = [];
  qs("#editMeta").querySelectorAll("select[data-key]").forEach(function(sel){
    var k = sel.getAttribute("data-key");
    var val = "";
    if (sel.value === "__add__" || sel.value === "__once__") {
      var inp = sel.parentElement.querySelector("input");
      val = inp ? inp.value.trim() : "";
      if (val !== "" && sel.value === "__add__") addToConfig.push({k: k, v: val});
    } else {
      val = sel.value;
    }
    if (val !== "") meta[k] = val;
  });
  // 先写入配置（新增值），再保存题目
  var doSave = function(){
    var body = {
      bank: state.bank, id: editing.id,
      file: qs("#editFile").value.trim(),
      meta: meta,
      prompt: qs("#editPrompt").value,
      answer: qs("#editAnswer").value,
      explain: qs("#editExplain").value,
      note: qs("#editNote").value
    };
    fetch("/api/question/save", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify(body)
    }).then(function(r){ return r.json(); }).then(function(d){
      if (d.ok) {
        qs("#editMsg").textContent = "✅ 已保存";
        closeEdit();
        loadAll();
      } else {
        qs("#editMsg").textContent = "❌ " + (d.error || "保存失败");
      }
    });
  };
  if (addToConfig.length > 0) {
    var pending = addToConfig.length, failed = false;
    addToConfig.forEach(function(ac){
      fetch("/api/meta-value/add", {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({bank: state.bank, field: ac.k, value: ac.v})
      }).then(function(r){ return r.json(); }).then(function(d){
        if (!d.ok) failed = true;
        if (--pending === 0) {
          if (failed) qs("#editMsg").textContent = "⚠️ 部分新值写入配置失败";
          doSave();
        }
      });
    });
  } else {
    doSave();
  }
}

function openSettings(){
  qs("#modal").classList.add("show");
  qs("#linkMsg").textContent = "";
  qs("#linkInput").value = "";
  loadCfgGlobal();
  loadCfgProject();
}
function closeSettings(){ qs("#modal").classList.remove("show"); }

// 页签切换
qs("#tabGlobal").onclick = function(){
  qs("#tabGlobal").classList.add("active");
  qs("#tabProject").classList.remove("active");
  qs("#cfgGlobal").hidden = false;
  qs("#cfgProject").hidden = true;
};
qs("#tabProject").onclick = function(){
  qs("#tabProject").classList.add("active");
  qs("#tabGlobal").classList.remove("active");
  qs("#cfgGlobal").hidden = true;
  qs("#cfgProject").hidden = false;
};

// 加载全局配置
function loadCfgGlobal(){
  Promise.all([
    fetch("/api/config/global").then(function(r){ return r.json(); }),
    fetch("/api/banks").then(function(r){ return r.json(); })
  ]).then(function(res){
    var g = res[0], banks = res[1];
    qs("#cfgGlobalPath").textContent = g.path;
    // 题库目录 → 名称映射
    var nameMap = {};
    banks.forEach(function(b){ nameMap[b.dir] = b.name; });
    // 默认配置
    var dbox = qs("#cfgDefaults");
    dbox.innerHTML = "";
    Object.keys(g.defaults || {}).forEach(function(k){
      var it = document.createElement("div");
      it.className = "cfg-item";
      it.innerHTML = "<span>" + esc(k) + "</span><span class=\"vals\">" + esc(g.defaults[k]) + "</span>";
      dbox.appendChild(it);
    });
    // 题库列表（全部对等，当前打开的不可移除）
    var lbox = qs("#cfgLinks");
    lbox.innerHTML = "";
    (g.links || []).forEach(function(l, i){
      var isCur = (l === state.bank);
      var it = document.createElement("div");
      it.className = "cfg-item";
      it.innerHTML = "<span><b>" + esc(nameMap[l] || l) + "</b> " + (isCur ? "<b style=\"color:var(--accent)\">（当前）</b>" : "") + "<div class=\"path\">" + esc(l) + "</div></span>" + (isCur ? "<span class=\"vals\">不可移除</span>" : "<span style=\"display:flex;gap:4px\"><button class=\"cfg-open\" title=\"打开该题库\">打开</button><button class=\"cfg-del\" title=\"移除\">✕</button></span>");
      if (!isCur) {
        it.querySelector(".cfg-open").onclick = function(){ openBank(l); };
        it.querySelector(".cfg-del").onclick = function(){ removeLink(i); };
      }
      lbox.appendChild(it);
    });
    if (!g.links || g.links.length === 0) lbox.innerHTML = '<span class="tip">（暂无题库，请添加）</span>';
    // 字段定义
    var fbox = qs("#cfgFields");
    fbox.innerHTML = "";
    (g.meta_fields || []).forEach(function(f){
      var it = document.createElement("div");
      it.className = "cfg-item";
      it.innerHTML = "<span>" + esc(f.label || f.name) + " <small style=\"color:var(--muted)\">" + esc(f.name) + "</small></span><span class=\"vals\">" + esc((f.values||[]).join(" / ")) + "</span>";
      fbox.appendChild(it);
    });
  });
}

// 打开题库（切换下拉栏选中）
function openBank(dir){
  var sel = qs("#bankSel");
  sel.value = dir;
  state.bank = dir;
  state.filters = {};
  closeSettings();
  loadAll();
}

// 加载项目配置
function loadCfgProject(){
  fetch("/api/config/project?bank=" + encodeURIComponent(state.bank)).then(function(r){ return r.json(); }).then(function(p){
    qs("#cfgProjectPath").textContent = p.path;
    qs("#cfgProjectInfo").innerHTML =
      "<div class=\"cfg-item\"><span>题库名</span><span class=\"vals\">" + esc(p.bank) + "</span></div>" +
      "<div class=\"cfg-item\"><span>运行软件</span><span class=\"vals\">" + esc(p.app) + "</span></div>";
    var fbox = qs("#cfgProjectFields");
    fbox.innerHTML = "";
    (p.meta_fields || []).forEach(function(f){
      var it = document.createElement("div");
      it.className = "cfg-item";
      it.innerHTML = "<span>" + esc(f.label || f.name) + "</span><span class=\"vals\">" + esc((f.values||[]).join(" / ")) + "</span>";
      fbox.appendChild(it);
    });
    if (!p.meta_fields || p.meta_fields.length === 0) fbox.innerHTML = '<span class="tip">（暂无自定义字段）</span>';
  });
}

// 解除链接（写全局配置）
function removeLink(i){
  fetch("/api/config/global").then(function(r){ return r.json(); }).then(function(g){
    var links = g.links || [];
    links.splice(i, 1);
    return fetch("/api/config/global/save", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({links: links})
    });
  }).then(function(r){ return r.json(); }).then(function(d){
    if (d.ok) {
      loadCfgGlobal();
      loadBanks();
    }
  });
}

function doLink(){
  var dir = qs("#linkInput").value.trim();
  if (!dir) { qs("#linkMsg").textContent = "请输入题库目录路径"; return; }
  fetch("/api/link", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({dir: dir})
  }).then(function(r){
    return r.json().then(function(d){ return {ok: r.ok, d: d}; });
  }).then(function(res){
    if (res.ok) {
      qs("#linkMsg").textContent = "✅ 已链接";
      closeSettings();
      loadBanks();
    } else {
      qs("#linkMsg").textContent = "❌ " + (res.d.error || "链接失败");
    }
  }).catch(function(){ qs("#linkMsg").textContent = "❌ 请求失败"; });
}

init();
`

