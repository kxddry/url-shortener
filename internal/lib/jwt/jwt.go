package jwt

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

var (
	ErrInvalidSigning = errors.New("unexpected signing method")
	ErrInvalidToken   = errors.New("invalid token")
	ErrMissingUserID  = errors.New("user_id missing or invalid")
	ErrNoHeader       = errors.New("no header")
	ErrInvalidHeader  = errors.New("invalid auth header format")
)

func UID(tokenString, appSecret string) (int64, error) {
	const op = "lib.jwt.ExtractUID"

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigning
		}
		return []byte(appSecret), nil
	})

	if err != nil || !token.Valid {
		return 0, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	userID, ok := claims["uid"]
	if !ok {
		return 0, fmt.Errorf("%s: %w", op, ErrMissingUserID)
	}

	switch v := userID.(type) {
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	default:
		return 0, fmt.Errorf("%s: %w", op, ErrMissingUserID)
	}
}

func UIDfromHeader(r *http.Request, appSecret string) (int64, error) {
	const op = "lib.jwt.UIDfromHeader"

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, fmt.Errorf("%s: %w", op, ErrNoHeader)
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return 0, fmt.Errorf("%s: %w", op, ErrInvalidHeader)
	}

	tokenString := parts[1]

	id, err := UID(tokenString, appSecret)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	return id, nil
}
