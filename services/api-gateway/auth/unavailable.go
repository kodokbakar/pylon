package auth

import (
	"context"
	"fmt"
)

type UnavailableService struct {
	reason string
}

func NewUnavailableService(reason string) *UnavailableService {
	return &UnavailableService{reason: reason}
}

func (s *UnavailableService) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	_ = ctx
	_ = input

	return nil, fmt.Errorf("%w: %s", ErrUnavailable, s.reason)
}

func (s *UnavailableService) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	_ = ctx
	_ = input

	return nil, fmt.Errorf("%w: %s", ErrUnavailable, s.reason)
}

func (s *UnavailableService) Refresh(ctx context.Context, input RefreshInput) (*RefreshResult, error) {
	_ = ctx
	_ = input

	return nil, fmt.Errorf("%w: %s", ErrUnavailable, s.reason)
}
