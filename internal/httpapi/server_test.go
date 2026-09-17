package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"feetable/internal/feetable"
	"fmt"
	"image/png"
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

func TestManagementHTTPContracts(t *testing.T) {
	s, err := feetable.Open(filepath.Join(t.TempDir(), "manage.sqlite"), true)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	h := New(s, fstest.MapFS{})
	call := func(method, path, body string, want int) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != want {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
		}
		return w
	}
	for _, query := range []string{"sort=bad", "year=2024", "year=oops&month=2", "year=0&month=0", "year=2024&month=13"} {
		call("GET", "/api/tables?"+query, "", 400)
	}
	for i := 0; i < 2; i++ {
		call("POST", "/api/tables", `{"year":2024,"month":2}`, 200)
		call("POST", fmt.Sprintf("/api/tables/%d/records", i+1), `{"day":29,"location1":"起点","location2":"终点","quantity":"10.5","unitPrice":"2.88","amount":"0","tag":null,"revision":1}`, 200)
	}
	w := call("GET", "/api/tables?sort=month_desc&year=2024&month=2", "", 200)
	var list feetable.TableList
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil || list.Total != 2 {
		t.Fatal(list, err)
	}
	for _, body := range []string{`{}`, `{"revision":2}`, `{"revision":2,"sources":null}`, `{"revision":2,"sources":[{"id":2}]}`, `{"revision":2,"sources":[{"id":2,"revision":2,"extra":1}]}`} {
		call("POST", "/api/tables/1/merge", body, 400)
	}
	call("POST", "/api/tables/1/merge", `{"revision":2,"sources":[{"id":2,"revision":1}]}`, 409)
	call("POST", "/api/tables/1/merge", `{"revision":2,"sources":[{"id":2,"revision":2}]}`, 200)
	call("GET", "/api/tables/2", "", 404)
	w = call("GET", "/api/tables/1/report", "", 200)
	var report feetable.Report
	if err := json.Unmarshal(w.Body.Bytes(), &report); err != nil || report.Total != "60.48" || report.Count != 2 {
		t.Fatal(report, err)
	}
	call("GET", "/api/tables/1/export?format=png&revision=2", "", 409)
	w = call("GET", "/api/tables/1/export?format=png&revision=3", "", 200)
	config, err := png.DecodeConfig(bytes.NewReader(w.Body.Bytes()))
	if err != nil || config.Width != 2832 {
		t.Fatal(config, err)
	}
	w = call("GET", "/api/locations", "", 200)
	var locations []feetable.Suggestion
	if err := json.Unmarshal(w.Body.Bytes(), &locations); err != nil {
		t.Fatal(err)
	}
	location := locations[0]
	path := fmt.Sprintf("/api/locations/%d", location.ID)
	call("PUT", path, `{"name":"新地点"}`, 400)
	body, _ := json.Marshal(feetable.LocationInput{Name: "新地点 / #?", PreviousName: location.Name})
	call("PUT", path, string(body), 200)
	call("DELETE", path+"?name="+url.QueryEscape(location.Name), "", 409)
	call("DELETE", path+"?name="+url.QueryEscape("新地点 / #?"), "", 200)
	w = call("GET", "/api/tables/1/report", "", 200)
	if !bytes.Contains(w.Body.Bytes(), []byte("起点")) || !bytes.Contains(w.Body.Bytes(), []byte("终点")) {
		t.Fatal("location edits changed report")
	}
}
