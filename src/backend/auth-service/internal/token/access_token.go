package token

import (
	"auth-service/internal/domain"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
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

func (i *Issuer) IssueAccessToken(userID string, role domain.Role, sessionID string) (AccessToken, error) {
	now := i.now()

	jti, err := randomID()
	if err != nil {
		slog.Error("generate token id", "error", err)
		return AccessToken{}, fmt.Errorf("generate token id: %w", err)
	}

	claims := AccessClaims{
		Role:      role.String(), // we keep a role to minimize auth calls. if we didn't, gateway would verify and validate token, but have to call auth anyways to get authorizations.
		SessionID: sessionID,     // session id is pointing to refresh token family. Refresh token family ensures that same user can have different sessions. We can always use this to revoke sessions.
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    i.issuer,
			Subject:   userID, // this is id from user service. Used to minimize auth calls.
			Audience:  jwt.ClaimStrings{i.audience},
			IssuedAt:  jwt.NewNumericDate(now),            // time token is issued at
			ExpiresAt: jwt.NewNumericDate(now.Add(i.ttl)), // time token expires
			ID:        jti,                                // jti is unique identifier for each token. It's optional and prevents having one user having two identical tokens. It can also be used for future features.
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tok.Header["kid"] = i.keyID // keyID is just public key gateway could use to verify token. We can always put it in auth service with a trade-off being, we would need to call auth-service each time we need public key (on every request)

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
