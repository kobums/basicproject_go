package jwt

import (
	"basicproject_go/global/config"
	"basicproject_go/models"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthTokenClaims struct {
	User               models.User `json:"user"`
	jwt.StandardClaims             // 표준 토큰 Claims
}

func secretCode() []byte {
	if config.JwtSecret == "" {
		panic("jwtSecret is not set in .env.yml")
	}
	return []byte(config.JwtSecret)
}

func Check(str string) (*AuthTokenClaims, error) {
	if len(str) < 7 || str[:7] != "Bearer " {
		err := errors.New("tokek is broken")
		return nil, err
	}

	token := str[7:]

	claims := AuthTokenClaims{}
	key := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unexpected Signing Method")
		}
		return secretCode(), nil
	}

	_, err := jwt.ParseWithClaims(token, &claims, key)
	if err != nil {
		return nil, err
	}

	return &claims, nil
}

func MakeToken(item models.User) string {
	now := time.Now()
	at := AuthTokenClaims{
		User: item,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  jwt.At(now),
			ExpiresAt: jwt.At(now.Add(time.Hour * 24 * 365 * 10)),
		},
	}

	atoken := jwt.NewWithClaims(jwt.SigningMethodHS256, &at)
	signedAuthToken, _ := atoken.SignedString(secretCode())

	return signedAuthToken
}

func CheckPasswd(dbPasswd string, inputPasswd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(dbPasswd), []byte(inputPasswd))
	return err == nil
}

func GeneratePasswd(passwd string) (string, error) {
	hashedPasswd, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	} else {
		return string(hashedPasswd), err
	}
}
