package users

import (
	"context"
	"fmt"
)

type Service interface {
	Login(ctx context.Context, username, password string) (*User, error)
	GetUser(ctx context.Context, id int64) (*User, error)
	GetUsers(ctx context.Context) ([]User, error)
}

type service struct {
	repo Repo
}

func NewService(r Repo) (Service, error) {
	return &service{
		repo: r,
	}, nil
}

func (s *service) Login(ctx context.Context, username, password string) (*User, error) {
	user, err := s.repo.GetByUserName(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("Inexistant user")
	}
	if user.Password != password {
		return nil, fmt.Errorf("Wrong password")
	}
	return user, nil
}

func (s *service) GetUser(ctx context.Context, id int64) (*User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Inexistant user: %w", err)
	}
	return user, nil
}

func (s *service) GetUsers(ctx context.Context) ([]User, error) {
	users, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving users: %w", err)
	}
	return users, nil
}
