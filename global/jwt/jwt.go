package jwt

// 통일 JWT (Go / Spring / Rust 템플릿이 같은 토큰을 주고받는다).
//   - 알고리즘 HS256 고정, 비밀키 JWT_SECRET (32 바이트 이상, config.Validate 가 확인한다)
//   - 헤더 Authorization: Bearer <token>
//   - claims {"sub": "42", "id": 42, "email": "a@b.c", "level": 1, "iat": 1767225600, "exp": 1767312000}
//     sub 는 id 문자열, iat / exp 는 초 단위 정수. 읽을 때 id 가 없으면 sub 를 숫자로 쓴다.

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"basicproject_go/global/config"
	"basicproject_go/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Claims 는 통일 JWT 의 payload 다.
type Claims struct {
	Id    int64  `json:"id"`
	Email string `json:"email"`
	Level int    `json:"level"`
	jwt.RegisteredClaims
}

var (
	ErrNoToken     = errors.New("authorization header must be 'Bearer <token>'")
	ErrInvalidUser = errors.New("token has no user id")
)

// ExpiresInMs 는 발급 토큰의 유효 시간(밀리초)이다. 로그인 응답의 expiresInMs 로 내려간다.
func ExpiresInMs() int64 {
	return expiration().Milliseconds()
}

func expiration() time.Duration {
	if config.JwtExpiration > 0 {
		return config.JwtExpiration
	}
	return config.DefaultJwtExpiration
}

func secret() ([]byte, error) {
	if len(config.JwtSecret) < config.MinJwtSecretLength {
		return nil, fmt.Errorf("JWT_SECRET must be at least %d bytes", config.MinJwtSecretLength)
	}
	return []byte(config.JwtSecret), nil
}

// NewToken 은 사용자 정보로 HS256 토큰을 만든다.
func NewToken(id int64, email string, level int, now time.Time) (string, error) {
	key, err := secret()
	if err != nil {
		return "", err
	}
	claims := Claims{
		Id:    id,
		Email: email,
		Level: level,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(id, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration())),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
}

// MakeToken 은 로그인 / 회원가입한 사용자의 토큰을 만든다.
func MakeToken(user models.User) (string, error) {
	return NewToken(user.Id, user.Email, user.Level, time.Now())
}

// Parse 는 토큰 문자열을 검증하고 claims 를 돌려준다 (서명, HS256, exp).
func Parse(token string) (*Claims, error) {
	key, err := secret()
	if err != nil {
		return nil, err
	}
	var claims Claims
	_, err = jwt.ParseWithClaims(token, &claims, func(*jwt.Token) (interface{}, error) {
		return key, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	if claims.Id == 0 && claims.Subject != "" {
		claims.Id, _ = strconv.ParseInt(claims.Subject, 10, 64)
	}
	if claims.Id <= 0 {
		return nil, ErrInvalidUser
	}
	return &claims, nil
}

// Check 는 Authorization 헤더 값("Bearer <token>")을 검증한다.
func Check(header string) (*Claims, error) {
	scheme, token, ok := strings.Cut(strings.TrimSpace(header), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
		return nil, ErrNoToken
	}
	return Parse(strings.TrimSpace(token))
}

func CheckPasswd(dbPasswd string, inputPasswd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(dbPasswd), []byte(inputPasswd))
	return err == nil
}

// GeneratePasswd 는 bcrypt(DefaultCost) 해시를 만든다. 생성 코드(비밀번호 컬럼)도 이 함수를 쓴다.
func GeneratePasswd(passwd string) (string, error) {
	hashedPasswd, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPasswd), nil
}
