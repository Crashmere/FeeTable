package httpapi

import (
	"context"
	"encoding/json"
	"feetable/internal/feetable"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestHTTPContracts(t *testing.T) {
	s, e := feetable.Open(filepath.Join(t.TempDir(), "api.sqlite"), true)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	h := New(s, fstest.MapFS{"index.html": {Data: []byte("<html>/feetable/assets/main.js</html>")}, "assets/main.js": {Data: []byte("ok")}})
	call := func(method, path, body, origin string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	for _, tc := range []struct {
		method, path, body string
		want               int
	}{{"GET", "/healthz", "", 200}, {"GET", "/tables/1", "", 200}, {"GET", "/assets/missing.js", "", 404}, {"POST", "/api/tables", "null", 400}, {"POST", "/api/tables", `{"year":2026,"month":9,"extra":1}`, 400}, {"GET", "/api/tables?page=0", "", 400}} {
		if w := call(tc.method, tc.path, tc.body, ""); w.Code != tc.want {
			t.Fatal(tc, w.Code, w.Body.String())
		}
	}
	if w := call("POST", "/api/tables", `{"year":2026,"month":9}`, "https://evil.example"); w.Code != http.StatusForbidden {
		t.Fatal(w.Code)
	}
	w := call("POST", "/api/tables", `{"year":2026,"month":9}`, "")
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	var table feetable.Table
	json.Unmarshal(w.Body.Bytes(), &table)
	tagName := "甲/乙?#"
	price := "1"
	_, e = s.SaveRecord(context.Background(), table.ID, 0, feetable.RecordInput{Day: 1, Location1: "起点", Location2: "终点", Quantity: "0.145", UnitPrice: &price, Revision: 1, Tag: &tagName})
	if e != nil {
		t.Fatal(e)
	}
	w = call("GET", "/api/tables/1/export?format=xlsx&tag="+url.QueryEscape(tagName)+"&revision=2", "", "")
	if w.Code != 200 || !strings.Contains(w.Header().Get("Content-Disposition"), "attachment") {
		t.Fatal(w.Code, w.Body.String())
	}
	if w = call("GET", "/api/tables/1/export?format=png&revision=1", "", ""); w.Code != 409 {
		t.Fatal(w.Code)
	}
}
