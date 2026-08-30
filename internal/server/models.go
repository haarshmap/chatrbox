package server

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

type Users struct {
	ID       int64  `bun:"id,pk,autoincrement"`
	Username string `bun:"username,unique,notnull" validate:"required"`
	Password string `bun:"password,notnull" validate:"required,cap,num,spec"`
	Is_Admin bool   `bun:"is_admin,notnull,default:false"`
}

type Rooms struct {
	RoomID   int64  `bun:"roomid,pk,autoincrement"`
	RoomCode string `bun:"roomcode,unique,notnull" json:"roomcode"`
}

type RoomMembers struct {
	RoomID   int64  `bun:"room_id"`
	Username string `bun:"Username,pk"`
	Room     *Rooms `bun:"rel:belongs-to,join:room_id=roomid"`
	Users    *Users `bun:"rel:belongs-to,join:Username=username"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type PageData struct {
	Title string
}

type Client struct {
	Hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	username string
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}
