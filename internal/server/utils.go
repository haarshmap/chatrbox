package server

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt/v5"
)

func CookieClaims(r *http.Request, cookieName string, key []byte) (*Claims, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil, err
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return key, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

var (
	capsLetterCheck  = regexp.MustCompile(`[A-Z]`)
	numCheck         = regexp.MustCompile(`[0-9]`)
	specialCharCheck = regexp.MustCompile(`[#?!@$%^&*-]`)
)

func passwordCheckForCaps(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return capsLetterCheck.MatchString(value)
}

func passwordCheckForNum(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return numCheck.MatchString(value)
}

func passwordCheckForSpecChar(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return specialCharCheck.MatchString(value)
}
