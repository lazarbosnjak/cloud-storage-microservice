package gateway

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"api-gateway/internal/startup/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
)

type JwtHandler struct {
	config *config.Config
	pub    *ecdsa.PublicKey
}

func NewJwtHandler(cfg *config.Config) (*JwtHandler, error) {
	h := &JwtHandler{config: cfg}
	pub, err := h.loadECPublicKey()
	if err != nil {
		return nil, err
	}
	h.pub = pub
	return h, nil
}

type AccessTokenClaims struct {
	Role      string `json:"role"`
	SessionID string `json:"sid,omitempty"`
	jwt.RegisteredClaims
}

type identityKey struct{}

type Identity struct {
	UserID    string
	Role      string
	SessionID string
}

var publicPaths = map[string]struct{}{
	"/api/auth/login":    {},
	"/api/auth/register": {},
	"/api/auth/refresh":  {},
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

func (h *JwtHandler) verifyJwt(tokenString string) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}
	_, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (any, error) { return h.pub, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodES256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer("auth-service"),
	)
	if err != nil {
		return nil, err
	}
	if claims.Subject == "" {
		return nil, errors.New("missing subject")
	}
	return claims, nil
}

func (h *JwtHandler) Middleware(next http.Handler, reject http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := publicPaths[r.URL.Path]; ok {
			next.ServeHTTP(w, r)
			return
		}

		tokenStr, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok {
			reject(w, r)
			return
		}

		claims, err := h.verifyJwt(tokenStr)
		if err != nil {
			slog.Debug("jwt rejected", "error", err)
			reject(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), identityKey{}, Identity{
			UserID:    claims.Subject,
			Role:      claims.Role,
			SessionID: claims.SessionID,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WithIdentityMetadata() runtime.ServeMuxOption {
	return runtime.WithMetadata(func(ctx context.Context, r *http.Request) metadata.MD {
		id, ok := r.Context().Value(identityKey{}).(Identity)
		if !ok {
			return nil
		}

		return metadata.Pairs(
			"x-user-id", id.UserID,
			"x-user-role", id.Role,
			"x-session-id", id.SessionID,
		)
	})
}

// BlockMetadataSpoofing stops clients from injecting metadata via Grpc-Metadata-* HTTP headers
func BlockMetadataSpoofing() runtime.ServeMuxOption {
	return runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
		if strings.HasPrefix(strings.ToLower(key), "grpc-metadata-") {
			return "", false
		}
		return runtime.DefaultHeaderMatcher(key)
	})
}
