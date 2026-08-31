package server

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
)

var (
	db       *bun.DB
	h        *Hub
	Upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

var ctx = context.Background()

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	templ, err := template.ParseFiles("templates/index.tmpl")
	if err != nil {
		slog.Log(r.Context(), 8, "Failed to parse index.tmpl")
		fmt.Fprintf(w, "%v", err)

	}

	data := PageData{
		Title: "Testing",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = templ.Execute(w, data)
	if err != nil {
		slog.Log(r.Context(), 8, "Failed to execute index.tmpl")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RegisterHandlerPage(w http.ResponseWriter, r *http.Request) {
	templ, err := template.ParseFiles("templates/register.tmpl")
	if err != nil {
		slog.Log(r.Context(), 8, "Failed to parse register.tmpl")
		fmt.Fprintf(w, "%v", err)

	}

	data := PageData{
		Title: "Testing",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = templ.Execute(w, data)
	if err != nil {
		slog.Log(r.Context(), 8, "Failed to execute register.tmpl")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func LoginHandlerPage(w http.ResponseWriter, r *http.Request) {
	templ, err := template.ParseFiles("templates/login.tmpl")
	if err != nil {
		slog.Log(r.Context(), 8, "Failed to parse login.tmpl")
		fmt.Fprintf(w, "%v", err)

	}

	data := PageData{
		Title: "Testing",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = templ.Execute(w, data)
	if err != nil {
		slog.Log(r.Context(), 8, "Failed to execute login.tmpl")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func DashboardHandlerPage(w http.ResponseWriter, r *http.Request) {
	templ, err := template.ParseFiles("templates/dashboard.tmpl")
	if err != nil {
		slog.Log(r.Context(), 8, "Failed to parse dashboard.tmpl")
		fmt.Fprintf(w, "%v", err)
	}

	data := PageData{
		Title: "Testing",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = templ.Execute(w, data)
	if err != nil {
		slog.Log(r.Context(), 8, "Failed to execute index.tmpl")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RoomHandlerPage(w http.ResponseWriter, r *http.Request) {
	templ, err := template.ParseFiles("templates/room.tmpl")
	if err != nil {
		slog.Log(r.Context(), 8, "Failed to parse room.tmpl")
		fmt.Fprintf(w, "%v", err)
		return
	}

	data := PageData{
		Title: "Testing",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = templ.Execute(w, data)
	if err != nil {
		slog.Log(r.Context(), 8, "Failed to execute login.tmpl")
		fmt.Fprintf(w, "%v", err)
		return
	}

}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var err error

	name := r.FormValue("username")
	password := r.FormValue("password")

	validate := validator.New()

	validate.RegisterValidation("cap", passwordCheckForCaps)
	validate.RegisterValidation("num", passwordCheckForNum)
	validate.RegisterValidation("spec", passwordCheckForSpecChar)

	type ValidationErrorResponse struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}

	err = validate.Var(password, "required,cap,num,spec")
	if err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			http.Error(w, "Validation error", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		fieldErr := validationErrors[0]
		switch fieldErr.Tag() {
		case "cap":
			json.NewEncoder(w).Encode(ValidationErrorResponse{
				Error:   "Cap_Letter",
				Message: "Password must contain an uppercase letter",
			})
			return
		case "num":
			json.NewEncoder(w).Encode(ValidationErrorResponse{
				Error:   "Number",
				Message: "Password must contain a number",
			})
			return
		case "spec":
			json.NewEncoder(w).Encode(ValidationErrorResponse{
				Error:   "Special_Char",
				Message: "Password must contain a special character",
			})
			return
		}
		return
	}

	r.ParseForm()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		slog.Log(r.Context(), 8, "failed to hash password")
		fmt.Fprintf(w, "failed to hash password: %v", err)
		return
	}
	user := &Users{Username: name, Password: string(hash)}
	_, err = db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		slog.Log(r.Context(), 8, "failed to initialise row")
		fmt.Fprintf(w, "failed to initialise row: %v", err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	type ErrorResponse struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}

	name := r.FormValue("username")
	password := r.FormValue("password")

	r.ParseForm()

	user := new(Users)
	err = db.NewSelect().Model(user).Where("Username = ?", name).Scan(ctx)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "No_User_Found",
			Message: "We couldnt find the user",
		})
		return
	}

	passcheck := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if passcheck != nil {
		slog.Log(r.Context(), 8, "Password invalid")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Invalid_Password",
			Message: "Write the password again",
		})
		return
	}

	secret := []byte(os.Getenv("SECRET_KEY"))

	claims := &Claims{
		Username: name,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		slog.Log(r.Context(), 8, "failed to sign the key")
	}

	cookie := http.Cookie{
		Name:     "sessionToken",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	}
	http.SetCookie(w, &cookie)

	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusFound)

}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	var jwtKey = []byte(os.Getenv("SECRET_KEY"))

	if err != nil {
		slog.Log(r.Context(), 8, "Failed to upgrade")
		fmt.Fprintf(w, "%v", err)
		return
	}

	claims, err := CookieClaims(r, "sessionToken", jwtKey)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	client := &Client{Hub: h, conn: conn, send: make(chan []byte, 256), username: claims.Username}
	client.Hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "sessionToken",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	}

	http.SetCookie(w, &cookie)

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusFound)
		return
	}

	slog.Log(r.Context(), 0, "User logged out")

}

var chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateShortCode(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	Code := GenerateShortCode(6)
	var jwtKey = []byte(os.Getenv("SECRET_KEY"))
	var err error
	var claims *Claims

	claims, err = CookieClaims(r, "sessionToken", jwtKey)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	user := &Users{Username: claims.Username, Is_Admin: true}
	_, err = db.NewUpdate().Model(user).Set("Is_Admin=?", user.Is_Admin).Where("username=?", claims.Username).Exec(ctx)

	room := &Rooms{RoomCode: Code}
	_, err = db.NewInsert().Model(room).Exec(ctx)
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		slog.Log(r.Context(), 8, "Failed to create a room")
	}

	members := &RoomMembers{RoomID: room.RoomID, Username: claims.Username}
	_, err = db.NewInsert().Model(members).Exec(ctx)

	w.Header().Set("HX-Redirect", "/room/"+room.RoomCode)
	w.WriteHeader(http.StatusOK)
}

func JoinHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var jwtKey = []byte(os.Getenv("SECRET_KEY"))
	var claims *Claims

	room := new(Rooms)

	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		fmt.Println("parse error:", err)
	}
	roomCode := r.FormValue("roomcode")
	slog.Log(r.Context(), slog.LevelInfo, "to check roomcode", "roomCode ", roomCode)

	err = db.NewSelect().Model(room).Where("roomcode = ?", roomCode).Scan(ctx)

	_, err = db.NewSelect().Model((*Rooms)(nil)).Where("roomcode=?", roomCode).Exists(r.Context())
	if err != nil {
		slog.Error("failed to find room", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	claims, err = CookieClaims(r, "sessionToken", jwtKey)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	exists, err := db.NewSelect().Model((*RoomMembers)(nil)).Where("room_id = ? AND username = ?", claims.Username, room.RoomCode).Exists(r.Context())
	if exists {
		fmt.Fprintf(w, "user already there twin")
		return
	}

	members := &RoomMembers{RoomID: room.RoomID, Username: claims.Username}
	_, err = db.NewInsert().Model(members).Exec(ctx)

	w.Header().Set("HX-Redirect", "/room/"+room.RoomCode)
	w.WriteHeader(http.StatusOK)
}

func LeaveRoomHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var jwtKey = []byte(os.Getenv("SECRET_KEY"))

	members := new(RoomMembers)

	claims, err := CookieClaims(r, "sessionToken", jwtKey)
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		slog.Log(r.Context(), 8, "Failed to cast claims")
		return
	}

	err = db.NewSelect().Model(members).Where("username = ?", claims.Username).Scan(ctx)
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		slog.Log(r.Context(), 8, "failed to scan row")
		return
	}

	_, err = db.NewDelete().Model((*RoomMembers)(nil)).Where("username = ? AND room_id = ?", members.Username, members.RoomID).Exec(ctx)
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		slog.Log(r.Context(), 8, "failed at the query")
		return
	}

}
