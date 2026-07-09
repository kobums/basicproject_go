package routers

// Spring 의 JwtAuthenticationFilter + ApiExceptionHandler + SecurityConfig 를 옮긴 공통 유틸.
// - JwtRequired: Authorization: Bearer <token> 를 검증하고 사용자 id/email 을 Locals 에 채운다.
// - errorJSON: Spring ErrorResponse 와 동일한 형태의 에러 응답을 낸다.

import (
	"strings"
	stdtime "time"

	"basicproject_go/global/jwt"

	"github.com/gofiber/fiber/v2"
)

const (
	localsUserID = "authUserID"
	localsEmail  = "authEmail"
)

// LIKE 절 원시 SQL 삽입을 막기 위해 따옴표/백슬래시/세미콜론을 제거한다.
var stringReplacer = strings.NewReplacer("'", "", "\"", "", "\\", "", ";", "")

// JwtRequired 는 인증이 필요한 라우트 앞에 붙는 미들웨어다.
// 토큰이 없거나 유효하지 않으면 401 을 반환한다.
func JwtRequired(c *fiber.Ctx) error {
	claims, err := jwt.Check(c.Get("Authorization"))
	if err != nil {
		return errorJSON(c, fiber.StatusUnauthorized, "인증이 필요합니다.")
	}
	c.Locals(localsUserID, claims.User.Id)
	c.Locals(localsEmail, claims.User.Email)
	return c.Next()
}

// currentUserID 는 JwtRequired 가 저장해 둔 인증 사용자 id 를 돌려준다.
func currentUserID(c *fiber.Ctx) int64 {
	if v, ok := c.Locals(localsUserID).(int64); ok {
		return v
	}
	return 0
}

// errorJSON 은 Spring ApiExceptionHandler.ErrorResponse 와 동일한 JSON 을 반환한다.
func errorJSON(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"status":      status,
		"message":     message,
		"fieldErrors": fiber.Map{},
		"timestamp":   stdtime.Now().Format("2006-01-02T15:04:05"),
	})
}

// validationError 는 필드 검증 실패(400) 응답. 프론트가 fieldErrors 를 읽어 폼에 표시한다.
func validationError(c *fiber.Ctx, fields map[string]string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"status":      fiber.StatusBadRequest,
		"message":     "validation failed",
		"fieldErrors": fields,
		"timestamp":   stdtime.Now().Format("2006-01-02T15:04:05"),
	})
}

// sanitizeKeyword 는 Custom(원시 SQL) LIKE 절에 넣기 전에 위험 문자를 제거한다.
func sanitizeKeyword(kw string) string {
	r := stringReplacer
	return r.Replace(kw)
}
