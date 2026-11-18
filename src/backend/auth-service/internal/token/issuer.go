package token

import (
	"auth-service/internal/domain"
	"auth-service/internal/startup/config"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AccessClaims struct {
	Role      string `json:"role"`
	SessionID string `json:"sid,omitempty"`
	jwt.RegisteredClaims
}

type AccessToken struct {
	Value     string
	ExpiresIn time.Duration
}

type Issuer struct {
	key      *ecdsa.PrivateKey
	keyID    string
	issuer   string
	audience string
	ttl      time.Duration
	now      func() time.Time
}

func NewIssuer(cfg config.Token) (*Issuer, error) {
	pemBytes, err := os.ReadFile(cfg.PrivateKeyPath)
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
	return newIssuer(key, cfg), nil
}

func newIssuer(key *ecdsa.PrivateKey, cfg config.Token) *Issuer {
	return &Issuer{
		key:      key,
		keyID:    cfg.KeyID,
		issuer:   cfg.Issuer,
		audience: cfg.Audience,
		ttl:      cfg.AccessTTL,
		now:      time.Now,
	}
}

func (i *Issuer) IssueAccessToken(userID string, role domain.Role, sessionID string) (AccessToken, error) {
	now := i.now()

	jti, err := randomID()
	if err != nil {
		slog.Error("generate token id", "error", err)
		return AccessToken{}, fmt.Errorf("generate token id: %w", err)
	}

	claims := AccessClaims{
		Role:      role.String(),
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    i.issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{i.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(i.ttl)),
			ID:        jti,
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tok.Header["kid"] = i.keyID

	signed, err := tok.SignedString(i.key)
	if err != nil {
		slog.Error("sign access token", "error", err)
		return AccessToken{}, fmt.Errorf("sign access token: %w", err)
	}
	return AccessToken{Value: signed, ExpiresIn: i.ttl}, nil
}

func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
