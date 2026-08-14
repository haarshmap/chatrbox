package server

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

type Users struct {
	ID       int64  `json:"username" bun:"id,pk,autoincrement"`
	Username string `bun:"username,unique,notnull"`
	Password string `bun:"password,notnull"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type PageData struct {
	Title string
}

type Client struct {
	Hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}
