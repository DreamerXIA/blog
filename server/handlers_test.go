package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

const testToken = "test-token"

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Fatal("缺少 TEST_DATABASE_URL 环境变量（测试用 Postgres 连接串）")
	}
	store, err := Open(url)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	if err := store.reset(); err != nil {
		t.Fatalf("reset store: %v", err)
	}
	s := &server{store: store, token: testToken}
	ts := httptest.NewServer(s.routes())
	t.Cleanup(ts.Close)
	return ts
}

func request(t *testing.T, ts *httptest.Server, method, path, token string, body any) *http.Response {
	t.Helper()
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, ts.URL+path, rd)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("do request %s %s: %v", method, path, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func decode[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	var v T
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return v
}

func TestHealth(t *testing.T) {
	ts := newTestServer(t)
	resp := request(t, ts, "GET", "/api/health", "", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	got := decode[map[string]string](t, resp)
	if got["status"] != "ok" {
		t.Fatalf("status field = %q, want ok", got["status"])
	}
}

func TestProfileGetSeeded(t *testing.T) {
	ts := newTestServer(t)
	resp := request(t, ts, "GET", "/api/profile", "", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	p := decode[Profile](t, resp)
	if p.Name == "" {
		t.Fatalf("seeded profile name is empty")
	}
}

func TestProfileUpdateRequiresToken(t *testing.T) {
	ts := newTestServer(t)
	body := Profile{Name: "张三", Email: "a@b.c"}
	resp := request(t, ts, "PUT", "/api/profile", "", body)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
	resp = request(t, ts, "PUT", "/api/profile", "wrong-token", body)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestProfileUpdateAndRead(t *testing.T) {
	ts := newTestServer(t)
	body := Profile{
		Name:          "李四",
		Phone:         "13800000000",
		Email:         "lisi@example.com",
		TechDirection: "后端",
		LearningGoals: "掌握 Go",
	}
	resp := request(t, ts, "PUT", "/api/profile", testToken, body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	got := decode[Profile](t, resp)
	if got.Name != "李四" || got.Email != "lisi@example.com" {
		t.Fatalf("unexpected profile: %+v", got)
	}

	// 读取应返回更新后的数据
	resp = request(t, ts, "GET", "/api/profile", "", nil)
	got = decode[Profile](t, resp)
	if got.TechDirection != "后端" {
		t.Fatalf("tech_direction = %q, want 后端", got.TechDirection)
	}
}

func TestProfileUpdateRequiresName(t *testing.T) {
	ts := newTestServer(t)
	body := Profile{Name: "  "}
	resp := request(t, ts, "PUT", "/api/profile", testToken, body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestCreateLogRequiresToken(t *testing.T) {
	ts := newTestServer(t)
	body := map[string]string{"type": "work", "title": "t", "content": "c"}
	resp := request(t, ts, "POST", "/api/logs", "", body)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestCreateLogValidation(t *testing.T) {
	ts := newTestServer(t)
	cases := []map[string]string{
		{"type": "work", "title": "", "content": "c"},      // 缺标题
		{"type": "work", "title": "t", "content": ""},      // 缺正文
		{"type": "invalid", "title": "t", "content": "c"},  // 类型非法
		{"type": "", "title": "t", "content": "c"},         // 类型为空
	}
	for _, c := range cases {
		resp := request(t, ts, "POST", "/api/logs", testToken, c)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("case %+v: status = %d, want 400", c, resp.StatusCode)
		}
	}
}

func TestCreateAndListLog(t *testing.T) {
	ts := newTestServer(t)
	body := map[string]string{"type": "work", "title": "写接口", "content": "完成了日志 API"}
	resp := request(t, ts, "POST", "/api/logs", testToken, body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	created := decode[Log](t, resp)
	if created.ID == 0 || created.Title != "写接口" {
		t.Fatalf("unexpected created log: %+v", created)
	}
	if created.CreatedAt == "" || created.UpdatedAt == "" {
		t.Fatalf("timestamps missing: %+v", created)
	}

	// 跨请求读取，验证持久化
	resp = request(t, ts, "GET", "/api/logs", "", nil)
	logs := decode[[]Log](t, resp)
	if len(logs) != 1 {
		t.Fatalf("list length = %d, want 1", len(logs))
	}
	if logs[0].Title != "写接口" {
		t.Fatalf("list title = %q, want 写接口", logs[0].Title)
	}
}

func TestListLogsDescOrder(t *testing.T) {
	ts := newTestServer(t)
	for _, title := range []string{"第一条", "第二条", "第三条"} {
		body := map[string]string{"type": "study", "title": title, "content": "内容"}
		resp := request(t, ts, "POST", "/api/logs", testToken, body)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("create %q: status = %d", title, resp.StatusCode)
		}
	}
	resp := request(t, ts, "GET", "/api/logs", "", nil)
	logs := decode[[]Log](t, resp)
	if len(logs) != 3 {
		t.Fatalf("length = %d, want 3", len(logs))
	}
	if logs[0].Title != "第三条" || logs[2].Title != "第一条" {
		t.Fatalf("order wrong: %q, %q", logs[0].Title, logs[2].Title)
	}
}

func TestListLogsTypeFilter(t *testing.T) {
	ts := newTestServer(t)
	create := func(logType, title string) {
		body := map[string]string{"type": logType, "title": title, "content": "c"}
		resp := request(t, ts, "POST", "/api/logs", testToken, body)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("create %q: status = %d", title, resp.StatusCode)
		}
	}
	create("work", "工作一")
	create("study", "学习一")
	create("daily", "日报一")

	resp := request(t, ts, "GET", "/api/logs?type=work", "", nil)
	logs := decode[[]Log](t, resp)
	if len(logs) != 1 || logs[0].Type != "work" {
		t.Fatalf("work filter wrong: %+v", logs)
	}

	resp = request(t, ts, "GET", "/api/logs?type=study", "", nil)
	logs = decode[[]Log](t, resp)
	if len(logs) != 1 || logs[0].Title != "学习一" {
		t.Fatalf("study filter wrong: %+v", logs)
	}
}

func TestListLogsInvalidType(t *testing.T) {
	ts := newTestServer(t)
	resp := request(t, ts, "GET", "/api/logs?type=bogus", "", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestGetLogByID(t *testing.T) {
	ts := newTestServer(t)
	body := map[string]string{"type": "summary", "title": "阶段总结", "content": "本阶段完成"}
	resp := request(t, ts, "POST", "/api/logs", testToken, body)
	created := decode[Log](t, resp)

	resp = request(t, ts, "GET", "/api/logs/"+strconv.FormatInt(created.ID, 10), "", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	got := decode[Log](t, resp)
	if got.Title != "阶段总结" {
		t.Fatalf("title = %q, want 阶段总结", got.Title)
	}
}

func TestGetLogNotFound(t *testing.T) {
	ts := newTestServer(t)
	resp := request(t, ts, "GET", "/api/logs/999", "", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

