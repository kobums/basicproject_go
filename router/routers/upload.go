package routers

import (
	"basicproject_go/controllers/api"

	"github.com/gofiber/fiber/v2"
)

// SetupUploadRoutes 는 이미지 업로드 라우트를 등록한다.
//
//	POST /api/upload/index  (multipart "upload" 필드, 인증 필요)
func SetupUploadRoutes(group fiber.Router) {
	// 인증된 사용자만 업로드할 수 있다 (비인증 임의 파일 업로드 / 호스팅 방지).
	group.Post("/upload/index", JwtRequired, func(c *fiber.Ctx) error {
		var controller api.UploadController
		controller.Init(c)
		controller.Index()
		controller.Close()
		return c.JSON(controller.Result)
	})
}
