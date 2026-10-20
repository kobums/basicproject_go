package routers

import (
	"basicproject_go/controllers/api"

	"github.com/gofiber/fiber/v2"
)

// SetupUploadRoutes sets up routes for projects domain
func SetupUploadRoutes(group fiber.Router) {

	// 인증된 사용자만 업로드 가능 (비인증 임의 파일 업로드/호스팅 방지)
	group.Post("/upload/index", JwtRequired, func(c *fiber.Ctx) error {
		var controller api.UploadController
		controller.Init(c)
		controller.Index()
		controller.Close()
		return c.JSON(controller.Result)
	})

}