package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"neighbourhood/services/auth/internal/config"
	"neighbourhood/services/auth/internal/domain"
)

type RedisRepository struct {
	client *redis.Client
	ctx    context.Context
}

func New(cfg config.RedisConfig) (*RedisRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return &RedisRepository{
		client: client,
		ctx:    ctx,
	}, nil
}

func (r *RedisRepository) Close() error {
	return r.client.Close()
}

// Session methods

func (r *RedisRepository) Create(session *domain.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	ttl := time.Until(session.ExpiresAt)

	// Store by session ID
	if err := r.client.Set(r.ctx, sessionKey(session.ID), data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}

	// Store by access token for quick lookup
	if err := r.client.Set(r.ctx, accessTokenKey(session.AccessToken), session.ID, ttl).Err(); err != nil {
		return fmt.Errorf("failed to store access token mapping: %w", err)
	}

	// Store by refresh token for quick lookup
	if err := r.client.Set(r.ctx, refreshTokenKey(session.RefreshToken), session.ID, ttl).Err(); err != nil {
		return fmt.Errorf("failed to store refresh token mapping: %w", err)
	}

	// Add to user's session set
	if err := r.client.SAdd(r.ctx, userSessionsKey(session.UserID), session.ID).Err(); err != nil {
		return fmt.Errorf("failed to add to user sessions: %w", err)
	}

	return nil
}

func (r *RedisRepository) GetByID(id string) (*domain.Session, error) {
	data, err := r.client.Get(r.ctx, sessionKey(id)).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session domain.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

func (r *RedisRepository) GetByAccessToken(token string) (*domain.Session, error) {
	sessionID, err := r.client.Get(r.ctx, accessTokenKey(token)).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session ID: %w", err)
	}

	return r.GetByID(sessionID)
}

func (r *RedisRepository) GetByRefreshToken(token string) (*domain.Session, error) {
	sessionID, err := r.client.Get(r.ctx, refreshTokenKey(token)).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session ID: %w", err)
	}

	return r.GetByID(sessionID)
}

func (r *RedisRepository) GetByUserID(userID string) ([]*domain.Session, error) {
	sessionIDs, err := r.client.SMembers(r.ctx, userSessionsKey(userID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	sessions := make([]*domain.Session, 0, len(sessionIDs))

	for _, sessionID := range sessionIDs {
		session, err := r.GetByID(sessionID)
		if err != nil {
			// Session might have expired, skip it
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (r *RedisRepository) Delete(id string) error {
	session, err := r.GetByID(id)
	if err != nil {
		return err
	}

	// Delete session
	if err := r.client.Del(r.ctx, sessionKey(id)).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Delete access token mapping
	if err := r.client.Del(r.ctx, accessTokenKey(session.AccessToken)).Err(); err != nil {
		return fmt.Errorf("failed to delete access token: %w", err)
	}

	// Delete refresh token mapping
	if err := r.client.Del(r.ctx, refreshTokenKey(session.RefreshToken)).Err(); err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	// Remove from user's session set
	if err := r.client.SRem(r.ctx, userSessionsKey(session.UserID), id).Err(); err != nil {
		return fmt.Errorf("failed to remove from user sessions: %w", err)
	}

	return nil
}

func (r *RedisRepository) DeleteByUserID(userID string) error {
	sessionIDs, err := r.client.SMembers(r.ctx, userSessionsKey(userID)).Result()
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	for _, sessionID := range sessionIDs {
		if err := r.Delete(sessionID); err != nil {
			// Log error but continue
			continue
		}
	}

	// Clear the user sessions set
	if err := r.client.Del(r.ctx, userSessionsKey(userID)).Err(); err != nil {
		return fmt.Errorf("failed to delete user sessions set: %w", err)
	}

	return nil
}

func (r *RedisRepository) DeleteExpired() error {
	// Redis handles expiration automatically via TTL
	return nil
}

// Login attempt methods

func (r *RedisRepository) Record(email string) error {
	key := loginAttemptKey(email)

	// Increment attempt counter
	count, err := r.client.Incr(r.ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to record login attempt: %w", err)
	}

	// Set expiration on first attempt
	if count == 1 {
		if err := r.client.Expire(r.ctx, key, 1*time.Hour).Err(); err != nil {
			return fmt.Errorf("failed to set expiration: %w", err)
		}
	}

	return nil
}

func (r *RedisRepository) Get(email string) (*domain.LoginAttempt, error) {
	key := loginAttemptKey(email)

	count, err := r.client.Get(r.ctx, key).Int()
	if err == redis.Nil {
		return &domain.LoginAttempt{
			Email:    email,
			Attempts: 0,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get login attempts: %w", err)
	}

	lockKey := loginLockKey(email)
	lockedUntil, err := r.client.Get(r.ctx, lockKey).Result()

	var lockedUntilTime *time.Time
	if err == nil {
		t, err := time.Parse(time.RFC3339, lockedUntil)
		if err == nil {
			lockedUntilTime = &t
		}
	}

	return &domain.LoginAttempt{
		Email:       email,
		Attempts:    count,
		LockedUntil: lockedUntilTime,
		LastAttempt: time.Now(),
	}, nil
}

func (r *RedisRepository) Reset(email string) error {
	key := loginAttemptKey(email)
	lockKey := loginLockKey(email)

	if err := r.client.Del(r.ctx, key, lockKey).Err(); err != nil {
		return fmt.Errorf("failed to reset login attempts: %w", err)
	}

	return nil
}

func (r *RedisRepository) IsLocked(email string) (bool, error) {
	lockKey := loginLockKey(email)

	exists, err := r.client.Exists(r.ctx, lockKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check lock status: %w", err)
	}

	// If locked, check if attempts exceed threshold and lock if needed
	if exists == 0 {
		attempt, err := r.Get(email)
		if err != nil {
			return false, err
		}

		// Lock after 5 failed attempts (configurable)
		if attempt.Attempts >= 5 {
			lockUntil := time.Now().Add(30 * time.Minute)
			if err := r.client.Set(r.ctx, lockKey, lockUntil.Format(time.RFC3339), 30*time.Minute).Err(); err != nil {
				return false, fmt.Errorf("failed to set lock: %w", err)
			}
			return true, nil
		}
	}

	return exists > 0, nil
}

// Helper functions for Redis keys

func sessionKey(id string) string {
	return fmt.Sprintf("session:%s", id)
}

func accessTokenKey(token string) string {
	return fmt.Sprintf("access_token:%s", token)
}

func refreshTokenKey(token string) string {
	return fmt.Sprintf("refresh_token:%s", token)
}

func userSessionsKey(userID string) string {
	return fmt.Sprintf("user_sessions:%s", userID)
}

func loginAttemptKey(email string) string {
	return fmt.Sprintf("login_attempt:%s", email)
}

func loginLockKey(email string) string {
	return fmt.Sprintf("login_lock:%s", email)
}
