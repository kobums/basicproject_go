package routers

// 인증 라우트 (손으로 관리하는 파일, gomachine 이 덮어쓰지 않는다). gomachine README "인증 API 계약" 그대로다.
//   POST /api/auth/signup  {email, password, name} → 201 {token, tokenType: "Bearer", expiresInMs, user} | 400 검증 | 409 이미 있는 email
//   POST /api/auth/login   {email, password}       → 200 {token, tokenType: "Bearer", expiresInMs, user} | 401
//   GET  /api/auth/me      Authorization: Bearer   → 200 user | 401
// user = {id, email, name, level, createddate} (비밀번호 없음), 에러 = {status, message, fieldErrors?}.
// 토큰은 통일 JWT (global/jwt) 이고, 생성 라우트는 같은 토큰을 controllers.Controller.Session 으로 읽는다.

import (
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"

	"basicproject_go/controllers/dto"
	"basicproject_go/global/jwt"
	"basicproject_go/global/log"
	"basicproject_go/models"

	"github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
)

// 회원가입 기본 등급 (db/schema.sql 의 u_level DEFAULT 1 과 같다). 관리자 등급은 model.json 의 adminLevel (기본 3) 이다.
const signupLevel = 1

// db/schema.sql 의 컬럼 길이.
const (
	maxEmailLength = 255
	maxNameLength  = 100
)

// mysqlDuplicateEntry 는 unique 키 위반 에러 번호다 (가입 확인과 저장 사이에 같은 email 이 들어온 경우).
const mysqlDuplicateEntry = 1062

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
	} else if len(req.Email) > maxEmailLength {
		fields["email"] = "이메일이 너무 깁니다."
	} else if addr, err := mail.ParseAddress(req.Email); err != nil || addr.Address != req.Email {
		fields["email"] = "이메일 형식이 올바르지 않습니다."
	}
	if utf8.RuneCountInString(req.Name) > maxNameLength {
		fields["name"] = "이름은 100자 이하로 입력하세요."
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
	if manager.LastError != nil {
		log.Error().Err(manager.LastError).Msg("signup lookup failed")
		return errorJSON(c, fiber.StatusInternalServerError, "회원가입에 실패했습니다.")
	}

	hashed, err := jwt.GeneratePasswd(req.Password)
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "비밀번호 처리 중 오류가 발생했습니다.")
	}

	item := newSignupUser(req, hashed)
	if err := manager.Insert(item); err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == mysqlDuplicateEntry {
			return errorJSON(c, fiber.StatusConflict, "이미 사용 중인 이메일입니다.")
		}
		log.Error().Err(err).Msg("signup insert failed")
		return errorJSON(c, fiber.StatusInternalServerError, "회원가입에 실패했습니다.")
	}
	item.Id = manager.GetIdentity()

	// 저장된 행(createddate 포함)을 다시 읽어 돌려준다. 못 읽으면 저장한 값으로 응답한다.
	user := models.User{Id: item.Id, Email: item.Email, Name: item.Name, Level: item.Level, Createddate: item.Createddate}
	if saved := manager.Get(item.Id); saved != nil {
		user = *saved
	}
	return sendToken(c, fiber.StatusCreated, user)
}

func handleLogin(c *fiber.Ctx) error {
	var req loginRequest
	// 로그인 실패는 이유와 상관없이 401 이다 (계약: 200 | 401).
	if err := c.BodyParser(&req); err != nil {
		return errorJSON(c, fiber.StatusUnauthorized, "이메일 또는 비밀번호가 올바르지 않습니다.")
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		return errorJSON(c, fiber.StatusUnauthorized, "이메일 또는 비밀번호가 올바르지 않습니다.")
	}

	conn := models.NewConnection()
	if conn == nil {
		return errorJSON(c, fiber.StatusServiceUnavailable, "데이터베이스에 연결할 수 없습니다.")
	}
	defer conn.Close()
	manager := models.NewUserManager(conn)

	user := manager.GetWhere([]interface{}{models.Where{Column: "email", Value: req.Email, Compare: "="}})
	if user == nil && manager.LastError != nil {
		log.Error().Err(manager.LastError).Msg("login lookup failed")
		return errorJSON(c, fiber.StatusInternalServerError, "로그인 중 오류가 발생했습니다.")
	}
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
		if manager.LastError != nil {
			log.Error().Err(manager.LastError).Msg("me lookup failed")
			return errorJSON(c, fiber.StatusInternalServerError, "사용자 정보를 읽을 수 없습니다.")
		}
		// 토큰은 유효하지만 회원이 지워졌다.
		return errorJSON(c, fiber.StatusUnauthorized, "인증이 필요합니다.")
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

// newSignupUser 는 가입할 회원 행이다. level 은 항상 signupLevel (클라이언트가 정할 수 없다).
func newSignupUser(req signupRequest, hashed string) *models.UserUpdate {
	return &models.UserUpdate{Email: req.Email, Name: req.Name, Passwd: hashed, Level: signupLevel}
}
