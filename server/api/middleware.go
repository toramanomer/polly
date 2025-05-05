package api

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
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

type turnstileResponse struct {
	Success     bool     `json:"success"`
	ErrorCodes  []string `json:"error-codes"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
}

var TURNSTILE_SECRET_KEY = os.Getenv("TURNSTILE_SECRET_KEY")

func WithTurnstileProtection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-CF-Turnstile-Token")
		if token == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"type":  "missing_turnstile_token",
				"title": "Turnstile token is required",
			})
			return
		}

		requestBody := url.Values{
			"secret":   {TURNSTILE_SECRET_KEY},
			"response": {token},
		}

		if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			requestBody.Set("remoteip", host)
		}

		resp, err := http.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", requestBody)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"type":  "turnstile_verification_failed",
				"title": "Failed to verify Turnstile token",
			})
			return
		}
		defer resp.Body.Close()

		var turnstileResp turnstileResponse
		if err := json.NewDecoder(resp.Body).Decode(&turnstileResp); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"type":  "turnstile_verification_failed",
				"title": "Failed to parse Turnstile response",
			})
			return
		}

		if !turnstileResp.Success {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"type":        "turnstile_verification_failed",
				"title":       "Turnstile verification failed",
				"error_codes": turnstileResp.ErrorCodes,
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
