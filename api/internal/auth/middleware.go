package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	userIDKey contextKey = "userID"
	adminKey  contextKey = "admin"
)

func UserID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}

func IsAdmin(ctx context.Context) bool {
	admin, _ := ctx.Value(adminKey).(bool)
	return admin
}

// RequireAdmin must be used inside RequireAuth (it reads the admin flag
// RequireAuth stores in the request context).
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsAdmin(r.Context()) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[strings.TrimSpace(o)] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if allowed[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (s *Service) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims := jwt.MapClaims{}
		tok, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return s.secret, nil
		})
		if err != nil || !tok.Valid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		sub, ok := claims["sub"].(float64)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		isAdmin, _ := claims["adm"].(bool)

		ctx := context.WithValue(r.Context(), userIDKey, int64(sub))
		ctx = context.WithValue(ctx, adminKey, isAdmin)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
