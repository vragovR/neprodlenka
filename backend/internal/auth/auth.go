package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"neprodlenka/internal/clock"
	"neprodlenka/internal/config"
	"neprodlenka/internal/store"
)

const (
	TokenTTL      = 30 * 24 * time.Hour
	attemptWindow = 15 * time.Minute
	loginAttempts = 5
	ipAttempts    = 20
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTooManyAttempts    = errors.New("too many login attempts")
	ErrUnauthorized       = errors.New("unauthorized")
)

type claims struct {
	jwt.RegisteredClaims
	Pwd string `json:"pwd"`
}

type Service struct {
	store     *store.Store
	secret    []byte
	now       clock.Clock
	limiter   *limiter
	dummyHash []byte
}

func New(s *store.Store, cfg config.Config, now clock.Clock) (*Service, error) {
	dummy, err := bcrypt.GenerateFromPassword([]byte("dummy password"), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("generate dummy hash: %w", err)
	}
	return &Service{
		store:     s,
		secret:    []byte(cfg.JWTSecret),
		now:       now,
		limiter:   newLimiter(attemptWindow),
		dummyHash: dummy,
	}, nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func (s *Service) Login(ctx context.Context, login, password, ip string) (string, error) {
	now := s.now()
	loginKey, ipKey := "login:"+login, "ip:"+ip
	if !s.limiter.allowed(now, loginKey, loginAttempts) || !s.limiter.allowed(now, ipKey, ipAttempts) {
		return "", ErrTooManyAttempts
	}

	admin, ok, err := s.store.AdminByLogin(ctx, login)
	if err != nil {
		return "", err
	}
	hash := s.dummyHash
	if ok {
		hash = []byte(admin.PasswordHash)
	}
	if bcrypt.CompareHashAndPassword(hash, []byte(password)) != nil || !ok {
		s.limiter.fail(now, loginKey, ipKey)
		return "", ErrInvalidCredentials
	}
	s.limiter.reset(loginKey)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Subject:   strconv.FormatInt(admin.ID, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(TokenTTL)),
		Pwd:       fingerprint(admin.PasswordHash),
	})
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (store.Admin, error) {
	var c claims
	_, err := jwt.ParseWithClaims(token, &c,
		func(*jwt.Token) (any, error) { return s.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(s.now),
	)
	if err != nil {
		return store.Admin{}, ErrUnauthorized
	}

	id, err := strconv.ParseInt(c.Subject, 10, 64)
	if err != nil {
		return store.Admin{}, ErrUnauthorized
	}
	admin, ok, err := s.store.AdminByID(ctx, id)
	if err != nil {
		return store.Admin{}, err
	}
	if !ok || c.Pwd != fingerprint(admin.PasswordHash) {
		return store.Admin{}, ErrUnauthorized
	}
	return admin, nil
}

func fingerprint(passwordHash string) string {
	sum := sha256.Sum256([]byte(passwordHash))
	return hex.EncodeToString(sum[:8])
}
