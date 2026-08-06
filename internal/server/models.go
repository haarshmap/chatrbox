package server

import "github.com/golang-jwt/jwt/v5"

type Users struct {
	ID       int64  `bun:"id,pk,autoincrement"`
	Username string `bun:"username,unique,notnull"`
	Password string `bun:"password,notnull"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}
