package server

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/go-redis/redis_rate/v10"
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
			return []byte(os.Getenv("SECRET_KEY")), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			slog.Log(r.Context(), 8, "unauthorized")
			fmt.Fprintf(w, "Unauthorized %v", err)
		}

		next.ServeHTTP(w, r)
	})
}

//getting ip addr

func getIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return ip
	}

	return ""
}

// rate limiter
func RateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ClientIP := getIP(r)
		rate, err := limiter.Allow(r.Context(), ClientIP, redis_rate.PerMinute(20))
		if err != nil {
			fmt.Fprintf(os.Stdout, "%v", err)
			http.Error(w, "error", http.StatusBadGateway)
		}
		if rate.Allowed == 0 {
			slog.Log(r.Context(), 8, "Too many requests at once")
			fmt.Fprintf(os.Stdout, "%v", err)
			http.Error(w, "error", http.StatusTooManyRequests)
		}
		next.ServeHTTP(w, r)
	})
}
