package domain

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID            string
	Email         string
	PasswordHash  string
	FirstName     string
	LastName      string
	AvatarURL     string
	EmailVerified bool
	Active        bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// OAuthAccount represents an OAuth provider account linked to a user
type OAuthAccount struct {
	ID           string
	UserID       string
	Provider     string
	ProviderID   string
	Email        string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Session represents an active user session
type Session struct {
	ID           string
	UserID       string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	UserAgent    string
	IPAddress    string
	CreatedAt    time.Time
}

// LoginAttempt tracks failed login attempts for rate limiting
type LoginAttempt struct {
	Email       string
	Attempts    int
	LockedUntil *time.Time
	LastAttempt time.Time
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(user *User) error
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	Update(user *User) error
	Delete(id string) error
}

// OAuthRepository defines the interface for OAuth account data access
type OAuthRepository interface {
	CreateOAuth(account *OAuthAccount) error
	GetByProviderAndID(provider, providerID string) (*OAuthAccount, error)
	GetByUserID(userID string) ([]*OAuthAccount, error)
	UpdateOAuth(account *OAuthAccount) error
	DeleteOAuth(id string) error
}

// SessionRepository defines the interface for session management
type SessionRepository interface {
	Create(session *Session) error
	GetByID(id string) (*Session, error)
	GetByUserID(userID string) ([]*Session, error)
	GetByAccessToken(token string) (*Session, error)
	GetByRefreshToken(token string) (*Session, error)
	Delete(id string) error
	DeleteByUserID(userID string) error
	DeleteExpired() error
}

// LoginAttemptRepository defines the interface for login attempt tracking
type LoginAttemptRepository interface {
	Record(email string) error
	Get(email string) (*LoginAttempt, error)
	Reset(email string) error
	IsLocked(email string) (bool, error)
}
