package router

import (
	"basicproject_go/router/routers"

	"github.com/gofiber/fiber/v2"
)

// SetRouter 는 /api 아래에 라우트를 등록한다.
//   - 손으로 관리하는 파일: auth.go (로그인 / 회원가입 / me), upload.go, health
//   - gomachine 생성 파일: {table}.go + generated_routes.go (make gen 으로 다시 만든다)
func SetRouter(r *fiber.App) {
	apiGroup := r.Group("/api")

	apiGroup.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	routers.SetupAuthRoutes(apiGroup)
	routers.SetupUploadRoutes(apiGroup)
	routers.SetupGeneratedRoutes(apiGroup)
}
