package api

import (
	"basicproject_go/controllers"
	"basicproject_go/global/config"
	"net/http"
	"os"
	"path"
)

type UploadController struct {
	controllers.Controller
}

// @POST()
func (c *UploadController) Index() {
	name := c.Context.FormValue("name", "upload")

	dirName := c.Now.DateAsOnlyNumber()
	fullPath := path.Join(config.UploadPath, dirName)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		err = os.Mkdir(fullPath, os.ModePerm)
		if err != nil {
			c.Error(err)
			return
		}
	}

	// 이미지(jpg/jpeg/png/gif/webp, 5MB 이하)만 허용한다. 확장자 화이트리스트 + 내용 시그니처 검사,
	// 저장 파일명은 서버가 생성한다 (global/upload.ValidateImage).
	originalFilename, filename, err := c.GetUpload(dirName, name)
	if err != nil {
		c.Context.Status(http.StatusBadRequest)
		c.Error(err)
		return
	}

	c.Set("filename", filename)
	c.Set("original", originalFilename)
}