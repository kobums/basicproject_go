package routers

// 인증 라우트 (손으로 관리하는 파일, gomachine 이 덮어쓰지 않는다).
//   POST /api/auth/signup  회원가입 (level 1) 후 토큰 발급
//   POST /api/auth/login   로그인 후 토큰 발급
//   GET  /api/auth/me      현재 토큰의 사용자 정보 (인증 필요)
// 토큰은 통일 JWT (global/jwt) 이고, 생성 라우트는 같은 토큰을 controllers.Controller.Session 으로 읽는다.

import (
	"net/mail"
	"strings"

	"basicproject_go/controllers/dto"
	"basicproject_go/global/jwt"
	"basicproject_go/global/log"
	"basicproject_go/models"

	"github.com/gofiber/fiber/v2"
)

// 회원가입 기본 등급. 관리자 등급은 model.json 의 adminLevel (기본 3) 이다.
const signupLevel = 1

// bcrypt 는 72 바이트까지만 쓴다.
const maxPasswordLength = 72

const minPasswordLength = 8

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
	} else if _, err := mail.ParseAddress(req.Email); err != nil {
		fields["email"] = "이메일 형식이 올바르지 않습니다."
	}
	if len(req.Password) < minPasswordLength || len(req.Password) > maxPasswordLength {
		fields["password"] = "비밀번호는 8~72자로 입력하세요."
	}
	if len(fields) > 0 {
		return validationError(c, fields)
	}
	if req.Name == "" {
		req.Name = strings.SplitN(req.Email, "@", 2)[0]
	}

	conn := models.NewConnection()
	if conn == nil {
		return errorJSON(c, fiber.StatusServiceUnavailable, "데이터베이스에 연결할 수 없습니다.")
	}
	defer conn.Close()
	manager := models.NewUserManager(conn)

	if manager.GetWhere([]interface{}{models.Where{Column: "email", Value: req.Email, Compare: "="}}) != nil {
		return errorJSON(c, fiber.StatusConflict, "이미 사용 중인 이메일입니다.")
	}

	hashed, err := jwt.GeneratePasswd(req.Password)
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "비밀번호 처리 중 오류가 발생했습니다.")
	}

	item := &models.UserUpdate{Email: req.Email, Name: req.Name, Passwd: hashed, Level: signupLevel}
	if err := manager.Insert(item); err != nil {
		log.Error().Err(err).Msg("signup insert failed")
		return errorJSON(c, fiber.StatusInternalServerError, "회원가입에 실패했습니다.")
	}
	item.Id = manager.GetIdentity()

	user := models.User{Id: item.Id, Email: item.Email, Name: item.Name, Level: item.Level, Createddate: item.Createddate}
	return sendToken(c, fiber.StatusCreated, user)
}

func handleLogin(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return errorJSON(c, fiber.StatusBadRequest, "잘못된 요청입니다.")
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		return errorJSON(c, fiber.StatusBadRequest, "이메일 또는 비밀번호가 올바르지 않습니다.")
	}

	conn := models.NewConnection()
	if conn == nil {
		return errorJSON(c, fiber.StatusServiceUnavailable, "데이터베이스에 연결할 수 없습니다.")
	}
	defer conn.Close()
	manager := models.NewUserManager(conn)

	user := manager.GetWhere([]interface{}{models.Where{Column: "email", Value: req.Email, Compare: "="}})
	if user == nil || !jwt.CheckPasswd(user.Passwd, req.Password) {
		return errorJSON(c, fiber.StatusUnauthorized, "이메일 또는 비밀번호가 올바르지 않습니다.")
	}

	return sendToken(c, fiber.StatusOK, *user)
}

func handleMe(c *fiber.Ctx) error {
	session := currentSession(c)
	if session == nil {
		return errorJSON(c, fiber.StatusUnauthorized, "인증이 필요합니다.")
	}

	conn := models.NewConnection()
	if conn == nil {
		return errorJSON(c, fiber.StatusServiceUnavailable, "데이터베이스에 연결할 수 없습니다.")
	}
	defer conn.Close()
	manager := models.NewUserManager(conn)

	user := manager.Get(session.Id)
	if user == nil {
		return errorJSON(c, fiber.StatusNotFound, "user not found")
	}
	return c.JSON(dto.UserFrom(user))
}

func sendToken(c *fiber.Ctx, status int, user models.User) error {
	token, err := jwt.MakeToken(user)
	if err != nil {
		log.Error().Err(err).Msg("token issue failed")
		return errorJSON(c, fiber.StatusInternalServerError, "토큰을 발급할 수 없습니다.")
	}
	return c.Status(status).JSON(dto.TokenResponse{
		Token:       token,
		TokenType:   "Bearer",
		ExpiresInMs: jwt.ExpiresInMs(),
		User:        dto.UserFrom(&user),
	})
}
