package service

import (
	"auth-service/internal/domain"
	"context"
	cryptorand "crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
)

type SignInResult struct {
	AccessToken        string
	ExpiresIn          time.Duration
	RefreshToken       string
	MustChangePassword bool

	OtpRequired  bool
	OtpChallenge string
}

type OtpStore interface {
	Create(ctx context.Context, otp *domain.OtpCode) error
	InvalidateAllForUser(ctx context.Context, userID string) error
	GetByID(ctx context.Context, id string) (*domain.OtpCode, error)
	Consume(ctx context.Context, id string) error
	IncrementAttempts(ctx context.Context, id string, attempts int) error
	DeleteExpired(ctx context.Context) (int64, error)
}

func (s *AuthService) SignIn(ctx context.Context, email string, password string) (SignInResult, error) {
	email = normalizeEmail(email)

	if s.validateEmail(email) != nil || s.validatePassword(password) != nil {
		return SignInResult{}, domain.ErrInvalidCredentials
	}

	cred, err := s.store.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return SignInResult{}, fmt.Errorf("get credential: %w", err)
	}

	now := time.Now()

	if cred != nil && cred.LockedUntil != nil && cred.LockedUntil.After(now) {
		return SignInResult{}, domain.ErrAccountLocked
	}

	// dummyHash is just holding random hash in case there are no users registered. Why? Because we want to make "unknown email" and "wrong password" last the same
	hash := s.dummyHash
	if cred != nil {
		hash = cred.PasswordHash
	}

	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return SignInResult{}, fmt.Errorf("compare password: %w", err)
	}

	if cred == nil {
		return SignInResult{}, domain.ErrInvalidCredentials
	}

	if !match {
		s.recordFailedSignIn(ctx, cred, now)
		return SignInResult{}, domain.ErrInvalidCredentials
	}

	if !cred.IsEnabled {
		return SignInResult{}, domain.ErrInvalidCredentials
	}

	if s.passwordExpired(cred.PasswordChangedAt) {
		return SignInResult{MustChangePassword: true}, nil
	}

	if cred.FailedAttempts > 0 || cred.LockedUntil != nil {
		if err := s.store.ResetFailedAttempts(ctx, cred.ID); err != nil {
			slog.Error("reset failed attempts", "error", err)
		}
	}

	if err := s.otp.InvalidateAllForUser(ctx, cred.ID); err != nil {
		return SignInResult{}, fmt.Errorf("invalidate old otp codes: %w", err)
	}

	code, codeHash, err := generateOtpCode()
	if err != nil {
		return SignInResult{}, fmt.Errorf("generate otp: %w", err)
	}

	challenge := &domain.OtpCode{
		UserID:    cred.ID,
		CodeHash:  codeHash,
		ExpiresAt: time.Now().Add(s.config.OTPPolicy.OtpTTL),
	}

	if err := s.otp.Create(ctx, challenge); err != nil {
		return SignInResult{}, fmt.Errorf("store otp: %w", err)
	}

	if err := s.EmailSender.SendOtpMail([]string{cred.Email}, code, s.config.OTPPolicy.OtpTTL); err != nil {
		return SignInResult{}, fmt.Errorf("sed otp mail: %w", err)
	}

	return SignInResult{
		OtpRequired:  true,
		OtpChallenge: challenge.ID,
	}, nil

}

func (s *AuthService) ChangePassword(ctx context.Context, email string, currentPassword string, newPassword string) error {
	email = normalizeEmail(email)

	if s.validateEmail(email) != nil {
		return domain.ErrInvalidCredentials
	}

	if err := s.validatePassword(newPassword); err != nil {
		return domain.ErrPassword
	}

	cred, err := s.store.GetByEmail(ctx, email)
	if err != nil {
		return domain.ErrInvalidCredentials
	}

	if cred.LockedUntil != nil && cred.LockedUntil.After(time.Now()) {
		return domain.ErrAccountLocked
	}

	match, err := argon2id.ComparePasswordAndHash(currentPassword, cred.PasswordHash)
	if err != nil {
		slog.Error("compare password", "error", err)
		return fmt.Errorf("compare password: %w", err)
	}
	if !match {
		s.recordFailedSignIn(ctx, cred, time.Now())
		return domain.ErrInvalidCredentials
	}

	// Making sure that old password is not same as new one
	if subtle.ConstantTimeCompare([]byte(newPassword), []byte(currentPassword)) == 1 {
		return domain.ErrInvalidCredentials
	}

	hash, err := argon2id.CreateHash(newPassword, s.params)
	if err != nil {
		slog.Error("hash new password", "error", err)
		return fmt.Errorf("hash new password: %w", err)
	}

	if err := s.store.ChangePassword(ctx, cred.ID, hash); err != nil {
		slog.Error("change password", "error", err)
		return fmt.Errorf("change password: %w", err)
	}

	return nil
}

func generateOtpCode() (string, []byte, error) {
	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", nil, fmt.Errorf("generate otp code: %w", err)
	}
	code := fmt.Sprintf("%06d", n)
	return code, hashToken(code), nil

}

func (s *AuthService) issueSession(ctx context.Context, cred *domain.Credential) (SignInResult, error) {
	familyID, err := uuid.NewV7()
	if err != nil {
		return SignInResult{}, fmt.Errorf("generate session id: %w", err)
	}

	raw, opaqueHash, err := generateOpaqueToken()
	if err != nil {
		return SignInResult{}, fmt.Errorf("generate refresh token: %w", err)
	}

	now := time.Now()

	rt := domain.RefreshToken{
		UserID:    cred.ID,
		FamilyID:  familyID.String(),
		TokenHash: opaqueHash,
		CreatedAt: now,
	}
	rt.ExpiresAt = now.Add(s.refreshIdleTTL)
	rt.FamilyExpiresAt = now.Add(s.refreshAbsoluteTTL)

	at, err := s.tokens.IssueAccessToken(cred.ID, cred.UserRole, rt.FamilyID)
	if err != nil {
		return SignInResult{}, fmt.Errorf("issue access token: %w", err)
	}

	return SignInResult{
		AccessToken:  at.Value,
		ExpiresIn:    at.ExpiresIn,
		RefreshToken: raw,
	}, nil
}

func (s *AuthService) VerifyOtp(ctx context.Context, challengeID string, code string) (SignInResult, error) {
	if len(challengeID) == 0 || len(code) != 6 {
		return SignInResult{}, domain.ErrOtpInvalid
	}

	otp, err := s.otp.GetByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return SignInResult{}, domain.ErrOtpInvalid
		}
		return SignInResult{}, fmt.Errorf("get opt: %w", err)
	}

	if otp.UsedAt != nil {
		return SignInResult{}, domain.ErrOtpInvalid
	}
	if time.Now().After(otp.ExpiresAt) {
		return SignInResult{}, domain.ErrOtpExpired
	}
	if otp.Attempts >= int(s.config.OTPPolicy.OtpMaxAttempts) {
		return SignInResult{}, domain.ErrOtpTooMany
	}

	if subtle.ConstantTimeCompare(hashToken(code), otp.CodeHash) != 1 {
		if err := s.otp.IncrementAttempts(ctx, otp.ID, otp.Attempts+1); err != nil {
			slog.Error("increment otp attempts", "error", err)
		}
		return SignInResult{}, domain.ErrOtpInvalid
	}

	if err := s.otp.Consume(ctx, otp.ID); err != nil {
		return SignInResult{}, fmt.Errorf("consume otp: %w", err)
	}

	cred, err := s.store.GetByID(ctx, otp.UserID)
	if err != nil {
		return SignInResult{}, fmt.Errorf("get credential: %w", err)
	}
	if !cred.IsEnabled {
		return SignInResult{}, domain.ErrInvalidCredentials
	}

	return s.issueSession(ctx, cred)
}
