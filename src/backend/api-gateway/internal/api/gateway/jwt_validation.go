package gateway

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"api-gateway/internal/startup/config"

	"github.com/golang-jwt/jwt/v5"
)

type JwtHandler struct {
	config *config.Config
}

type AccessToken struct {
	Role      string `json:"role"`
	SessionID string `json:"sid,omitempty"`
	jwt.RegisteredClaims
}

func (h *JwtHandler) loadECPublicKey() (*ecdsa.PublicKey, error) {
	pemBytes, err := os.ReadFile(h.config.JwtPublicKeyPath)
	if err != nil {
		slog.Error("read jwt public key", "error", err)
		return nil, fmt.Errorf("read jwt public key: %w", err)
	}
	key, err := jwt.ParseECPublicKeyFromPEM(pemBytes)
	if err != nil {
		slog.Error("parse jwt public key", "error", err)
		return nil, fmt.Errorf("parse jwt prvate key: %w", err)
	}
	if key.Curve != elliptic.P256() {
		slog.Error("jwt public key must use the P-256 curve for ES256")
		return nil, fmt.Errorf("jwt public key must use the P-256 curve for ES256")
	}

	return key, nil
}

func (h *JwtHandler) verifyJwt(tokenString string, pub *ecdsa.PublicKey) (*AccessToken, error) {
	claims := &AccessToken{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (any, error) {
			return pub, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodES256.Alg()}),
	)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
