package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Different types of error returned by the VerifyToken function
var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

// Payload contains the payload data of the token
type Payload struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	UserID    int64     `json:"userId"`
	Issuer    string    `json:"iss"`
	IssuedAt  time.Time `json:"iat"`
	ExpiredAt time.Time `json:"exp"`
}

type RefreshPayload struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	Hash      string    `json:"hash"`
	UserID    int64     `json:"userId"`
	Issuer    string    `json:"iss"`
	IssuedAt  time.Time `json:"iat"`
	ExpiredAt time.Time `json:"exp"`
}

func NewPayload(userID int64, tokenType string, issuer string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:        tokenID,
		Type:      tokenType,
		UserID:    userID,
		Issuer:    issuer,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration * time.Second),
	}
	return payload, nil
}

func NewRefreshPayload(userID int64, hash string, tokenType string, issuer string, duration time.Duration) (*RefreshPayload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &RefreshPayload{
		ID:        tokenID,
		Type:      tokenType,
		Hash:      hash,
		UserID:    userID,
		Issuer:    issuer,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration * time.Second),
	}
	return payload, nil
}

func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}

func (payload *RefreshPayload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}
