package users

import (
	"context"
	"fmt"
)

type Service interface {
	Login(ctx context.Context, username, password string) (*User, error)
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
	//Security : we should mask the cause of the failure to the client entirely
	//TODO: once we have robust logging, actually log the error
	maskedError := fmt.Errorf("Inexistant user or wrong password")
	if err != nil {
		return nil, maskedError
	}
	if user.Password != password {
		return nil, maskedError
	}
	return user, nil
}
