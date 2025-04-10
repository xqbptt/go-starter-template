package dto

import (
	"backend/db"
	"time"
)

type LoginRequest struct {
	Provider string `json:"provider" binding:"required"`
	Payload  string `json:"payload" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
type CreateUserRequest struct {
	Provider string `json:"provider" binding:"required"`
	Payload  any    `json:"payload" binding:"required"`
}

type CreateUserPayloadNormal struct {
	Name     string  `json:"name" binding:"required,ascii,min=3,max=255"`
	Email    string  `json:"email" binding:"required,email,min=3,max=255"`
	Password string  `json:"password" binding:"required,min=8,max=255"`
	Picture  *string `json:"picture"`
}

type CreateUserPayloadGoogle struct {
	AuthorizationCode string `json:"authorizationCode" binding:"required"`
}

type ConnectAuthPlatformRequest struct {
	Provider string `json:"provider" binding:"required"`
	Payload  string `json:"payload" binding:"required"`
}

type GetUserResponse struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Picture       *string   `json:"picture"`
	AuthProviders []string  `json:"authProviders"`
	EmailVerified bool      `json:"emailVerified"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func GetUserResponseFromDB(db *db.GetUserRow) *GetUserResponse {
	response := GetUserResponse{
		ID:            db.ID,
		Name:          db.Name,
		Email:         db.Email,
		Picture:       db.Picture,
		EmailVerified: db.EmailVerified,
		AuthProviders: db.AuthProviders,
		CreatedAt:     db.CreatedAt.Time,
		UpdatedAt:     db.UpdatedAt.Time,
	}

	return &response
}
