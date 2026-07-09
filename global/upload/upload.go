package upload

// Spring FileStorageService 를 옮긴 이미지 저장 유틸.
// 업로드 파일은 {UploadPath}/uploads 아래에 저장하고, /uploads/{파일명} 으로 정적 서빙한다.

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"

	"basicproject_go/global"
	"basicproject_go/global/config"

	"github.com/gofiber/fiber/v2"
)

// 정적 서빙되는 경로이므로 이미지 외 파일(.html, .svg 등)로 인한 저장형 XSS 를 막기 위해 화이트리스트로 제한.
var allowedExt = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
var allowedType = map[string]bool{"image/jpeg": true, "image/png": true, "image/gif": true, "image/webp": true}

// Root 는 업로드 파일이 저장되는 실제 디렉터리.
func Root() string {
	return path.Join(config.UploadPath, "uploads")
}

// StoreImage 는 multipart 필드의 이미지를 저장하고 저장 파일명을 돌려준다.
// 파일이 없으면 ("", nil) 을 반환한다(이미지는 선택값).
func StoreImage(c *fiber.Ctx, field string) (string, error) {
	file, err := c.FormFile(field)
	if err != nil || file == nil {
		return "", nil
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExt[ext] {
		return "", errors.New("허용되지 않는 파일 형식입니다. (jpg, jpeg, png, gif, webp 만 가능)")
	}
	ct := strings.ToLower(file.Header.Get("Content-Type"))
	if !allowedType[ct] {
		return "", errors.New("이미지 파일만 업로드할 수 있습니다.")
	}

	dir := Root()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	filename := strings.ReplaceAll(global.UUID(), "-", "") + ext
	if err := c.SaveFile(file, path.Join(dir, filename)); err != nil {
		return "", err
	}
	return filename, nil
}

// DeleteQuietly 는 업로드된 파일만 삭제한다(외부 URL/경로는 건드리지 않음).
func DeleteQuietly(img string) {
	if img == "" || strings.HasPrefix(img, "http://") ||
		strings.HasPrefix(img, "https://") || strings.HasPrefix(img, "/") {
		return
	}
	_ = os.Remove(path.Join(Root(), img))
}
