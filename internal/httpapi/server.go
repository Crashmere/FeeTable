package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"unicode"

	reportexport "feetable/internal/export"
	"feetable/internal/feetable"
)

type server struct {
	store      *feetable.Store
	assets     fs.FS
	exportSlot chan struct{}
}

func New(store *feetable.Store, assets fs.FS) http.Handler {
	s := &server{store: store, assets: assets, exportSlot: make(chan struct{}, 1)}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		reply(w, map[string]string{"status": "ok"}, store.Ping(r.Context()))
	})
	mux.HandleFunc("GET /api/tables", func(w http.ResponseWriter, r *http.Request) {
		in := feetable.TableQuery{Page: page(r), Sort: r.URL.Query().Get("sort")}
		if r.URL.Query().Has("year") || r.URL.Query().Has("month") {
			var err error
			in.Year, err = strconv.Atoi(r.URL.Query().Get("year"))
			if err != nil || in.Year < 1 {
				reply(w, nil, &feetable.Error{Code: "VALIDATION", Message: "年份无效"})
				return
			}
			in.Month, err = strconv.Atoi(r.URL.Query().Get("month"))
			if err != nil || in.Month < 1 {
				reply(w, nil, &feetable.Error{Code: "VALIDATION", Message: "月份无效"})
				return
			}
		}
		v, e := store.Tables(r.Context(), in)
		reply(w, v, e)
	})
	mux.HandleFunc("POST /api/tables", func(w http.ResponseWriter, r *http.Request) {
		var in feetable.TableInput
		if !decode(w, r, &in) {
			return
		}
		v, e := store.CreateTable(r.Context(), in)
		reply(w, v, e)
	})
	mux.HandleFunc("PUT /api/tables/{id}", func(w http.ResponseWriter, r *http.Request) {
		var in feetable.TableInput
		if !decode(w, r, &in) {
			return
		}
		v, e := store.UpdateTable(r.Context(), id(r, "id"), in)
		reply(w, v, e)
	})
	mux.HandleFunc("DELETE /api/tables/{id}", func(w http.ResponseWriter, r *http.Request) {
		reply(w, map[string]bool{"ok": true}, store.DeleteTable(r.Context(), id(r, "id"), revision(r)))
	})
	mux.HandleFunc("POST /api/tables/{id}/merge", func(w http.ResponseWriter, r *http.Request) {
		var in feetable.MergeInput
		if !decode(w, r, &in) {
			return
		}
		v, e := store.MergeTables(r.Context(), id(r, "id"), in)
		reply(w, v, e)
	})
	mux.HandleFunc("GET /api/tables/{id}", func(w http.ResponseWriter, r *http.Request) {
		v, e := store.Report(r.Context(), id(r, "id"), tag(r), page(r), false)
		reply(w, v, e)
	})
	mux.HandleFunc("GET /api/tables/{id}/report", func(w http.ResponseWriter, r *http.Request) {
		v, e := store.Report(r.Context(), id(r, "id"), tag(r), 1, true)
		reply(w, v, e)
	})
	mux.HandleFunc("POST /api/tables/{id}/records", func(w http.ResponseWriter, r *http.Request) {
		var in feetable.RecordInput
		if !decode(w, r, &in) {
			return
		}
		v, e := store.SaveRecord(r.Context(), id(r, "id"), 0, in)
		reply(w, v, e)
	})
	mux.HandleFunc("PUT /api/tables/{id}/records/{record}", func(w http.ResponseWriter, r *http.Request) {
		var in feetable.RecordInput
		if !decode(w, r, &in) {
			return
		}
		v, e := store.SaveRecord(r.Context(), id(r, "id"), id(r, "record"), in)
		reply(w, v, e)
	})
	mux.HandleFunc("DELETE /api/tables/{id}/records/{record}", func(w http.ResponseWriter, r *http.Request) {
		reply(w, map[string]bool{"ok": true}, store.DeleteRecord(r.Context(), id(r, "id"), id(r, "record"), revision(r)))
	})
	for _, kind := range []string{"locations", "tags"} {
		mux.HandleFunc("GET /api/"+kind, func(w http.ResponseWriter, r *http.Request) {
			v, e := store.Suggestions(r.Context(), kind)
			reply(w, v, e)
		})
		mux.HandleFunc("POST /api/"+kind, func(w http.ResponseWriter, r *http.Request) {
			var in struct {
				Name string `json:"name"`
			}
			if !decode(w, r, &in) {
				return
			}
			reply(w, map[string]bool{"ok": true}, store.AddSuggestion(r.Context(), kind, in.Name))
		})
	}
	mux.HandleFunc("PUT /api/locations/{id}", func(w http.ResponseWriter, r *http.Request) {
		var in feetable.LocationInput
		if !decode(w, r, &in) {
			return
		}
		reply(w, map[string]bool{"ok": true}, store.UpdateLocation(r.Context(), id(r, "id"), in))
	})
	mux.HandleFunc("DELETE /api/locations/{id}", func(w http.ResponseWriter, r *http.Request) {
		reply(w, map[string]bool{"ok": true}, store.DeleteLocation(r.Context(), id(r, "id"), r.URL.Query().Get("name")))
	})
	mux.HandleFunc("GET /api/tables/{id}/export", s.export)
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		respond(w, 404, map[string]any{"error": feetable.Error{Code: "NOT_FOUND", Message: "接口不存在"}})
	})
	mux.HandleFunc("/", s.static)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("handler panic", "error", recovered)
				reply(w, nil, fmt.Errorf("internal error"))
			}
		}()
		if r.Method != "GET" && r.Method != "HEAD" {
			if origin := r.Header.Get("Origin"); origin != "" {
				u, e := url.Parse(origin)
				scheme := "http"
				if r.TLS != nil {
					scheme = "https"
				}
				if e != nil || u.Host != r.Host || u.Scheme != scheme {
					respond(w, 403, map[string]any{"error": feetable.Error{Code: "ORIGIN", Message: "不允许跨站请求"}})
					return
				}
			}
		}
		mux.ServeHTTP(w, r)
	})
}
func id(r *http.Request, key string) int64 {
	v, _ := strconv.ParseInt(r.PathValue(key), 10, 64)
	return v
}
func revision(r *http.Request) int64 {
	v, _ := strconv.ParseInt(r.URL.Query().Get("revision"), 10, 64)
	return v
}
func page(r *http.Request) int {
	value := r.URL.Query().Get("page")
	if value == "" {
		return 1
	}
	v, e := strconv.Atoi(value)
	if e != nil || v < 1 || v > 1000000 {
		return 0
	}
	return v
}
func tag(r *http.Request) *string {
	v, ok := r.URL.Query()["tag"]
	if !ok || len(v) == 0 {
		return nil
	}
	return &v[0]
}
func decode(w http.ResponseWriter, r *http.Request, out any) bool {
	media, _, e := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if e != nil || media != "application/json" {
		reply(w, nil, &feetable.Error{Code: "VALIDATION", Message: "请求须为 JSON"})
		return false
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	var raw json.RawMessage
	if e = decoder.Decode(&raw); e != nil {
		reply(w, nil, &feetable.Error{Code: "VALIDATION", Message: "JSON 格式无效"})
		return false
	}
	var fields map[string]json.RawMessage
	if e = json.Unmarshal(raw, &fields); e != nil || fields == nil {
		reply(w, nil, &feetable.Error{Code: "VALIDATION", Message: "请求须为 JSON 对象"})
		return false
	}
	required := []string{}
	switch out.(type) {
	case *feetable.RecordInput:
		required = []string{"day", "location1", "location2", "quantity", "unitPrice", "amount", "tag", "revision"}
	case *feetable.TableInput:
		required = []string{"year", "month"}
	case *feetable.MergeInput:
		required = []string{"revision", "sources"}
	case *feetable.LocationInput:
		required = []string{"name", "previousName"}
	default:
		required = []string{"name"}
	}
	for _, key := range required {
		v, ok := fields[key]
		if !ok || (string(v) == "null" && key != "unitPrice" && key != "tag") {
			reply(w, nil, &feetable.Error{Code: "VALIDATION", Message: "缺少字段：" + key})
			return false
		}
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		reply(w, nil, &feetable.Error{Code: "VALIDATION", Message: "请求只能包含一个对象"})
		return false
	}
	decoder = json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if e = decoder.Decode(out); e != nil {
		reply(w, nil, &feetable.Error{Code: "VALIDATION", Message: "字段格式无效"})
		return false
	}
	return true
}
func respond(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func reply(w http.ResponseWriter, value any, e error) {
	if e == nil {
		respond(w, 200, value)
		return
	}
	var domain *feetable.Error
	status := 500
	if errors.As(e, &domain) {
		switch domain.Code {
		case "VALIDATION":
			status = 400
		case "NOT_FOUND":
			status = 404
		case "CONFLICT":
			status = 409
		case "LIMIT":
			status = 422
		}
	} else {
		slog.Error("request failed", "error", e)
		domain = &feetable.Error{Code: "INTERNAL", Message: "服务暂时不可用，请稍后重试"}
	}
	respond(w, status, map[string]any{"error": domain})
}
func (s *server) export(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format != "png" && format != "pdf" && format != "xlsx" {
		reply(w, nil, &feetable.Error{Code: "VALIDATION", Message: "导出格式无效"})
		return
	}
	select {
	case s.exportSlot <- struct{}{}:
		defer func() { <-s.exportSlot }()
	default:
		respond(w, 429, map[string]any{"error": feetable.Error{Code: "BUSY", Message: "正在生成另一份文件，请稍后重试"}})
		return
	}
	report, e := s.store.Report(r.Context(), id(r, "id"), tag(r), 1, true)
	if e != nil {
		reply(w, nil, e)
		return
	}
	if v := revision(r); v != 0 && v != report.Table.Revision {
		reply(w, nil, &feetable.Error{Code: "CONFLICT", Message: "表格已更新，请刷新预览后重新导出"})
		return
	}
	if report.Count == 0 {
		reply(w, nil, &feetable.Error{Code: "VALIDATION", Message: "没有记录可导出"})
		return
	}
	var data []byte
	mimeType := ""
	switch format {
	case "png":
		data, e = reportexport.PNG(report)
		mimeType = "image/png"
	case "pdf":
		data, e = reportexport.PDF(report)
		mimeType = "application/pdf"
	case "xlsx":
		data, e = reportexport.XLSX(report)
		mimeType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}
	if e != nil {
		reply(w, nil, e)
		return
	}
	suffix := ""
	if report.FilterTag != nil {
		suffix = "_" + strings.Map(func(c rune) rune {
			if unicode.IsLetter(c) || unicode.IsNumber(c) || c == '-' || c == '_' {
				return c
			}
			return '_'
		}, *report.FilterTag)
	}
	name := fmt.Sprintf("运费明细表_%d年%d月_%d%s.%s", report.Table.Year, report.Table.Month, report.Table.ID, suffix, format)
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", "no-store")
	disposition := "attachment"
	if r.URL.Query().Get("inline") == "1" {
		disposition = "inline"
	}
	w.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": name}))
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	_, _ = w.Write(data)
}
func (s *server) static(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if s.assets == nil {
		http.Error(w, "前端尚未构建", 503)
		return
	}
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if name == "" || name == "." {
		name = "index.html"
	}
	data, e := fs.ReadFile(s.assets, name)
	if e != nil {
		if strings.HasPrefix(name, "assets/") || path.Ext(name) != "" {
			http.NotFound(w, r)
			return
		}
		name = "index.html"
		data, e = fs.ReadFile(s.assets, name)
		if e != nil {
			http.NotFound(w, r)
			return
		}
	}
	w.Header().Set("Cache-Control", "no-cache")
	if strings.HasPrefix(name, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}
	if t := mime.TypeByExtension(path.Ext(name)); t != "" {
		w.Header().Set("Content-Type", t)
	}
	if r.Method != "HEAD" {
		_, _ = w.Write(data)
	}
}
