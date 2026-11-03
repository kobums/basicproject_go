package routers

// 인증 / 업로드 같은 손으로 만든 라우트의 공통 유틸 (생성 라우트는 controllers.Controller.Session 을 쓴다).
// - JwtRequired: Authorization: Bearer <token> 를 검증하고 세션을 Locals 에 채운다.
// - errorJSON: 인증 API 계약의 에러 본문 {status, message, fieldErrors?} 을 낸다 (Spring / Rust 템플릿과 같다).

import (
	"basicproject_go/controllers"

	"github.com/gofiber/fiber/v2"
)

const localsSession = "session"

// JwtRequired 는 인증이 필요한 라우트 앞에 붙는 미들웨어다.
// 토큰이 없거나 유효하지 않으면 401 을 돌려준다.
func JwtRequired(c *fiber.Ctx) error {
	session := controllers.SessionFrom(c)
	if session == nil {
		return errorJSON(c, fiber.StatusUnauthorized, "인증이 필요합니다.")
	}
	c.Locals(localsSession, session)
	return c.Next()
}

// currentSession 은 JwtRequired 가 저장해 둔 세션을 돌려준다 (없으면 nil).
func currentSession(c *fiber.Ctx) *controllers.Session {
	session, _ := c.Locals(localsSession).(*controllers.Session)
	return session
}

// ErrorResponse 는 인증 / 업로드 API 의 에러 본문이다. status 는 HTTP 상태 코드와 같다.
// fieldErrors 는 검증 실패(400)일 때만 {필드: 메시지} 로 들어간다.
type ErrorResponse struct {
	Status      int               `json:"status"`
	Message     string            `json:"message"`
	FieldErrors map[string]string `json:"fieldErrors,omitempty"`
}

// errorJSON 은 HTTP status 와 {status, message} 본문을 돌려준다.
func errorJSON(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(ErrorResponse{Status: status, Message: message})
}

// validationError 는 필드 검증 실패(400) 응답이다. 프론트가 fieldErrors 를 읽어 폼에 표시한다.
func validationError(c *fiber.Ctx, fields map[string]string) error {
	return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
		Status:      fiber.StatusBadRequest,
		Message:     "validation failed",
		FieldErrors: fields,
	})
}
