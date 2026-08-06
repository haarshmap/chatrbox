package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
)

var db *bun.DB

func registerHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	w.Write([]byte("currently at register"))
	name := r.FormValue("username")
	password := r.FormValue("password")

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		slog.Log(r.Context(), 8, "failed to hash password", err)
	}
	user := &Users{Username: name, Password: string(hash)}
	_, err = db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		slog.Log(r.Context(), 0, "failed to initialise row", err)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	ctx := context.Background()

	name := r.FormValue("username")
	password := r.FormValue("password")
	user := new(Users)
	err = db.NewSelect().Model(user).Where("Username = ?", name).Scan(ctx)

	passcheck := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if passcheck != nil {
		slog.Log(r.Context(), 8, "Password invalid")
		fmt.Fprintln(w, "\nInvalid password")
	}

	secret := []byte(os.Getenv("SIGNING_KEY"))

	claims := &Claims{
		Username: name,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		slog.Log(r.Context(), 8, "failed to sign the key")
	}

	slog.Log(r.Context(), 0, tokenString)

	cookie := http.Cookie{
		Name:     "sessionToken",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	}

	http.SetCookie(w, &cookie)
	fmt.Println("Cookie set:")
	fmt.Printf("%+v\n", cookie)
	fmt.Println("Set-Cookie header:", w.Header().Values("Set-Cookie"))
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

	fmt.Println("Cookie set:")
	fmt.Printf("%+v\n", cookie)
	slog.Log(r.Context(), 0, "User logged out")
	slog.Log(r.Context(), 0, cookie.Value)
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	slog.Log(r.Context(), 0, "currently at dashboard")
	fmt.Printf("at dashboard")
}
