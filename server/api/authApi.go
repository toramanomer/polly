package api

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/toramanomer/polly/primitives"
	"github.com/toramanomer/polly/repository"
)

type SignupRequest struct {
	Username primitives.Username `json:"username"`
	Email    primitives.Email    `json:"email"`
	Password primitives.Password `json:"password"`
}

func (req *SignupRequest) Validate() map[string][]string {
	errors := make(map[string][]string)

	if usernameErrors := req.Username.Validate(); usernameErrors != nil {
		errors["username"] = usernameErrors
	}

	if emailErrors := req.Email.Validate(); emailErrors != nil {
		errors["email"] = emailErrors
	}

	if passwordErrors := req.Password.Validate(); passwordErrors != nil {
		errors["password"] = passwordErrors
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

func (api *API) Signup(w http.ResponseWriter, r *http.Request) {
	var request SignupRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"type":  "invalid_request_body",
			"title": "The request body is invalid.",
		})
		return
	}

	if errors := request.Validate(); errors != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]any{
			"type":   "validation_error",
			"errors": errors,
		})
		return
	}

	user := repository.NewUser(repository.NewUserParams{
		Username: request.Username,
		Email:    request.Email,
		Password: request.Password,
	})

	if err := api.repository.CreateUser(r.Context(), user); err != nil {
		switch err {
		case repository.ErrEmailAlreadyExists:
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]any{
				"type":   "email_already_exists",
				"errors": map[string][]string{"email": {"Email already exists"}},
			})
		case repository.ErrUsernameAlreadyExists:
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]any{
				"type":   "username_already_exists",
				"errors": map[string][]string{"username": {"Username already exists"}},
			})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"type":  "internal_server_error",
				"title": "We could not process your request.",
			})
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

type SigninRequest struct {
	Email    primitives.Email    `json:"email"`
	Password primitives.Password `json:"password"`
}

func (req *SigninRequest) Validate() map[string][]string {
	errors := make(map[string][]string)

	if emailErrors := req.Email.Validate(); emailErrors != nil {
		errors["email"] = emailErrors
	}

	if passwordErrors := req.Password.Validate(); passwordErrors != nil {
		errors["password"] = passwordErrors
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

func (api *API) Signin(w http.ResponseWriter, r *http.Request) {
	var request SigninRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"type":  "invalid_request_body",
			"title": "The request body is invalid.",
		})
		return
	}

	if errors := request.Validate(); errors != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]any{
			"type":   "validation_error",
			"errors": errors,
		})
		return
	}

	user, err := api.repository.GetUserByEmail(r.Context(), request.Email)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"type":  "invalid_credentials",
			"title": "Invalid email or password.",
		})
		return
	}

	if !user.VerifyPassword(request.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"type":  "invalid_credentials",
			"title": "Invalid email or password.",
		})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SYMMETRIC_KEY")))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"type":  "internal_server_error",
			"title": "We could not process your request.",
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Set to true in production
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(time.Until(time.Now().Add(24 * time.Hour)).Seconds()),
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}

func (api *API) Me(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(os.Getenv("JWT_SYMMETRIC_KEY")), nil
	})

	if err != nil || !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    cookie.Value,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(time.Until(time.Now().Add(24 * time.Hour)).Seconds()),
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"token": cookie.Value,
	})
}

func (api *API) Signout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusOK)
}
