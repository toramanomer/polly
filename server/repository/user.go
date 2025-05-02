package repository

import (
	"github.com/google/uuid"
	"github.com/toramanomer/polly/primitives"
)

type User struct {
	ID           uuid.UUID           `json:"id"`
	Username     primitives.Username `json:"username"`
	Email        primitives.Email    `json:"email"`
	PasswordHash string              `json:"-"`
}

type NewUserParams struct {
	Username primitives.Username
	Email    primitives.Email
	Password primitives.Password
}

func NewUser(params NewUserParams) *User {
	return &User{
		ID:           uuid.New(),
		Username:     params.Username,
		Email:        params.Email,
		PasswordHash: params.Password.Hash(),
	}
}

func (u *User) VerifyPassword(password primitives.Password) bool {
	return password.Verify(u.PasswordHash)
}
