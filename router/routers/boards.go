package routers

// Spring BoardController / BoardService 를 옮긴 게시판 라우트.
//   GET    /api/boards          목록(type: title|content|all|author, keyword 검색) — 공개
//   GET    /api/boards/:id      단건 조회 — 공개
//   POST   /api/boards          작성(multipart, 이미지 선택) — 인증 필요
//   PUT    /api/boards/:id      수정(작성자 본인만) — 인증 필요
//   DELETE /api/boards/:id      삭제(작성자 본인만) — 인증 필요

import (
	"strconv"
	"strings"

	"basicproject_go/controllers/dto"
	"basicproject_go/global/upload"
	"basicproject_go/models"

	"github.com/gofiber/fiber/v2"
)

func SetupBoardsRoutes(group fiber.Router) {
	group.Get("/boards", handleBoardList)
	group.Get("/boards/:id", handleBoardGet)
	group.Post("/boards", JwtRequired, handleBoardCreate)
	group.Put("/boards/:id", JwtRequired, handleBoardUpdate)
	group.Delete("/boards/:id", JwtRequired, handleBoardDelete)
}

func handleBoardList(c *fiber.Ctx) error {
	page := c.QueryInt("page", 0)
	size := c.QueryInt("size", 10)
	if size <= 0 {
		size = 10
	}

	conn := models.NewConnection()
	defer conn.Close()
	manager := models.NewBoardManager(conn)

	// 검색 조건(Where/Custom)만 담는다. Count 와 Find 에 공통으로 쓴다.
	where, ok := boardSearchArgs(conn, c.Query("type"), c.Query("keyword"))
	if !ok {
		// author 검색인데 일치하는 작성자가 없음 → 빈 목록.
		return c.JSON(dto.NewPage([]dto.BoardResponse{}, page, size, 0))
	}

	total := manager.Count(where)

	args := append([]interface{}{}, where...)
	args = append(args, models.Paging(page+1, size), models.Ordering("id desc"))
	boards := manager.Find(args)

	authors := loadAuthors(conn, boards)
	content := make([]dto.BoardResponse, 0, len(boards))
	for i := range boards {
		b := boards[i]
		content = append(content, dto.BoardFrom(&b, dto.Author{Id: b.User, Name: authors[b.User]}))
	}
	return c.JSON(dto.NewPage(content, page, size, total))
}

func handleBoardGet(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	conn := models.NewConnection()
	defer conn.Close()
	manager := models.NewBoardManager(conn)

	board := manager.Get(id)
	if board == nil {
		return errorJSON(c, fiber.StatusNotFound, "board not found: "+c.Params("id"))
	}
	authors := loadAuthors(conn, []models.Board{*board})
	return c.JSON(dto.BoardFrom(board, dto.Author{Id: board.User, Name: authors[board.User]}))
}

func handleBoardCreate(c *fiber.Ctx) error {
	title := strings.TrimSpace(c.FormValue("title"))
	content := c.FormValue("content")
	if title == "" {
		return validationError(c, map[string]string{"title": "제목을 입력하세요."})
	}

	// 이미지는 선택값: 없으면 빈 문자열로 저장.
	filename, err := upload.StoreImage(c, "image")
	if err != nil {
		return errorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	conn := models.NewConnection()
	defer conn.Close()
	manager := models.NewBoardManager(conn)

	userID := currentUserID(c)
	board := &models.Board{Title: title, Content: content, Img: filename, User: userID}
	if err := manager.Insert(board); err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "게시글 저장에 실패했습니다.")
	}
	board.Id = manager.GetIdentity()

	authors := loadAuthors(conn, []models.Board{*board})
	return c.Status(fiber.StatusCreated).
		JSON(dto.BoardFrom(board, dto.Author{Id: userID, Name: authors[userID]}))
}

func handleBoardUpdate(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	conn := models.NewConnection()
	defer conn.Close()
	manager := models.NewBoardManager(conn)

	board := manager.Get(id)
	if board == nil {
		return errorJSON(c, fiber.StatusNotFound, "board not found: "+c.Params("id"))
	}
	if board.User != currentUserID(c) {
		return errorJSON(c, fiber.StatusForbidden, "본인이 작성한 게시글만 수정/삭제할 수 있습니다.")
	}

	title := strings.TrimSpace(c.FormValue("title"))
	if title == "" {
		return validationError(c, map[string]string{"title": "제목을 입력하세요."})
	}
	board.Title = title
	board.Content = c.FormValue("content")

	// 새 이미지가 있으면 교체하고 이전 파일은 삭제.
	filename, err := upload.StoreImage(c, "image")
	if err != nil {
		return errorJSON(c, fiber.StatusBadRequest, err.Error())
	}
	if filename != "" {
		old := board.Img
		board.Img = filename
		defer upload.DeleteQuietly(old)
	}

	if err := manager.Update(board); err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "게시글 수정에 실패했습니다.")
	}

	authors := loadAuthors(conn, []models.Board{*board})
	return c.JSON(dto.BoardFrom(board, dto.Author{Id: board.User, Name: authors[board.User]}))
}

func handleBoardDelete(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	conn := models.NewConnection()
	defer conn.Close()
	manager := models.NewBoardManager(conn)

	board := manager.Get(id)
	if board == nil {
		return errorJSON(c, fiber.StatusNotFound, "board not found: "+c.Params("id"))
	}
	if board.User != currentUserID(c) {
		return errorJSON(c, fiber.StatusForbidden, "본인이 작성한 게시글만 수정/삭제할 수 있습니다.")
	}

	if err := manager.Delete(id); err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "게시글 삭제에 실패했습니다.")
	}
	upload.DeleteQuietly(board.Img)
	return c.SendStatus(fiber.StatusNoContent)
}

// boardSearchArgs: type/keyword 로 Where/Custom 검색 인자를 만든다.
// 두 번째 반환값이 false 면 author 검색인데 일치 작성자가 없어 결과가 확정적으로 비어 있다는 뜻.
func boardSearchArgs(conn *models.Connection, searchType, keyword string) ([]interface{}, bool) {
	kw := strings.TrimSpace(keyword)
	if kw == "" {
		return nil, true
	}

	switch searchType {
	case "content":
		return []interface{}{models.Where{Column: "content", Value: kw, Compare: "like"}}, true
	case "all":
		like := "%" + kw + "%"
		return []interface{}{models.Custom{
			Query:  "(b_title like ? or b_content like ?)",
			Params: []interface{}{like, like},
		}}, true
	case "author":
		ids := userIDsByName(conn, kw)
		if len(ids) == 0 {
			return nil, false
		}
		return []interface{}{models.Where{Column: "user", Value: ids, Compare: "in"}}, true
	default: // "title"
		return []interface{}{models.Where{Column: "title", Value: kw, Compare: "like"}}, true
	}
}

// userIDsByName: 이름 부분일치로 작성자 id 목록을 조회한다(작성자 검색용).
func userIDsByName(conn *models.Connection, name string) []int64 {
	users := models.NewUserManager(conn).
		Find([]interface{}{models.Where{Column: "name", Value: name, Compare: "like"}})
	ids := make([]int64, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.Id)
	}
	return ids
}

// loadAuthors: 게시글 목록의 작성자 id → 이름 맵을 한 번의 IN 쿼리로 만든다(N+1 방지).
func loadAuthors(conn *models.Connection, boards []models.Board) map[int64]string {
	names := map[int64]string{}
	if len(boards) == 0 {
		return names
	}
	ids := make([]int64, 0, len(boards))
	seen := map[int64]bool{}
	for _, b := range boards {
		if !seen[b.User] {
			seen[b.User] = true
			ids = append(ids, b.User)
		}
	}
	users := models.NewUserManager(conn).
		Find([]interface{}{models.Where{Column: "id", Value: ids, Compare: "in"}})
	for _, u := range users {
		names[u.Id] = u.Name
	}
	return names
}
