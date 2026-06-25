package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionPayload struct {
	UserID    string `json:"user_id"`
	ExpiresAt int64  `json:"expires_at"`
}

type SessionService struct {
	client     *redis.Client
	prefix     string
	ttl        time.Duration
	slideAfter time.Duration
	slideTo    time.Duration
}

func NewSessionService(client *redis.Client, ttl, slideAfter time.Duration) *SessionService {
	return &SessionService{
		client:     client,
		prefix:     "session:",
		ttl:        ttl,
		slideAfter: slideAfter,
		slideTo:    ttl,
	}
}

func (s *SessionService) Create(ctx context.Context, userID string) (string, *SessionPayload, error) {
	token := randomUpper(48)

	payload := &SessionPayload{
		UserID:    userID,
		ExpiresAt: time.Now().Add(s.ttl).Unix(),
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return "", nil, err
	}

	if err := s.client.Set(ctx, s.key(token), raw, s.ttl).Err(); err != nil {
		return "", nil, err
	}
	return token, payload, nil
}

func (s *SessionService) Validate(ctx context.Context, token string) (*SessionPayload, error) {
	if token == "" {
		return nil, nil
	}

	raw, err := s.client.Get(ctx, s.key(token)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var payload SessionPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	ttl, err := s.client.TTL(ctx, s.key(token)).Result()
	if err == nil && ttl > 0 && ttl < s.slideAfter {
		// Slide the expiry, but only update the in-memory payload after Redis
		// has actually persisted it — otherwise a failed Set would leave the
		// returned ExpiresAt out of sync with what's stored.
		renewed := payload
		renewed.ExpiresAt = time.Now().Add(s.slideTo).Unix()
		if updated, marshalErr := json.Marshal(&renewed); marshalErr == nil {
			if setErr := s.client.Set(ctx, s.key(token), updated, s.slideTo).Err(); setErr == nil {
				payload.ExpiresAt = renewed.ExpiresAt
			}
		}
	}

	if payload.ExpiresAt <= time.Now().Unix() {
		_ = s.Destroy(ctx, token)
		return nil, nil
	}

	return &payload, nil
}

func (s *SessionService) Destroy(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.client.Del(ctx, s.key(token)).Err()
}

func (s *SessionService) key(token string) string {
	return s.prefix + token
}
