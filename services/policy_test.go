package services

// 생성 라우트의 권한 규칙(model.json: board read "login" + session, User 본인 / 관리자 읽기, protect level)과
// 인증 / 업로드 API 계약(gomachine README "인증 API 계약")을 standin DB 드라이버로 확인한다.

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"basicproject_go/global/jwt"

	"github.com/gofiber/fiber/v2"
)

const testDate = "2026-01-02 03:04:05"

func bearer(t *testing.T, id int64, level int) map[string]string {
	t.Helper()
	token, err := jwt.NewToken(id, "u@x.io", level, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return map[string]string{"Authorization": "Bearer " + token}
}

// userRow 는 user_tb select 한 행이다 (u_id, u_email, u_passwd, u_name, u_level, u_createddate).
func userRow(id int64, email, hash string, level int64) []driver.Value {
	return []driver.Value{id, email, hash, "name", level, testDate}
}

// boardRow 는 board_tb 한 행이다. join 이면 작성자(user_tb) 컬럼이 뒤에 붙는다 (Find / Read 는 left join, GetRaw 는 아니다).
func boardRow(id, owner int64, join bool) []driver.Value {
	row := []driver.Value{id, "title", "content", nil, owner, testDate}
	if join {
		row = append(row, userRow(owner, "o@x.io", "hash", 1)...)
	}
	return row
}

// boardRows 는 board_tb 조회에 owner 가 쓴 글 하나씩을 돌려준다 (id 는 1 부터).
func boardRows(owners ...int64) func(string, []driver.Value) ([][]driver.Value, error) {
	return func(q string, _ []driver.Value) ([][]driver.Value, error) {
		if !strings.Contains(q, "from board_tb") {
			return nil, nil
		}
		if strings.Contains(q, "count(") {
			return [][]driver.Value{{int64(len(owners))}}, nil
		}
		join := strings.Contains(q, "left join")
		var rows [][]driver.Value
		for i, o := range owners {
			rows = append(rows, boardRow(int64(i+1), o, join))
		}
		return rows, nil
	}
}

func execCalls() []standinCall {
	var out []standinCall
	for _, c := range standinCalls() {
		if c.Exec {
			out = append(out, c)
		}
	}
	return out
}

func keys(m map[string]interface{}) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

var userKeys = []string{"createddate", "email", "id", "level", "name"}

// board: read "login" — 비로그인은 auth error, 로그인 사용자는 남의 글도 읽고, 수정 / 삭제는 작성자만.
func TestBoardReadLoginPolicy(t *testing.T) {
	app := newTestApp(t)

	for _, url := range []string{"/api/board", "/api/board/1", "/api/board/count", "/api/board/search/title?title=a"} {
		standinReset(boardRows(8))
		status, body, _ := do(t, app, "GET", url, "", nil)
		if status != 200 || body["code"] != "auth error" {
			t.Errorf("anonymous GET %s = %d %v, want auth error", url, status, body)
		}
		if n := len(standinCalls()); n != 0 {
			t.Errorf("anonymous GET %s reached DB (%d calls)", url, n)
		}
	}

	// 로그인한 7 번 사용자: 8, 9 번이 쓴 글 목록을 읽는다 (작성자 조건 없음, 작성자 join).
	me := bearer(t, 7, 1)
	standinReset(boardRows(8, 9))
	status, body, _ := do(t, app, "GET", "/api/board", "", me)
	items, _ := body["items"].([]interface{})
	if status != 200 || body["code"] != "ok" || len(items) != 2 {
		t.Fatalf("GET /api/board as non-owner = %d %v, want 2 items", status, body)
	}
	if q := standinFind("from board_tb"); q == nil || strings.Contains(q.Query, "b_user = ?") {
		t.Errorf("board list query has owner filter: %+v", q)
	}
	first, _ := items[0].(map[string]interface{})
	extra, _ := first["extra"].(map[string]interface{})
	if author, _ := extra["user"].(map[string]interface{}); author == nil || author["id"] != float64(8) || author["passwd"] != nil {
		t.Errorf("board item author join = %v", first["extra"])
	}

	standinReset(boardRows(8))
	status, body, _ = do(t, app, "GET", "/api/board/1", "", me)
	if item, _ := body["item"].(map[string]interface{}); status != 200 || body["code"] != "ok" || item["id"] != float64(1) {
		t.Errorf("GET /api/board/1 (other's row) = %d %v, want item", status, body)
	}

	// 남의 글 수정 / 삭제는 auth error 이고 쓰기 SQL 이 나가지 않는다.
	for _, tc := range []struct{ method, url, body string }{
		{"PUT", "/api/board/1", `{"title":"hacked"}`},
		{"PATCH", "/api/board/1", `{"title":"hacked"}`},
		{"PUT", "/api/board", `{"id":1,"title":"hacked"}`},
		{"DELETE", "/api/board/1", ""},
		{"DELETE", "/api/board/batch", `[{"id":1}]`},
	} {
		standinReset(boardRows(8))
		status, body, _ := do(t, app, tc.method, tc.url, tc.body, me)
		if status != 200 || body["code"] != "auth error" {
			t.Errorf("%s %s on other's row = %d %v, want auth error", tc.method, tc.url, status, body)
		}
		if ex := execCalls(); len(ex) != 0 {
			t.Errorf("%s %s on other's row wrote: %+v", tc.method, tc.url, ex)
		}
	}

	// 작성자 본인은 수정 / 삭제할 수 있다.
	owner := bearer(t, 8, 1)
	standinReset(boardRows(8))
	status, body, _ = do(t, app, "PUT", "/api/board/1", `{"title":"mine"}`, owner)
	if status != 200 || body["code"] != "ok" || len(execCalls()) != 1 || !strings.HasPrefix(execCalls()[0].Query, "update board_tb") {
		t.Errorf("owner PUT = %d %v, writes %+v", status, body, execCalls())
	}
	standinReset(boardRows(8))
	status, body, _ = do(t, app, "DELETE", "/api/board/1", "", owner)
	if status != 200 || body["code"] != "ok" || len(execCalls()) != 1 || !strings.HasPrefix(execCalls()[0].Query, "delete from board_tb") {
		t.Errorf("owner DELETE = %d %v, writes %+v", status, body, execCalls())
	}

	// 새 글의 작성자는 세션 사용자다 (본문의 user 는 무시).
	standinReset(boardRows())
	status, body, _ = do(t, app, "POST", "/api/board", `{"title":"t","content":"c","user":99}`, me)
	ex := execCalls()
	if status != 200 || body["code"] != "ok" || len(ex) != 1 || !containsValue(ex[0].Args, int64(7)) || containsValue(ex[0].Args, int64(99)) {
		t.Errorf("POST /api/board = %d %v, insert %+v, want user 7", status, body, ex)
	}
}

func containsValue(args []driver.Value, v driver.Value) bool {
	for _, a := range args {
		if reflect.DeepEqual(a, v) {
			return true
		}
	}
	return false
}

// User: 관리자가 아니면 본인 행만 읽는다 (목록에 u_id = 세션 id 조건, 남의 단건은 auth error). 관리자(level 3)는 모든 행.
func TestUserSelfOrAdminRead(t *testing.T) {
	app := newTestApp(t)
	users := func(q string, _ []driver.Value) ([][]driver.Value, error) {
		if !strings.Contains(q, "from user_tb") {
			return nil, nil
		}
		if strings.Contains(q, "count(") {
			return [][]driver.Value{{int64(1)}}, nil
		}
		return [][]driver.Value{userRow(7, "me@x.io", "hash", 1)}, nil
	}

	standinReset(users)
	status, body, _ := do(t, app, "GET", "/api/user", "", nil)
	if status != 200 || body["code"] != "auth error" || len(standinCalls()) != 0 {
		t.Errorf("anonymous GET /api/user = %d %v", status, body)
	}

	standinReset(users)
	status, body, _ = do(t, app, "GET", "/api/user", "", bearer(t, 7, 1))
	items, _ := body["items"].([]interface{})
	if status != 200 || body["code"] != "ok" || len(items) != 1 {
		t.Fatalf("GET /api/user as user = %d %v", status, body)
	}
	q := standinFind("from user_tb")
	if q == nil || !strings.Contains(q.Query, "u_id = ") || !containsValue(q.Args, int64(7)) {
		t.Errorf("non-admin user list query has no self filter: %+v", q)
	}
	if item, _ := items[0].(map[string]interface{}); item["passwd"] != nil || item["id"] != float64(7) {
		t.Errorf("user item = %v", item)
	}

	// 같은 조건이 개수 / 검색에도 붙는다.
	for _, url := range []string{"/api/user/count", "/api/user/search/name?name=x"} {
		standinReset(users)
		do(t, app, "GET", url, "", bearer(t, 7, 1))
		if q := standinFind("from user_tb"); q == nil || !strings.Contains(q.Query, "u_id = ") || !containsValue(q.Args, int64(7)) {
			t.Errorf("GET %s query has no self filter: %+v", url, q)
		}
	}

	standinReset(users)
	status, body, _ = do(t, app, "GET", "/api/user/8", "", bearer(t, 7, 1))
	if status != 200 || body["code"] != "auth error" || len(standinCalls()) != 0 {
		t.Errorf("GET /api/user/8 as user 7 = %d %v, want auth error before DB", status, body)
	}

	standinReset(users)
	status, body, _ = do(t, app, "GET", "/api/user", "", bearer(t, 1, 3))
	if status != 200 || body["code"] != "ok" {
		t.Errorf("GET /api/user as admin = %d %v", status, body)
	}
	if q := standinFind("from user_tb"); q == nil || strings.Contains(q.Query, "u_id = ") {
		t.Errorf("admin user list query has self filter: %+v", q)
	}
}

// 생성된 POST /api/user 도 level 을 클라이언트 값 대신 DB 기본값 1 로 넣는다 (protect level).
func TestGeneratedUserInsertLevel(t *testing.T) {
	app := newTestApp(t)
	standinReset(func(string, []driver.Value) ([][]driver.Value, error) { return nil, nil })
	status, body, _ := do(t, app, "POST", "/api/user", `{"email":"a@x.io","passwd":"password1","name":"a","level":9}`, nil)
	ex := execCalls()
	if status != 200 || body["code"] != "ok" || len(ex) != 1 {
		t.Fatalf("POST /api/user = %d %v, writes %+v", status, body, ex)
	}
	// insert into user_tb (u_email, u_passwd, u_name, u_level, u_createddate)
	if len(ex[0].Args) != 5 || ex[0].Args[3] != int64(1) {
		t.Errorf("insert args = %v, want level 1", ex[0].Args)
	}
}

func TestSignupContract(t *testing.T) {
	app := newTestApp(t)
	// email 조회(값 = email 문자열)는 없음, id 조회는 저장된 행.
	standinReset(func(q string, args []driver.Value) ([][]driver.Value, error) {
		if containsValue(args, int64(standinInsertID)) {
			return [][]driver.Value{userRow(standinInsertID, "new@x.io", "hash", 1)}, nil
		}
		return nil, nil
	})

	status, body, _ := do(t, app, "POST", "/api/auth/signup", `{"email":"new@x.io","password":"password1","name":"New","level":9}`, nil)
	if status != 201 {
		t.Fatalf("signup = %d %v, want 201", status, body)
	}
	checkTokenResponse(t, "signup", body, standinInsertID, 1)

	ex := execCalls()
	if len(ex) != 1 || !strings.HasPrefix(ex[0].Query, "insert into user_tb") || ex[0].Args[3] != int64(1) {
		t.Errorf("signup insert = %+v, want level 1", ex)
	}
	if hash, _ := ex[0].Args[1].(string); hash == "password1" || !jwt.CheckPasswd(hash, "password1") {
		t.Errorf("signup password not hashed: %v", ex[0].Args[1])
	}

	// 이미 있는 email 은 409.
	standinReset(func(string, []driver.Value) ([][]driver.Value, error) {
		return [][]driver.Value{userRow(5, "new@x.io", "hash", 1)}, nil
	})
	status, body, _ = do(t, app, "POST", "/api/auth/signup", `{"email":"new@x.io","password":"password1","name":"New"}`, nil)
	if status != 409 || body["status"] != float64(409) || body["message"] == nil || len(execCalls()) != 0 {
		t.Errorf("duplicate signup = %d %v", status, body)
	}

	// 검증 실패는 400 + fieldErrors.
	status, body, _ = do(t, app, "POST", "/api/auth/signup", `{"email":"new@x.io","password":"short"}`, nil)
	fields, _ := body["fieldErrors"].(map[string]interface{})
	if status != 400 || body["status"] != float64(400) || fields["password"] == nil {
		t.Errorf("short password signup = %d %v", status, body)
	}
}

func TestLoginAndMeContract(t *testing.T) {
	app := newTestApp(t)
	hash, err := jwt.GeneratePasswd("password1")
	if err != nil {
		t.Fatal(err)
	}
	standinReset(func(string, []driver.Value) ([][]driver.Value, error) {
		return [][]driver.Value{userRow(42, "a@x.io", hash, 2)}, nil
	})

	status, body, _ := do(t, app, "POST", "/api/auth/login", `{"email":"a@x.io","password":"password1"}`, nil)
	if status != 200 {
		t.Fatalf("login = %d %v", status, body)
	}
	checkTokenResponse(t, "login", body, 42, 2)

	status, body, _ = do(t, app, "POST", "/api/auth/login", `{"email":"a@x.io","password":"wrong-password"}`, nil)
	if status != 401 || body["status"] != float64(401) || body["token"] != nil {
		t.Errorf("wrong password login = %d %v", status, body)
	}

	status, body, _ = do(t, app, "GET", "/api/auth/me", "", bearer(t, 42, 2))
	if status != 200 || !reflect.DeepEqual(keys(body), userKeys) || body["id"] != float64(42) || body["createddate"] != testDate {
		t.Errorf("me = %d %v, want user %v", status, body, userKeys)
	}

	// 없는 사용자 email 은 401.
	standinReset(func(string, []driver.Value) ([][]driver.Value, error) { return nil, nil })
	status, body, _ = do(t, app, "POST", "/api/auth/login", `{"email":"none@x.io","password":"password1"}`, nil)
	if status != 401 || body["status"] != float64(401) {
		t.Errorf("unknown email login = %d %v", status, body)
	}

	// 토큰은 유효하지만 회원이 없으면 401.
	status, body, _ = do(t, app, "GET", "/api/auth/me", "", bearer(t, 42, 2))
	if status != 401 || body["status"] != float64(401) {
		t.Errorf("me for deleted user = %d %v", status, body)
	}
}

// checkTokenResponse 는 {token, tokenType: "Bearer", expiresInMs, user} 와 토큰 claim 을 확인한다.
func checkTokenResponse(t *testing.T, name string, body map[string]interface{}, id int64, level int) {
	t.Helper()
	if got := keys(body); !reflect.DeepEqual(got, []string{"expiresInMs", "token", "tokenType", "user"}) {
		t.Errorf("%s keys = %v", name, got)
	}
	if body["tokenType"] != "Bearer" || body["expiresInMs"] != float64(86400000) {
		t.Errorf("%s tokenType / expiresInMs = %v / %v", name, body["tokenType"], body["expiresInMs"])
	}
	user, _ := body["user"].(map[string]interface{})
	if !reflect.DeepEqual(keys(user), userKeys) || user["id"] != float64(id) || user["level"] != float64(level) || user["createddate"] != testDate {
		t.Errorf("%s user = %v", name, user)
	}
	token, _ := body["token"].(string)
	claims, err := jwt.Check("Bearer " + token)
	if err != nil || claims.Id != id || claims.Level != level {
		t.Errorf("%s token claims = %+v, %v", name, claims, err)
	}
}

func multipartBody(t *testing.T, field, filename string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	if field != "" {
		part, err := w.CreateFormFile(field, filename)
		if err != nil {
			t.Fatal(err)
		}
		part.Write(content)
	} else {
		w.WriteField("x", "y")
	}
	w.Close()
	return &buf, w.FormDataContentType()
}

func upload(t *testing.T, app *fiber.App, field, filename string, content []byte) (int, map[string]interface{}) {
	t.Helper()
	buf, ctype := multipartBody(t, field, filename, content)
	req := httptest.NewRequest("POST", "/api/upload/index", buf)
	req.Header.Set("Content-Type", ctype)
	for k, v := range bearer(t, 7, 1) {
		req.Header.Set(k, v)
	}
	resp, err := app.Test(req, 10000)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	m := map[string]interface{}{}
	_ = json.Unmarshal(raw, &m)
	return resp.StatusCode, m
}

func TestUploadContract(t *testing.T) {
	app := newTestApp(t)
	var img bytes.Buffer
	if err := png.Encode(&img, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}

	status, body := upload(t, app, "upload", "a.png", img.Bytes())
	filename, _ := body["filename"].(string)
	if status != 200 || !reflect.DeepEqual(keys(body), []string{"filename", "url"}) || !strings.HasSuffix(filename, ".png") || body["url"] != "/uploads/"+filename {
		t.Fatalf("upload = %d %v, want {filename, url}", status, body)
	}

	// 돌려준 url 로 정적 파일을 읽는다.
	resp, err := app.Test(httptest.NewRequest("GET", "/uploads/"+filename, nil), 10000)
	if err != nil || resp.StatusCode != 200 {
		t.Errorf("GET uploaded url = %v %v", resp, err)
	} else if raw, _ := io.ReadAll(resp.Body); !bytes.Equal(raw, img.Bytes()) {
		t.Errorf("uploaded content differs")
	}

	status, body = upload(t, app, "upload", "a.png", []byte("<html>not an image</html>"))
	if status != 400 || body["status"] != float64(400) || body["message"] == nil {
		t.Errorf("non-image upload = %d %v", status, body)
	}

	status, body = upload(t, app, "", "", nil)
	fields, _ := body["fieldErrors"].(map[string]interface{})
	if status != 400 || fields["upload"] == nil {
		t.Errorf("upload without file = %d %v", status, body)
	}
}
