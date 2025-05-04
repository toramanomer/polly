package api

import (
	"context"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const (
	ctxKeyUserID contextKey = "userID"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(os.Getenv("JWT_SYMMETRIC_KEY")), nil
		})

		if err != nil {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			maybeUserID, ok := claims["sub"].(string)
			if !ok {
				http.Error(w, "not authorized", http.StatusUnauthorized)
				return
			}

			userID, err := uuid.Parse(maybeUserID)
			if err != nil || uuid.Nil == userID {
				http.Error(w, "not authorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ctxKeyUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})
}

func ResolveUserID(r *http.Request) uuid.UUID {
	return r.Context().Value(ctxKeyUserID).(uuid.UUID)
}
