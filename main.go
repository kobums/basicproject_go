package main

import (
	"os"
	"path"

	"basicproject_go/global/config"
	"basicproject_go/global/log"
	"basicproject_go/global/setting"
	"basicproject_go/models"
	"basicproject_go/services"
)

func main() {
	// config.Init 은 global/log 패키지 초기화 때 이미 불렸다 (.env → 환경변수 → .env.yml).
	log.Info().Str("Version", config.Version).Str("Mode", config.Mode).Msg("Start")

	if err := config.Validate(); err != nil {
		log.Error().Msg(err.Error())
		os.Exit(1)
	}

	if err := os.MkdirAll(path.Join(config.UploadPath, "temp"), 0o755); err != nil {
		log.Error().Err(err).Msg("can't create upload directory")
	}

	models.InitCache()
	setting.GetInstance()

	services.Http()
}
