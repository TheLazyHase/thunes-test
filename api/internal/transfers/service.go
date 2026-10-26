package transfers

import (
	"context"
	"fmt"

	"thunes-api/internal/users"
)

const CODE_INSUFFICIENT_FUND = "not_enough_money"
const CODE_TRANSFER_ITSELF = `transfer_to_itself`
const CODE_NEGATIVE_AMOUNT = "negative_nul_amount"

type Service interface {
	GetTransfers(ctx context.Context, user users.User) ([]Transfer, error)
	CreateTransfer(ctx context.Context, source users.User, destination users.User, amount int64) (*Transfer, string, error)
}

type service struct {
	repo Repo
	userRepo users.Repo
}

func NewService(r Repo, ruser users.Repo) (Service, error) {
	return &service{
		repo: r,
		userRepo: ruser,
	}, nil
}

func (s *service) GetTransfers(ctx context.Context, user users.User) ([]Transfer, error) {
	transactions, err := s.repo.GetByUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("Service GetTransfer: %w", err)
	}

	//Populate usernames on transfers
	//no good ways to guess the number of users without massively overestimating it
	neededUsers := make([]int64, 0)
	for _, t := range transactions {
		if t.SourceID != user.ID {
			neededUsers = append(neededUsers, t.SourceID)
		}
		if t.DestinationID != user.ID {
			neededUsers = append(neededUsers, t.DestinationID)
		}
	}
	mapUsers, err := s.userRepo.GetByIDs(ctx, neededUsers)
	if err != nil {
		return nil, fmt.Errorf("Hydrating users: %w", err)
	}

	for i, t := range transactions {
		if t.SourceID == user.ID {
			transactions[i].SourceName = user.Username
		} else {
			if u, ok := mapUsers[t.SourceID]; ok {
				transactions[i].SourceName = u.Username
			} else {
				return nil, fmt.Errorf("Source of transfer not found: id %d", t.SourceID)
			}
		}
		if t.DestinationID == user.ID {
			transactions[i].DestinationName = user.Username
		} else {
			if u, ok := mapUsers[t.DestinationID]; ok {
				transactions[i].DestinationName = u.Username
			} else {
				return nil, fmt.Errorf("Destination of transfer not found: id %d", t.SourceID)
			}
		}
	}

	return transactions, nil
}

func (s *service) CreateTransfer(ctx context.Context, source users.User, destination users.User, amount int64) (*Transfer, string, error) {
	//Transfers are always positive
	if amount <= 0 {
		return nil, CODE_NEGATIVE_AMOUNT, fmt.Errorf("Transfer with negative value: %d", amount)
	}

	if source.ID == destination.ID {
		return nil, CODE_TRANSFER_ITSELF, fmt.Errorf("Transfer must be between two different person: %s to %s", source.Username, destination.Username)
	}

	//the repo check balance too when creating the transaction too, but this check lessen load and make code clearer
	if *source.Balance < amount {
		return nil, CODE_INSUFFICIENT_FUND, fmt.Errorf("Insuffisant balance: %d vs %d", source.Balance, amount)
	}

	transaction, err := s.repo.CreateTransfer(
		ctx,
		Transfer{
			SourceID: source.ID,
			SourceName: source.Username,
			DestinationID: destination.ID,
			DestinationName: destination.Username,
			Amount: amount,
		},
	)
	if err != nil {
		return transaction, "", fmt.Errorf("Service GetTransfer: %w", err)
	}
	return transaction, "", nil
}
