package token

import (
	"crypto/md5"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const tokenTypeAccess = "access"
const tokenTypeRefresh = "refresh"

// JWTMaker is a JSON Web Token maker
type JWTMaker struct {
	issuer          string
	accessSecret    *rsa.PrivateKey
	accessPublic    *rsa.PublicKey
	refreshSecret   *rsa.PrivateKey
	refreshPublic   *rsa.PublicKey
	accessDuration  time.Duration
	refreshDuration time.Duration
}

func NewJWTMaker(issuer string, accessSecretKey string, accessPublicKey string, refreshSecretKey string, refreshPublicKey string, accesssDuration time.Duration, refreshDuration time.Duration) (Maker, error) {
	accessSecret, err := GetPrivateKey(accessSecretKey)
	if err != nil {
		return nil, err
	}
	accessPublic, err := GetPublicKey(accessPublicKey)
	if err != nil {
		return nil, err
	}

	refreshSecret, err := GetPrivateKey(refreshSecretKey)
	if err != nil {
		return nil, err
	}
	refreshPublic, err := GetPublicKey(refreshPublicKey)
	if err != nil {
		return nil, err
	}

	return &JWTMaker{
		issuer:          issuer,
		accessSecret:    accessSecret,
		accessPublic:    accessPublic,
		refreshSecret:   refreshSecret,
		refreshPublic:   refreshPublic,
		accessDuration:  accesssDuration,
		refreshDuration: refreshDuration,
	}, nil
}

func (maker *JWTMaker) CreateAccessToken(userID int64) (string, error) {
	payload, err := NewPayload(userID, tokenTypeAccess, maker.issuer, maker.accessDuration)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, payload)

	return token.SignedString(maker.accessSecret)
}

func (maker *JWTMaker) CreateRefreshToken(userID int64, tokenHash string) (string, error) {

	cusKey := maker.GenerateCustomKey(userID, tokenHash)

	payload, err := NewRefreshPayload(userID, cusKey, tokenTypeRefresh, maker.issuer, maker.refreshDuration)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, payload)

	return token.SignedString(maker.refreshSecret)
}

func (maker *JWTMaker) ValidateAccessToken(token string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodRSA)
		if !ok {
			return nil, ErrInvalidToken
		}
		return maker.accessPublic, nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	payload, ok := jwtToken.Claims.(*Payload)
	if !ok || !jwtToken.Valid || payload.UserID == 0 || payload.Issuer != maker.issuer || payload.Type != tokenTypeAccess {
		return nil, ErrInvalidToken
	}

	return payload, nil
}

func (maker *JWTMaker) ValidateRefreshToken(token string) (*RefreshPayload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodRSA)
		if !ok {
			return nil, ErrInvalidToken
		}
		return maker.refreshPublic, nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &RefreshPayload{}, keyFunc)
	if err != nil {
		return nil, ErrInvalidToken
	}

	payload, ok := jwtToken.Claims.(*RefreshPayload)
	if !ok || !jwtToken.Valid || payload.UserID == 0 || payload.Issuer != maker.issuer || payload.Type != tokenTypeRefresh {
		return nil, ErrInvalidToken
	}

	return payload, nil
}

func (maker *JWTMaker) ValidateRefreshHash(hash string, userID int64, originalHash string) error {
	if hash == maker.GenerateCustomKey(userID, originalHash) {
		return nil
	}

	return ErrInvalidToken
}

func (maker *JWTMaker) GenerateCustomKey(userID int64, tokenHash string) string {
	hash := md5.Sum([]byte(strconv.Itoa(int(userID)) + tokenHash))
	return hex.EncodeToString(hash[:])
}

func GetPrivateKey(value string) (*rsa.PrivateKey, error) {
	private, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(value))
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("could not parse private key file")
	}
	return private, nil
}

func GetPublicKey(value string) (*rsa.PublicKey, error) {
	public, err := jwt.ParseRSAPublicKeyFromPEM([]byte(value))
	if err != nil {
		return nil, errors.New("could not parse public key file")
	}
	return public, nil
}
