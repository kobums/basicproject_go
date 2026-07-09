package router

import (
	"basicproject_go/router/routers"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func getArrayCommal(name string) []int64 {
	values := strings.Split(name, ",")

	var items []int64
	for _, item := range values {
        n, _ := strconv.ParseInt(item, 10, 64)
		items = append(items, n)
	}

	return items
}

func getArrayCommai(name string) []int {
	values := strings.Split(name, ",")

	var items []int
	for _, item := range values {
        n, _ := strconv.Atoi(item)
		items = append(items, n)
	}

	return items
}

func SetRouter(r *fiber.App) {
	apiGroup := r.Group("/api")

	// Spring 에서 이식한 도메인 라우트.
	// 인증(JwtRequired)은 각 라우트에서 개별로 건다: 조회(GET)는 공개,
	// 생성·수정·삭제와 /auth/me 는 인증 필요.
	routers.SetupAuthRoutes(apiGroup)
	routers.SetupBoardsRoutes(apiGroup)
	routers.SetupUsersRoutes(apiGroup)
	routers.SetupUploadRoutes(apiGroup)
}