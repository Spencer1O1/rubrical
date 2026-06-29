package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	ProviderGoogle          = "google"
	DefaultSessionTTL       = 30 * 24 * time.Hour
	PasswordResetTokenTTL   = time.Hour
)

var (
	ErrEmailTaken      = errors.New("email is already registered")
	ErrNoPassword      = errors.New("password login is not available for this account")
	ErrInvalidReset    = errors.New("invalid or expired reset link")
)

type Service struct {
	pool       *pgxpool.Pool
	sessionTTL time.Duration
}

func NewService(pool *pgxpool.Pool, sessionTTL time.Duration) *Service {
	if sessionTTL <= 0 {
		sessionTTL = DefaultSessionTTL
	}
	return &Service{pool: pool, sessionTTL: sessionTTL}
}

func (s *Service) CreateUserWithPassword(ctx context.Context, email, password, displayName string) (User, error) {
	email = NormalizeEmail(email)
	if err := ValidateEmail(email); err != nil {
		return User{}, err
	}
	hash, err := HashPassword(password)
	if err != nil {
		return User{}, err
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		displayName = email
	}

	var user User
	err = s.pool.QueryRow(ctx, `
		INSERT INTO users (email, display_name, password_hash, email_verified_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, email, display_name
	`, email, displayName, hash).Scan(&user.ID, &user.Email, &user.DisplayName)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return User{}, ErrEmailTaken
		}
		return User{}, err
	}
	return user, nil
}

func (s *Service) AuthenticatePassword(ctx context.Context, email, password string) (User, error) {
	email = NormalizeEmail(email)
	var user User
	var hash *string
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, display_name, password_hash
		FROM users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.DisplayName, &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrInvalidCredentials
	}
	if err != nil {
		return User{}, err
	}
	if hash == nil || *hash == "" {
		return User{}, ErrNoPassword
	}
	if !VerifyPassword(*hash, password) {
		return User{}, ErrInvalidCredentials
	}
	return user, nil
}

func (s *Service) CreateSession(ctx context.Context, userID int64) (Session, error) {
	token, err := NewSessionToken()
	if err != nil {
		return Session{}, err
	}
	expiresAt := time.Now().UTC().Add(s.sessionTTL)
	_, err = s.pool.Exec(ctx, `
		INSERT INTO sessions (token_hash, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, HashToken(token), userID, expiresAt)
	if err != nil {
		return Session{}, err
	}
	return Session{Token: token, UserID: userID, ExpiresAt: expiresAt}, nil
}

func (s *Service) UserFromSessionToken(ctx context.Context, token string) (User, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return User{}, ErrNoUser
	}
	var user User
	var expiresAt time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT u.id, u.email, u.display_name, s.expires_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1
	`, HashToken(token)).Scan(&user.ID, &user.Email, &user.DisplayName, &expiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNoUser
	}
	if err != nil {
		return User{}, err
	}
	if time.Now().UTC().After(expiresAt) {
		_, _ = s.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, HashToken(token))
		return User{}, ErrNoUser
	}
	_, _ = s.pool.Exec(ctx, `
		UPDATE sessions SET last_seen_at = NOW() WHERE token_hash = $1
	`, HashToken(token))
	return user, nil
}

func (s *Service) RevokeSession(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, HashToken(token))
	return err
}

func (s *Service) RevokeAllSessions(ctx context.Context, userID int64) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

func (s *Service) UserByIdentity(ctx context.Context, provider, subject string) (User, error) {
	var user User
	err := s.pool.QueryRow(ctx, `
		SELECT u.id, u.email, u.display_name
		FROM user_identities i
		JOIN users u ON u.id = i.user_id
		WHERE i.provider = $1 AND i.provider_subject = $2
	`, provider, subject).Scan(&user.ID, &user.Email, &user.DisplayName)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNoUser
	}
	return user, err
}

func (s *Service) UserByEmail(ctx context.Context, email string) (User, error) {
	email = NormalizeEmail(email)
	var user User
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, display_name FROM users WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.DisplayName)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNoUser
	}
	return user, err
}

func (s *Service) LinkIdentity(ctx context.Context, userID int64, provider, subject string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO user_identities (user_id, provider, provider_subject)
		VALUES ($1, $2, $3)
		ON CONFLICT (provider, provider_subject) DO UPDATE SET user_id = EXCLUDED.user_id
	`, userID, provider, subject)
	return err
}

func (s *Service) FindOrCreateGoogleUser(ctx context.Context, subject, email, displayName string, emailVerified bool) (User, error) {
	if existing, err := s.UserByIdentity(ctx, ProviderGoogle, subject); err == nil {
		return existing, nil
	} else if !errors.Is(err, ErrNoUser) {
		return User{}, err
	}

	email = NormalizeEmail(email)
	if email != "" {
		if user, err := s.UserByEmail(ctx, email); err == nil {
			if err := s.LinkIdentity(ctx, user.ID, ProviderGoogle, subject); err != nil {
				return User{}, err
			}
			if emailVerified {
				_, _ = s.pool.Exec(ctx, `
					UPDATE users SET email_verified_at = COALESCE(email_verified_at, NOW()), updated_at = NOW()
					WHERE id = $1
				`, user.ID)
			}
			return user, nil
		} else if !errors.Is(err, ErrNoUser) {
			return User{}, err
		}
	}

	if email == "" {
		return User{}, fmt.Errorf("google account did not provide an email")
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		displayName = email
	}

	var verifiedAt any
	if emailVerified {
		verifiedAt = time.Now().UTC()
	}

	var user User
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (email, display_name, email_verified_at)
		VALUES ($1, $2, $3)
		RETURNING id, email, display_name
	`, email, displayName, verifiedAt).Scan(&user.ID, &user.Email, &user.DisplayName)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			user, err = s.UserByEmail(ctx, email)
			if err != nil {
				return User{}, err
			}
		} else {
			return User{}, err
		}
	}

	if err := s.LinkIdentity(ctx, user.ID, ProviderGoogle, subject); err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *Service) CreatePasswordResetToken(ctx context.Context, email string) (string, User, error) {
	user, err := s.UserByEmail(ctx, email)
	if errors.Is(err, ErrNoUser) {
		return "", User{}, nil
	}
	if err != nil {
		return "", User{}, err
	}

	token, err := NewSessionToken()
	if err != nil {
		return "", User{}, err
	}
	expiresAt := time.Now().UTC().Add(PasswordResetTokenTTL)
	_, err = s.pool.Exec(ctx, `
		INSERT INTO password_reset_tokens (token_hash, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, HashToken(token), user.ID, expiresAt)
	if err != nil {
		return "", User{}, err
	}
	return token, user, nil
}

func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return ErrInvalidReset
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var userID int64
	var expiresAt time.Time
	var usedAt *time.Time
	err = tx.QueryRow(ctx, `
		SELECT user_id, expires_at, used_at
		FROM password_reset_tokens
		WHERE token_hash = $1
	`, HashToken(token)).Scan(&userID, &expiresAt, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidReset
	}
	if err != nil {
		return err
	}
	if usedAt != nil || time.Now().UTC().After(expiresAt) {
		return ErrInvalidReset
	}

	_, err = tx.Exec(ctx, `
		UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2
	`, hash, userID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		UPDATE password_reset_tokens SET used_at = NOW() WHERE token_hash = $1
	`, HashToken(token))
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Service) PurgeExpiredSessions(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at < NOW()`)
	return err
}
