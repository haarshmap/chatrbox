package server

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

type Users struct {
	ID       int64  `bun:"id,pk,autoincrement"`
	Username string `bun:"username,unique,notnull"`
	Password string `bun:"password,notnull"`
	Is_Admin bool   `bun:"is_admin,notnull,default:false"`
}

type Rooms struct {
	RoomID   int64  `bun:"roomid,pk,autoincrement"`
	RoomCode string `bun:"room_code,unique,notnull" json:"roomcode"`
}

type RoomMembers struct {
	RoomID int64  `bun:"room_id,pk"`
	UserID int64  `bun:"user_id,pk"`
	Room   *Rooms `bun:"rel:belongs-to,join:room_id=roomid"`
	Users  *Users `bun:"rel:belongs-to,join:user_id=id"`
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
