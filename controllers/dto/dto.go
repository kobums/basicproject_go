package dto

// 인증 API 응답 구조체 (gomachine README "인증 API 계약": Go / Spring / Rust 템플릿이 같은 JSON).
// 테이블 CRUD 응답은 gomachine 생성 코드가 만든다 ({items, total, code} 등).

import "basicproject_go/models"

// UserResponse 는 비밀번호를 뺀 사용자 정보다. key 는 생성된 models.User 의 json 태그와 같다
// ({id, email, name, level, createddate}, createddate 는 yyyy-MM-dd HH:mm:ss).
type UserResponse struct {
	Id          int64  `json:"id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	Level       int    `json:"level"`
	Createddate string `json:"createddate"`
}

func UserFrom(u *models.User) UserResponse {
	return UserResponse{Id: u.Id, Email: u.Email, Name: u.Name, Level: u.Level, Createddate: u.Createddate}
}

// TokenResponse 는 로그인(200) / 회원가입(201) 응답이다. TokenType 은 항상 "Bearer".
type TokenResponse struct {
	Token       string       `json:"token"`
	TokenType   string       `json:"tokenType"`
	ExpiresInMs int64        `json:"expiresInMs"`
	User        UserResponse `json:"user"`
}
