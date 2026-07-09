package routers

// Spring AuthController / AuthService 를 옮긴 인증 라우트.
//   POST /api/auth/signup  회원가입 후 토큰 발급
//   POST /api/auth/login   로그인(평문 비밀번호는 BCrypt 로 자동 마이그레이션)
//   GET  /api/auth/me      현재 토큰의 사용자 정보(인증 필요)

import (
	"strings"

	"basicproject_go/controllers/dto"
	"basicproject_go/global/jwt"
	"basicproject_go/models"

	"github.com/gofiber/fiber/v2"
)

type signupRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func SetupAuthRoutes(group fiber.Router) {
	group.Post("/auth/signup", handleSignup)
	group.Post("/auth/login", handleLogin)
	group.Get("/auth/me", JwtRequired, handleMe)
}

func handleSignup(c *fiber.Ctx) error {
	var req signupRequest
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

	// 이메일 중복 검사 (UserService.create 재사용에 대응).
	if manager.GetWhere([]interface{}{models.Where{Column: "email", Value: req.Email, Compare: "="}}) != nil {
		return errorJSON(c, fiber.StatusBadRequest, "이미 사용 중인 이메일입니다: "+req.Email)
	}

	hashed, err := jwt.GeneratePasswd(req.Password)
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "비밀번호 처리 중 오류가 발생했습니다.")
	}

	item := &models.UserUpdate{Email: req.Email, Name: req.Name, Passwd: hashed}
	if err := manager.Insert(item); err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "회원가입에 실패했습니다.")
	}
	item.Id = manager.GetIdentity()

	user := models.User{Id: item.Id, Email: item.Email, Name: item.Name, Date: item.Date}
	token := jwt.MakeToken(user)
	return c.Status(fiber.StatusCreated).JSON(tokenResponse(token, user))
}

func handleLogin(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return errorJSON(c, fiber.StatusBadRequest, "잘못된 요청입니다.")
	}
	req.Email = strings.TrimSpace(req.Email)

	conn := models.NewConnection()
	defer conn.Close()
	manager := models.NewUserManager(conn)

	user := manager.GetWhere([]interface{}{models.Where{Column: "email", Value: req.Email, Compare: "="}})
	if user == nil || !verifyPassword(manager, user, req.Password) {
		return errorJSON(c, fiber.StatusBadRequest, "이메일 또는 비밀번호가 올바르지 않습니다.")
	}

	token := jwt.MakeToken(*user)
	return c.JSON(tokenResponse(token, *user))
}

func handleMe(c *fiber.Ctx) error {
	conn := models.NewConnection()
	defer conn.Close()
	manager := models.NewUserManager(conn)

	user := manager.Get(currentUserID(c))
	if user == nil {
		return errorJSON(c, fiber.StatusNotFound, "user not found")
	}
	return c.JSON(dto.UserFrom(user))
}

// verifyPassword: 저장된 값이 BCrypt 해시면 그대로 검증한다.
// 학습용으로 평문 저장된 기존 데이터는 평문 비교 후 일치하면 BCrypt 로 자동 마이그레이션한다.
// (Spring AuthService.matches 대응)
func verifyPassword(manager *models.UserManager, user *models.User, raw string) bool {
	stored := user.Passwd
	if strings.HasPrefix(stored, "$2") {
		return jwt.CheckPasswd(stored, raw)
	}
	if stored != "" && stored == raw {
		if hashed, err := jwt.GeneratePasswd(raw); err == nil {
			update := user.ConvertUpdate()
			update.Passwd = hashed
			if err := manager.Update(update); err == nil {
				user.Passwd = hashed
			}
		}
		return true
	}
	return false
}

func tokenResponse(token string, user models.User) dto.TokenResponse {
	return dto.TokenResponse{
		Token:       token,
		TokenType:   "Bearer",
		ExpiresInMs: jwt.ExpiresInMs,
		User:        dto.UserFrom(&user),
	}
}
