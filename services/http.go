package services

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"basicproject_go/global/config"
	"basicproject_go/global/log"
	"basicproject_go/models"
	"basicproject_go/router"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// BodyLimit 은 요청 본문 최대 크기(20MB)다. 이미지 자체 상한은 global/upload.MaxImageSize (5MB).
// 더 큰 요청이 필요하면 엔드포인트별로 조정한다.
const BodyLimit = 20 * 1024 * 1024

// CorsAllowHeaders 는 브라우저가 보낼 수 있는 요청 헤더다 (Authorization: Bearer 토큰 포함).
const CorsAllowHeaders = "Origin,Content-Type,Accept,Content-Length,Accept-Language,Accept-Encoding,Authorization"

// CorsAllowMethods 는 생성 라우트(GET/POST/PUT/PATCH/DELETE)와 preflight 를 허용한다.
const CorsAllowMethods = "GET,POST,HEAD,PUT,PATCH,DELETE,OPTIONS"

// NewApp 은 미들웨어와 라우트를 등록한 fiber 앱을 만든다 (DB 접속은 첫 요청 때 한다).
func NewApp() *fiber.App {
	app := fiber.New(fiber.Config{
		BodyLimit:             BodyLimit,
		Prefork:               false,
		CaseSensitive:         true,
		StrictRouting:         true,
		DisableStartupMessage: true,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
	})

	if len(config.Cors) > 0 {
		origins := strings.Join(config.Cors, ",")
		app.Use(cors.New(cors.Config{
			AllowOrigins: origins,
			// "*" 와 credentials 는 같이 쓸 수 없다 (fiber 가 panic). 토큰은 헤더로 보내므로 쿠키가 필요 없다.
			AllowCredentials: !strings.Contains(origins, "*"),
			AllowMethods:     CorsAllowMethods,
			AllowHeaders:     CorsAllowHeaders,
		}))
	}

	if config.Log.Web {
		app.Use(fiberzerolog.New(fiberzerolog.Config{
			Logger: log.Get(),
		}))
	}

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestCompression,
	}))

	// 업로드 파일 정적 서빙. 로그 디렉터리(webdata/log)는 SQL 파라미터 등 민감정보가, certs 는 TLS 개인키가
	// 들어갈 수 있으므로 제외한다.
	app.Static("/webdata", config.UploadPath, fiber.Static{
		Next: isHiddenStatic,
	})

	// 업로드 이미지 정적 서빙: /uploads/{파일명} → {UploadPath}/uploads/{파일명} (global/upload.StoreImage)
	app.Static("/uploads", config.UploadPath+"/uploads")

	router.SetRouter(app)

	// SPA 빌드 결과(DocumentRoot/index.html)를 나머지 경로에 돌려준다.
	app.Get("/*", func(ctx *fiber.Ctx) error {
		return ctx.SendFile(fmt.Sprintf("./%v/index.html", config.DocumentRoot), true)
	})

	return app
}

func Http() {
	log.Info().Str("service", "HTTP").Str("port", config.Port).Msg("Start Service")

	app := NewApp()
	defer models.ClosePool()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-c
		log.Info().Msg("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	if config.Mode == "develop" || !config.Tls.Use {
		if err := app.Listen(":" + config.Port); err != nil {
			log.Error().Msg(err.Error())
		}
	} else {
		cer, err := tls.LoadX509KeyPair(config.Tls.Cert, config.Tls.Key)
		if err != nil {
			log.Error().Msg("TLS error")
			log.Error().Msg(err.Error())
			return
		}

		cert := &tls.Config{Certificates: []tls.Certificate{cer}}

		ln, err := tls.Listen("tcp", ":"+config.Port, cert)
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		if err := app.Listener(ln); err != nil {
			log.Error().Msg(err.Error())
		}
	}
}

// isHiddenStatic 은 /webdata 정적 서빙에서 숨길 경로인지 판단한다.
// c.Path() 는 퍼센트 인코딩이 풀리지 않은 원본 경로라 "/webdata/%6cog/..." 처럼 우회할 수 있으므로,
// 실제 파일 조회에 쓰이는 fasthttp 의 디코딩·정규화된 경로로 판단한다. (대소문자 무시 파일시스템 대비 소문자 비교)
func isHiddenStatic(c *fiber.Ctx) bool {
	p := strings.ToLower(path.Clean("/" + string(c.Context().Path())))
	for _, dir := range []string{"/webdata/log", "/webdata/certs"} {
		if p == dir || strings.HasPrefix(p, dir+"/") {
			return true
		}
	}
	return false
}
