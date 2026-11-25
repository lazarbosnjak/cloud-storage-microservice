package token

import (
	"auth-service/internal/domain"
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Issuer struct {
	key      *ecdsa.PrivateKey
	keyID    string
	issuer   string
	audience string
	ttl      time.Duration
	now      func() time.Time
}

func NewIssuer(token domain.Token) (*Issuer, error) {
	pemBytes, err := os.ReadFile(token.PrivateKeyPath)
	if err != nil {
		slog.Error("read jwt private key", "error", err)
		return nil, fmt.Errorf("read jwt private key: %w", err)
	}
	key, err := jwt.ParseECPrivateKeyFromPEM(pemBytes)
	if err != nil {
		slog.Error("parse jwt private key", "error", err)
		return nil, fmt.Errorf("parse jwt prvate key: %w", err)
	}
	if key.Curve != elliptic.P256() {
		slog.Error("jwt private key must use the P-256 curve for ES256")
		return nil, fmt.Errorf("jwt private key must use the P-256 curve for ES256")
	}
	return newIssuer(key, token), nil
}

func newIssuer(key *ecdsa.PrivateKey, token domain.Token) *Issuer {
	return &Issuer{
		key:      key,
		keyID:    token.KeyID,
		issuer:   token.Issuer,
		audience: token.Audience,
		ttl:      token.AccessTTL,
		now:      time.Now,
	}
}
