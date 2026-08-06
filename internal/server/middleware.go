package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func CheckCookieAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("sessionToken")
		fmt.Println("Cookie:", cookie)
		fmt.Println("Error:", err)

		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := cookie.Value

		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			return []byte(os.Getenv("SIGNING_KEY")), nil
		})
		slog.Log(r.Context(), 8, "unauthorized", err)

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			slog.Log(r.Context(), 8, "unauthorized", err)
			return
		}

		next.ServeHTTP(w, r)
	})
}
