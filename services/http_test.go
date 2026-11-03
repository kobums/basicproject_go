package services

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"basicproject_go/global/config"
	"basicproject_go/global/jwt"

	"github.com/gofiber/fiber/v2"
)

const testOrigin = "http://localhost:5173"

// newTestApp 는 DB 없이 앱을 만든다. DB 풀은 첫 쿼리 때 열리고, 닫힌 포트를 가리키므로 쿼리는 바로 실패한다.
func newTestApp(t *testing.T) *fiber.App {
	t.Helper()
	config.JwtSecret = "0123456789abcdef0123456789abcdef"
	config.JwtExpiration = config.DefaultJwtExpiration
	config.Database.TypeString = "mysql"
	config.Database.ConnectionString = "u:p@tcp(127.0.0.1:1)/x?timeout=1s"
	config.Cors = []string{testOrigin}
	config.Log.Web = false
	return NewApp()
}

func do(t *testing.T, app *fiber.App, method, url, body string, headers map[string]string) (int, map[string]interface{}, http.Header) {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, url, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := app.Test(req, 10000)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	m := map[string]interface{}{}
	_ = json.Unmarshal(raw, &m)
	return resp.StatusCode, m, resp.Header
}

func TestAuthRoutes(t *testing.T) {
	app := newTestApp(t)

	status, body, _ := do(t, app, "GET", "/api/health", "", nil)
	if status != 200 || body["status"] != "ok" {
		t.Errorf("health = %d %v", status, body)
	}

	// 회원가입 라우트가 있고, 검증 실패는 DB 없이 400 이다.
	status, body, _ = do(t, app, "POST", "/api/auth/signup", `{"email":"bad","password":"x"}`, nil)
	if status != 400 || body["message"] != "validation failed" {
		t.Errorf("signup validation = %d %v", status, body)
	}
	fields, _ := body["fieldErrors"].(map[string]interface{})
	if fields["email"] == nil || fields["password"] == nil {
		t.Errorf("signup fieldErrors = %v", body["fieldErrors"])
	}

	status, _, _ = do(t, app, "POST", "/api/auth/login", `{}`, nil)
	if status != 400 {
		t.Errorf("login without credentials = %d", status)
	}

	status, _, _ = do(t, app, "GET", "/api/auth/me", "", nil)
	if status != 401 {
		t.Errorf("me without token = %d, want 401", status)
	}

	status, _, _ = do(t, app, "POST", "/api/upload/index", "", nil)
	if status != 401 {
		t.Errorf("upload without token = %d, want 401", status)
	}
}

// 생성 라우트(GET /api/board)는 토큰이 없거나 무효면 auth error, 유효하면 세션을 받아 DB 로 간다.
func TestGeneratedRoutesSession(t *testing.T) {
	app := newTestApp(t)
	token, err := jwt.NewToken(3, "a@b.c", 1, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	expired, _ := jwt.NewToken(3, "a@b.c", 1, time.Now().Add(-48*time.Hour))

	for _, tc := range []struct {
		name, auth, want string
	}{
		{"no token", "", "auth error"},
		{"garbage token", "Bearer abc.def.ghi", "auth error"},
		{"expired token", "Bearer " + expired, "auth error"},
		{"valid token", "Bearer " + token, "error"}, // 세션 통과, DB 없음
	} {
		headers := map[string]string{}
		if tc.auth != "" {
			headers["Authorization"] = tc.auth
		}
		status, body, _ := do(t, app, "GET", "/api/board", "", headers)
		if status != 200 || body["code"] != tc.want {
			t.Errorf("%s: GET /api/board = %d %v, want code %q", tc.name, status, body, tc.want)
		}
	}

	status, body, _ := do(t, app, "GET", "/api/board?page=x", "", map[string]string{"Authorization": "Bearer " + token})
	if status != 400 || body["code"] != "bad request" {
		t.Errorf("bad page = %d %v", status, body)
	}
}

func TestCorsPreflight(t *testing.T) {
	app := newTestApp(t)
	for _, method := range []string{"GET", "PATCH", "DELETE"} {
		status, _, h := do(t, app, "OPTIONS", "/api/board/1", "", map[string]string{
			"Origin":                         testOrigin,
			"Access-Control-Request-Method":  method,
			"Access-Control-Request-Headers": "authorization,content-type",
		})
		if status != 204 {
			t.Errorf("%s preflight status = %d", method, status)
		}
		if h.Get("Access-Control-Allow-Origin") != testOrigin {
			t.Errorf("allow-origin = %q", h.Get("Access-Control-Allow-Origin"))
		}
		if !strings.Contains(strings.ToLower(h.Get("Access-Control-Allow-Headers")), "authorization") {
			t.Errorf("allow-headers = %q, want Authorization", h.Get("Access-Control-Allow-Headers"))
		}
		if !strings.Contains(h.Get("Access-Control-Allow-Methods"), method) {
			t.Errorf("allow-methods = %q, want %s", h.Get("Access-Control-Allow-Methods"), method)
		}
	}

	// 허용하지 않은 출처에는 CORS 헤더를 주지 않는다.
	_, _, h := do(t, app, "OPTIONS", "/api/board", "", map[string]string{
		"Origin":                        "http://evil.example",
		"Access-Control-Request-Method": "GET",
	})
	if v := h.Get("Access-Control-Allow-Origin"); v != "" {
		t.Errorf("unknown origin allowed: %q", v)
	}
}

func TestBodyLimit(t *testing.T) {
	if BodyLimit != 20*1024*1024 {
		t.Errorf("BodyLimit = %d", BodyLimit)
	}
}
