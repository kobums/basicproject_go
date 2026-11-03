package routers

import (
	"errors"
	"strings"

	"basicproject_go/global/log"
	"basicproject_go/global/upload"

	"github.com/gofiber/fiber/v2"
)

// UploadField 는 업로드 파일의 multipart field 이름이다 (gomachine README "인증 API 계약").
const UploadField = "upload"

// UploadResponse 는 업로드 성공 응답이다. url 은 정적 서빙 경로 /uploads/{filename} 이다.
type UploadResponse struct {
	Filename string `json:"filename"`
	URL      string `json:"url"`
}

// SetupUploadRoutes 는 이미지 업로드 라우트를 등록한다.
//
//	POST /api/upload/index  multipart field "upload" (인증 필요) → 200 {filename, url} | 400 | 401
//
// 이미지(jpg/jpeg/png/gif/webp, 5MB 이하)만 받는다. 저장 이름은 서버가 만든다 (global/upload.StoreImage).
func SetupUploadRoutes(group fiber.Router) {
	// 인증된 사용자만 업로드할 수 있다 (비인증 임의 파일 업로드 / 호스팅 방지).
	group.Post("/upload/index", JwtRequired, handleUpload)
}

func handleUpload(c *fiber.Ctx) error {
	filename, err := upload.StoreImage(c, UploadField)
	if errors.Is(err, upload.ErrInvalidImage) {
		return errorJSON(c, fiber.StatusBadRequest, strings.TrimPrefix(err.Error(), upload.ErrInvalidImage.Error()+": "))
	}
	if err != nil {
		log.Error().Err(err).Msg("upload store failed")
		return errorJSON(c, fiber.StatusInternalServerError, "파일 저장에 실패했습니다.")
	}
	if filename == "" {
		return validationError(c, map[string]string{UploadField: "업로드할 파일이 없습니다."})
	}
	return c.JSON(UploadResponse{Filename: filename, URL: "/uploads/" + filename})
}
