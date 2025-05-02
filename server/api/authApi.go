package api

import (
	"encoding/json"
	"net/http"

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
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
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
