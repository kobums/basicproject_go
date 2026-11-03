package jwt

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"basicproject_go/global/config"

	gojwt "github.com/golang-jwt/jwt/v5"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func setup(t *testing.T) {
	t.Helper()
	oldSecret, oldExp := config.JwtSecret, config.JwtExpiration
	config.JwtSecret = testSecret
	config.JwtExpiration = config.DefaultJwtExpiration
	t.Cleanup(func() { config.JwtSecret, config.JwtExpiration = oldSecret, oldExp })
}

func payload(t *testing.T, token string) map[string]interface{} {
	t.Helper()
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("token has %d parts", len(parts))
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	m := map[string]interface{}{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

// 발급한 토큰이 통일 JWT 형식({sub, id, email, level, iat, exp}, HS256)이고 다시 읽으면 같은 값인지 확인한다.
func TestRoundTrip(t *testing.T) {
	setup(t)
	now := time.Unix(1767225600, 0)
	token, err := NewToken(42, "a@b.c", 1, now)
	if err != nil {
		t.Fatal(err)
	}

	header, _ := base64.RawURLEncoding.DecodeString(strings.Split(token, ".")[0])
	if !strings.Contains(string(header), `"alg":"HS256"`) {
		t.Errorf("header = %s, want HS256", header)
	}

	got := payload(t, token)
	want := map[string]interface{}{
		"sub": "42", "id": float64(42), "email": "a@b.c", "level": float64(1),
		"iat": float64(1767225600), "exp": float64(1767225600 + 86400),
	}
	if len(got) != len(want) {
		t.Errorf("claims = %v, want keys of %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("claim %s = %#v, want %#v", k, got[k], v)
		}
	}

	// 서명 검증은 현재 시각 기준이라 지금 발급한 토큰으로 확인한다.
	token, _ = NewToken(42, "a@b.c", 3, time.Now())
	claims, err := Check("Bearer " + token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Id != 42 || claims.Email != "a@b.c" || claims.Level != 3 || claims.Subject != "42" {
		t.Errorf("claims = %+v", claims)
	}
	if ExpiresInMs() != 86400000 {
		t.Errorf("ExpiresInMs = %d", ExpiresInMs())
	}
}

// 다른 백엔드가 id 없이 sub 만 넣은 토큰도 읽는다.
func TestSubFallback(t *testing.T) {
	setup(t)
	now := time.Now()
	token, _ := gojwt.NewWithClaims(gojwt.SigningMethodHS256, gojwt.MapClaims{
		"sub": "7", "email": "x@y.z", "iat": now.Unix(), "exp": now.Add(time.Hour).Unix(),
	}).SignedString([]byte(testSecret))
	claims, err := Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Id != 7 || claims.Level != 0 {
		t.Errorf("claims = %+v", claims)
	}
}

func TestRejects(t *testing.T) {
	setup(t)
	now := time.Now()
	sign := func(method gojwt.SigningMethod, key interface{}, claims gojwt.MapClaims) string {
		s, err := gojwt.NewWithClaims(method, claims).SignedString(key)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	valid := gojwt.MapClaims{"sub": "1", "id": 1, "iat": now.Unix(), "exp": now.Add(time.Hour).Unix()}
	cases := map[string]string{
		"expired":      sign(gojwt.SigningMethodHS256, []byte(testSecret), gojwt.MapClaims{"sub": "1", "id": 1, "exp": now.Add(-time.Minute).Unix()}),
		"wrong secret": sign(gojwt.SigningMethodHS256, []byte("another-secret-another-secret-xx"), valid),
		"HS384":        sign(gojwt.SigningMethodHS384, []byte(testSecret), valid),
		"none":         sign(gojwt.SigningMethodNone, gojwt.UnsafeAllowNoneSignatureType, valid),
		"no user":      sign(gojwt.SigningMethodHS256, []byte(testSecret), gojwt.MapClaims{"exp": now.Add(time.Hour).Unix()}),
		"garbage":      "not.a.token",
	}
	for name, token := range cases {
		if _, err := Check("Bearer " + token); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if _, err := Check(sign(gojwt.SigningMethodHS256, []byte(testSecret), valid)); err == nil {
		t.Error("token without Bearer prefix accepted")
	}

	config.JwtSecret = "short"
	if _, err := NewToken(1, "a@b.c", 1, now); err == nil {
		t.Error("short secret accepted")
	}
}

func TestPassword(t *testing.T) {
	hash, err := GeneratePasswd("secret-password")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(hash, "$2a$10$") {
		t.Errorf("hash = %s, want bcrypt DefaultCost", hash)
	}
	if !CheckPasswd(hash, "secret-password") || CheckPasswd(hash, "wrong") {
		t.Error("CheckPasswd mismatch")
	}
}
