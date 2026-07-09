package dto

// Spring 의 DTO(record) 들을 그대로 옮긴 응답 구조체 모음.
// 프론트엔드(basicproject_web / _vue)가 기대하는 JSON 형태와 1:1 로 맞춘다.

import (
	"strings"

	"basicproject_go/models"
)

// user_tb 응답 (비밀번호 제외). Spring UserResponse 대응.
type UserResponse struct {
	Id        int64  `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

func UserFrom(u *models.User) UserResponse {
	return UserResponse{Id: u.Id, Email: u.Email, Name: u.Name, CreatedAt: u.Date}
}

// user_tb 공개 응답 (이메일 제외). 목록/단건 등 비인증 공개 엔드포인트에서 사용.
// 익명 사용자가 이메일을 대량 수집(크롤링·계정 열거)하지 못하도록 PII(email)를 노출하지 않는다.
type UserSummary struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

func UserSummaryFrom(u *models.User) UserSummary {
	return UserSummary{Id: u.Id, Name: u.Name, CreatedAt: u.Date}
}

// 게시글 작성자 요약 (user_tb 조인 결과). Spring BoardResponse.Author 대응.
type Author struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

// board_tb + 작성자. Spring BoardResponse 대응.
type BoardResponse struct {
	Id        int64   `json:"id"`
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	ImgUrl    *string `json:"imgUrl"`
	Author    Author  `json:"author"`
	CreatedAt string  `json:"createdAt"`
}

func BoardFrom(b *models.Board, author Author) BoardResponse {
	return BoardResponse{
		Id:        b.Id,
		Title:     b.Title,
		Content:   b.Content,
		ImgUrl:    ToImgUrl(b.Img),
		Author:    author,
		CreatedAt: b.Date,
	}
}

// 외부 URL 이면 그대로, 업로드 파일명이면 /uploads/ 경로로 변환. 빈 값이면 null.
func ToImgUrl(img string) *string {
	if img == "" {
		return nil
	}
	if strings.HasPrefix(img, "http://") || strings.HasPrefix(img, "https://") || strings.HasPrefix(img, "/") {
		return &img
	}
	url := "/uploads/" + img
	return &url
}

// 페이지 응답. Spring PageResponse 대응 (page 는 0-based).
type PageResponse struct {
	Content       interface{} `json:"content"`
	Page          int         `json:"page"`
	Size          int         `json:"size"`
	TotalElements int         `json:"totalElements"`
	TotalPages    int         `json:"totalPages"`
	Last          bool        `json:"last"`
}

func NewPage(content interface{}, page int, size int, total int) PageResponse {
	totalPages := 0
	if size > 0 {
		totalPages = (total + size - 1) / size
	}
	last := page+1 >= totalPages
	return PageResponse{
		Content:       content,
		Page:          page,
		Size:          size,
		TotalElements: total,
		TotalPages:    totalPages,
		Last:          last,
	}
}

// 로그인/회원가입 응답. Spring TokenResponse 대응.
type TokenResponse struct {
	Token       string       `json:"token"`
	TokenType   string       `json:"tokenType"`
	ExpiresInMs int64        `json:"expiresInMs"`
	User        UserResponse `json:"user"`
}
