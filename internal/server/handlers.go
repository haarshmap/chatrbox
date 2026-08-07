package server

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
)

var db *bun.DB

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	templ, err := template.ParseFiles("templates/index.tmpl")
	if err != nil {
		slog.Log(r.Context(), 8, "Failed to parse index.tmpl")
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
		slog.Log(r.Context(), 8, "Failed to execute login.tmpl")
		fmt.Fprintf(w, "%v", err)
	}
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	name := r.FormValue("username")
	password := r.FormValue("password")

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		slog.Log(r.Context(), 8, "failed to hash password")
		fmt.Fprintf(w, "failed to hash password: %v", err)
	}
	user := &Users{Username: name, Password: string(hash)}
	_, err = db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		slog.Log(r.Context(), 0, "failed to initialise row")
		fmt.Fprintf(w, "failed to initialise row: %v", err)
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
