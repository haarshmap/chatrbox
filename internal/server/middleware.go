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
		if err != nil {
			fmt.Fprintf(w, "%v", err)
		}

		tokenString := cookie.Value

		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			return []byte(os.Getenv("SIGNING_KEY")), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			slog.Log(r.Context(), 8, "unauthorized")
			fmt.Fprintf(w, "Unauthorized %v", err)
		}

		next.ServeHTTP(w, r)
	})
}
