package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mihailtudos/gophermart/internal/config"
	"github.com/mihailtudos/gophermart/internal/domain"
)

var (
	ErrTokenExpired = errors.New("token expired")
	ErrInvalidToken = errors.New("invalid token")
)

type UserAuthInput struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Manager struct {
	jwtCfg config.JWTConfig
}

func NewManager(cfg config.JWTConfig) (*Manager, error) {
	return &Manager{jwtCfg: cfg}, nil
}

func (m *Manager) NewJWT(userID string, ttl *time.Duration) (string, error) {
	if ttl == nil {
		ttl = &m.jwtCfg.AccessTokenTTL
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(*ttl).Unix(),
		Subject:   userID,
	})

	return token.SignedString([]byte(m.jwtCfg.SigningKey))
}

func (m *Manager) Parse(accessToken string) (string, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (i interface{}, err error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(m.jwtCfg.SigningKey), nil
	})

	// if err != nil {
	// 	return "", err
	// }

	// handling specific token issues
	if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorExpired != 0 {
			return "", ErrTokenExpired
		} else {
			return "", ErrInvalidToken
		}
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("error get user claims from token")
	}

	return claims["sub"].(string), nil
}

func (m *Manager) NewRefreshToken() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func (m *Manager) CreateSession(userID int, token string) (domain.Session, error) {
	if userID <= 0 || token == "" {
		return domain.Session{}, fmt.Errorf("invalid arguments")
	}

	return domain.Session{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(m.jwtCfg.RefreshTokenTTL),
	}, nil
}
