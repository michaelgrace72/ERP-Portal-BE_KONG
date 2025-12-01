package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"go-gin-clean/internal/model"

	"github.com/redis/go-redis/v9"
)

const (
	SessionKeyPrefix = "session:"
	DefaultTTL       = 30 * time.Minute // 30 minutes default session expiration
)

type SessionService struct {
	redisClient *redis.Client
	sessionTTL  time.Duration
}

func NewSessionService(redisClient *redis.Client, ttl time.Duration) *SessionService {
	if ttl == 0 {
		ttl = DefaultTTL
	}
	return &SessionService{
		redisClient: redisClient,
		sessionTTL:  ttl,
	}
}

// GenerateReferenceToken creates a cryptographically secure random reference token
func (s *SessionService) GenerateReferenceToken() (string, error) {
	bytes := make([]byte, 32) // 256-bit token
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return "ref_" + hex.EncodeToString(bytes), nil
}

// CreateSession stores the session value in Redis with the reference token as key
func (s *SessionService) CreateSession(ctx context.Context, sessionValue *model.SessionValue) (string, error) {
	// Generate reference token
	refToken, err := s.GenerateReferenceToken()
	if err != nil {
		return "", err
	}

	// Set issued at and expires at
	now := time.Now().Unix()
	sessionValue.IssuedAt = now
	sessionValue.ExpiresAt = now + int64(s.sessionTTL.Seconds())

	// Marshal session value to JSON
	valueJSON, err := json.Marshal(sessionValue)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session value: %w", err)
	}

	// Store in Redis with TTL
	key := SessionKeyPrefix + refToken
	err = s.redisClient.Set(ctx, key, valueJSON, s.sessionTTL).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store session in Redis: %w", err)
	}

	return refToken, nil
}

// GetSession retrieves the session value from Redis using the reference token
func (s *SessionService) GetSession(ctx context.Context, refToken string) (*model.SessionValue, error) {
	key := SessionKeyPrefix + refToken
	
	valueJSON, err := s.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("session not found or expired")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session from Redis: %w", err)
	}

	var sessionValue model.SessionValue
	if err := json.Unmarshal([]byte(valueJSON), &sessionValue); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session value: %w", err)
	}

	// Check if session is expired
	if time.Now().Unix() > sessionValue.ExpiresAt {
		// Delete expired session
		_ = s.DeleteSession(ctx, refToken)
		return nil, fmt.Errorf("session expired")
	}

	return &sessionValue, nil
}

// RefreshSession extends the TTL of an existing session
func (s *SessionService) RefreshSession(ctx context.Context, refToken string) error {
	key := SessionKeyPrefix + refToken
	
	// Get current session
	sessionValue, err := s.GetSession(ctx, refToken)
	if err != nil {
		return err
	}

	// Update expiration time
	now := time.Now().Unix()
	sessionValue.ExpiresAt = now + int64(s.sessionTTL.Seconds())

	// Marshal updated session value
	valueJSON, err := json.Marshal(sessionValue)
	if err != nil {
		return fmt.Errorf("failed to marshal session value: %w", err)
	}

	// Update in Redis with new TTL
	err = s.redisClient.Set(ctx, key, valueJSON, s.sessionTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to refresh session in Redis: %w", err)
	}

	return nil
}

// DeleteSession removes a session from Redis
func (s *SessionService) DeleteSession(ctx context.Context, refToken string) error {
	key := SessionKeyPrefix + refToken
	err := s.redisClient.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete session from Redis: %w", err)
	}
	return nil
}

// ValidateSession checks if a session exists and is valid
func (s *SessionService) ValidateSession(ctx context.Context, refToken string) (bool, error) {
	_, err := s.GetSession(ctx, refToken)
	if err != nil {
		return false, err
	}
	return true, nil
}

// UpdateSessionTenant updates the tenant context in an existing session
func (s *SessionService) UpdateSessionTenant(ctx context.Context, refToken string, tenantID int64, tenantSlug string, roles []string, permissions []string, scope string) error {
	key := SessionKeyPrefix + refToken
	
	// Get current session
	sessionValue, err := s.GetSession(ctx, refToken)
	if err != nil {
		return err
	}

	// Update tenant information
	sessionValue.TenantID = tenantID
	sessionValue.TenantSlug = tenantSlug
	sessionValue.Roles = roles
	sessionValue.Permissions = permissions
	sessionValue.Scope = scope

	// Marshal updated session value
	valueJSON, err := json.Marshal(sessionValue)
	if err != nil {
		return fmt.Errorf("failed to marshal session value: %w", err)
	}

	// Get remaining TTL
	ttl, err := s.redisClient.TTL(ctx, key).Result()
	if err != nil {
		ttl = s.sessionTTL
	}

	// Update in Redis with remaining TTL
	err = s.redisClient.Set(ctx, key, valueJSON, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to update session in Redis: %w", err)
	}

	return nil
}

// GetAllUserSessions retrieves all active sessions for a user
func (s *SessionService) GetAllUserSessions(ctx context.Context, userID int64) ([]string, error) {
	var sessions []string
	
	// Scan all session keys
	iter := s.redisClient.Scan(ctx, 0, SessionKeyPrefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		refToken := key[len(SessionKeyPrefix):]
		
		// Get session value
		sessionValue, err := s.GetSession(ctx, refToken)
		if err != nil {
			continue
		}
		
		// Check if session belongs to the user
		if sessionValue.UserID == userID {
			sessions = append(sessions, refToken)
		}
	}
	
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan sessions: %w", err)
	}
	
	return sessions, nil
}

// DeleteAllUserSessions removes all sessions for a user
func (s *SessionService) DeleteAllUserSessions(ctx context.Context, userID int64) error {
	sessions, err := s.GetAllUserSessions(ctx, userID)
	if err != nil {
		return err
	}
	
	for _, refToken := range sessions {
		if err := s.DeleteSession(ctx, refToken); err != nil {
			// Log error but continue deleting other sessions
			continue
		}
	}
	
	return nil
}
