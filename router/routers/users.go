package routers

// Spring UserController / UserService 를 옮긴 사용자 라우트.
//   GET    /api/users        목록(keyword: 이름 검색) — 공개, 이메일 미노출
//   GET    /api/users/:id    단건 조회 — 공개, 이메일 미노출
//   POST   /api/users        생성 — 인증 필요
//   PUT    /api/users/:id    수정(본인만) — 인증 필요
//   DELETE /api/users/:id    삭제(본인만) — 인증 필요

import (
	"strconv"
	"strings"

	"basicproject_go/controllers/dto"
	"basicproject_go/global/jwt"
	"basicproject_go/models"

	"github.com/gofiber/fiber/v2"
)

type userCreateRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type userUpdateRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"` // 비우면 기존 비밀번호 유지
}

func SetupUsersRoutes(group fiber.Router) {
	group.Get("/users", handleUserList)
	group.Get("/users/:id", handleUserGet)
	group.Post("/users", JwtRequired, handleUserCreate)
	group.Put("/users/:id", JwtRequired, handleUserUpdate)
	group.Delete("/users/:id", JwtRequired, handleUserDelete)
}

func handleUserList(c *fiber.Ctx) error {
	page := c.QueryInt("page", 0)
	size := c.QueryInt("size", 10)
	if size <= 0 {
		size = 10
	}

	conn := models.NewConnection()
	defer conn.Close()
	manager := models.NewUserManager(conn)

	// 공개 엔드포인트이므로 이메일은 검색 대상에서 제외한다(이메일 열거 방지).
	var where []interface{}
	if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
		where = []interface{}{models.Where{Column: "name", Value: kw, Compare: "like"}}
	}

	total := manager.Count(where)

	args := append([]interface{}{}, where...)
	args = append(args, models.Paging(page+1, size), models.Ordering("id desc"))
	users := manager.Find(args)

	content := make([]dto.UserSummary, 0, len(users))
	for i := range users {
		content = append(content, dto.UserSummaryFrom(&users[i]))
	}
	return c.JSON(dto.NewPage(content, page, size, total))
}

func handleUserGet(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	conn := models.NewConnection()
	defer conn.Close()
	manager := models.NewUserManager(conn)

	user := manager.Get(id)
	if user == nil {
		return errorJSON(c, fiber.StatusNotFound, "user not found: "+c.Params("id"))
	}
	return c.JSON(dto.UserSummaryFrom(user))
}

func handleUserCreate(c *fiber.Ctx) error {
	var req userCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return errorJSON(c, fiber.StatusBadRequest, "잘못된 요청입니다.")
	}
	req.Email = strings.TrimSpace(req.Email)
	req.Name = strings.TrimSpace(req.Name)

	fields := map[string]string{}
	if req.Email == "" {
		fields["email"] = "이메일을 입력하세요."
	}
	if req.Password == "" {
		fields["password"] = "비밀번호를 입력하세요."
	}
	if len(fields) > 0 {
		return validationError(c, fields)
	}

	conn := models.NewConnection()
	defer conn.Close()
	manager := models.NewUserManager(conn)

	if manager.GetWhere([]interface{}{models.Where{Column: "email", Value: req.Email, Compare: "="}}) != nil {
		return errorJSON(c, fiber.StatusBadRequest, "이미 사용 중인 이메일입니다: "+req.Email)
	}

	hashed, err := jwt.GeneratePasswd(req.Password)
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "비밀번호 처리 중 오류가 발생했습니다.")
	}

	item := &models.UserUpdate{Email: req.Email, Name: req.Name, Passwd: hashed}
	if err := manager.Insert(item); err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "사용자 생성에 실패했습니다.")
	}
	item.Id = manager.GetIdentity()

	user := models.User{Id: item.Id, Email: item.Email, Name: item.Name, Date: item.Date}
	return c.Status(fiber.StatusCreated).JSON(dto.UserFrom(&user))
}

func handleUserUpdate(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	// 본인 계정만 수정할 수 있다.
	if id != currentUserID(c) {
		return errorJSON(c, fiber.StatusForbidden, "본인 계정만 수정/삭제할 수 있습니다.")
	}

	var req userUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return errorJSON(c, fiber.StatusBadRequest, "잘못된 요청입니다.")
	}

	conn := models.NewConnection()
	defer conn.Close()
	manager := models.NewUserManager(conn)

	user := manager.Get(id)
	if user == nil {
		return errorJSON(c, fiber.StatusNotFound, "user not found: "+c.Params("id"))
	}

	update := user.ConvertUpdate() // 기존 값 복사(비밀번호 포함)
	update.Name = strings.TrimSpace(req.Name)
	// 비밀번호를 새로 입력한 경우에만 해시해서 교체.
	if strings.TrimSpace(req.Password) != "" {
		if hashed, err := jwt.GeneratePasswd(req.Password); err == nil {
			update.Passwd = hashed
		}
	}
	if err := manager.Update(update); err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "사용자 수정에 실패했습니다.")
	}

	updated := models.User{Id: update.Id, Email: update.Email, Name: update.Name, Date: update.Date}
	return c.JSON(dto.UserFrom(&updated))
}

func handleUserDelete(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	if id != currentUserID(c) {
		return errorJSON(c, fiber.StatusForbidden, "본인 계정만 수정/삭제할 수 있습니다.")
	}

	conn := models.NewConnection()
	defer conn.Close()
	manager := models.NewUserManager(conn)

	if manager.Get(id) == nil {
		return errorJSON(c, fiber.StatusNotFound, "user not found: "+c.Params("id"))
	}
	if err := manager.Delete(id); err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "사용자 삭제에 실패했습니다.")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
