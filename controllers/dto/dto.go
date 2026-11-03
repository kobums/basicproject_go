package dto

// 인증 API 응답 구조체 (Spring AuthController 의 DTO 와 같은 JSON).
// 테이블 CRUD 응답은 gomachine 생성 코드가 만든다 ({items, total, code} 등).

import "basicproject_go/models"

// UserResponse 는 비밀번호를 뺀 사용자 정보다.
type UserResponse struct {
	Id        int64  `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Level     int    `json:"level"`
	CreatedAt string `json:"createdAt"`
}

func UserFrom(u *models.User) UserResponse {
	return UserResponse{Id: u.Id, Email: u.Email, Name: u.Name, Level: u.Level, CreatedAt: u.Createddate}
}

// TokenResponse 는 로그인 / 회원가입 응답이다.
type TokenResponse struct {
	Token       string       `json:"token"`
	TokenType   string       `json:"tokenType"`
	ExpiresInMs int64        `json:"expiresInMs"`
	User        UserResponse `json:"user"`
}
