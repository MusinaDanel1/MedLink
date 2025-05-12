package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type ContextKey string

const (
	UserIDKey ContextKey = "user_id"
	RoleKey   ContextKey = "role"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "Missing token", http.StatusUnauthorized)
				return
			}
			authHeader = "Bearer " + cookie.Value
		}
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			http.Error(w, "Missing JWT secret in environment variables", http.StatusInternalServerError)
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		userIDRaw, ok := claims["user_id"]
		if !ok {
			http.Error(w, "User ID missing in token", http.StatusUnauthorized)
			return
		}
		userID := fmt.Sprintf("%v", userIDRaw)

		roleRaw, ok := claims["role"]
		if !ok {
			http.Error(w, "Role missing in token", http.StatusUnauthorized)
			return
		}
		role := fmt.Sprintf("%v", roleRaw)

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, RoleKey, role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
