// 图片及附件（V1.8.0）：题库/image/<题目名>/ 目录约定
package web

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"requiz/src/model"
)

// imageRoot 题库图片根目录（题库/image/）
func imageRoot(b *model.Bank) string {
	return filepath.Join(b.Dir, "image")
}

// contentTypes 按扩展名推断 Content-Type
var contentTypes = map[string]string{
	".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg",
	".gif": "image/gif", ".webp": "image/webp", ".svg": "image/svg+xml",
	".bmp": "image/bmp", ".ico": "image/x-icon", ".pdf": "application/pdf",
	".txt": "text/plain; charset=utf-8", ".md": "text/plain; charset=utf-8",
	".zip": "application/zip", ".doc": "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".mp3": "audio/mpeg", ".mp4": "video/mp4",
}

func contentTypeOf(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if ct, ok := contentTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

// GET /image?bank=<题库>&file=<相对image路径> ：返回图片/附件（附件下载）
func apiImageHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := s.bankByDir(r.URL.Query().Get("bank"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		file := r.URL.Query().Get("file")
		if file == "" {
			http.Error(w, "缺少 file 参数", http.StatusBadRequest)
			return
		}
		// 兼容 md 引用前缀 image/
		file = strings.TrimPrefix(file, "image/")
		if strings.Contains(file, "..") {
			http.Error(w, "非法路径", http.StatusForbidden)
			return
		}
		root := imageRoot(b)
		path := filepath.Join(root, filepath.FromSlash(file))
		rel, err := filepath.Rel(root, path)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			http.Error(w, "路径越界", http.StatusForbidden)
			return
		}
		data, err := os.ReadFile(path)
		if err != nil {
			http.Error(w, "文件不存在", http.StatusNotFound)
			return
		}
		ct := contentTypeOf(path)
		w.Header().Set("Content-Type", ct)
		if strings.HasPrefix(ct, "image/") {
			w.Header().Set("Cache-Control", "no-store")
		} else {
			w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(path))
		}
		w.Write(data)
	}
}

// POST /api/image/upload ：上传图片/附件到 题库/image/<题目名>/
func apiImageUploadHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "仅支持 POST", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseMultipartForm(20 << 20); err != nil { // 20MB
			http.Error(w, `{"error":"文件过大或表单解析失败"}`, http.StatusBadRequest)
			return
		}
		bankDir := r.FormValue("bank")
		id := r.FormValue("id")
		if bankDir == "" || id == "" {
			http.Error(w, `{"error":"需要 bank 与 id"}`, http.StatusBadRequest)
			return
		}
		b, err := s.bankByDir(bankDir)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusNotFound)
			return
		}
		q, err := b.Find(id)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusNotFound)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, `{"error":"缺少 file 字段"}`, http.StatusBadRequest)
			return
		}
		defer file.Close()
		fname := filepath.Base(header.Filename) // 防路径注入
		if fname == "." || fname == "" {
			http.Error(w, `{"error":"文件名非法"}`, http.StatusBadRequest)
			return
		}
		dir := filepath.Join(imageRoot(b), strings.TrimSuffix(q.File, ".md"))
		if err := os.MkdirAll(dir, 0755); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		dst := filepath.Join(dir, fname)
		out, err := os.Create(dst)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		defer out.Close()
		if _, err := io.Copy(out, file); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		rel := "image/" + strings.TrimSuffix(q.File, ".md") + "/" + fname
		writeJSON(w, map[string]any{"ok": true, "path": rel, "markdown": "![" + fname + "](" + rel + ")"})
	}
}
