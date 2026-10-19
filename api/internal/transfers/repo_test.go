package transfers

import (
	"context"
	"fmt"
	"testing"

	"thunes-api/internal"
	"thunes-api/internal/users"

	"github.com/jmoiron/sqlx"
)

// We don't check specific errors for simplicity and to have more flexible tests

func emptyTransfer(ctx context.Context, db *sqlx.DB) error {
	query := `DELETE FROM transfers`
	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error deleting transfers: %w", err)
	}

	query = `DELETE FROM users`
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error deleting users: %w", err)
	}

	internal.SeedDevDB(db)

	return err
}

func setup(t *testing.T, ctx context.Context) (Repo, users.Repo, map[string]*users.User) {
	db, err := internal.GetLocalDB()
	if err != nil {
		t.Fatalf("error in local DB setup: %s", err.Error())
	}
	fmt.Println(fmt.Sprintf("db: %v, err: %v", db, err))

	err = emptyTransfer(ctx, db)
	if err != nil {
		t.Fatalf("error when emptying the db: %s", err.Error())
	}

	//transfer repo
	repo, err := NewRepo(db)
	if err != nil {
		t.Fatalf("error in repo initialization: %s", err.Error())
	}

	//user repo, for utility
	userRepo, err := users.NewRepo(db)
	if err != nil {
		t.Fatalf("error creating user repo: %s", err.Error())
	}

	names := []string{"Alice", "Bob", "Charlie"}
	userList := make(map[string]*users.User)
	for _, name := range names {
		user, err := userRepo.GetByUserName(ctx, name)
		if err != nil {
			t.Fatalf("%s not found: %s", name, err.Error())
		}

		//standard starting balance
		err = userRepo.SetBalance(ctx, user, 100)
		if err != nil {
			t.Fatalf("%s balance setting: %s", name, err.Error())
		}
		user.Balance = 100
		userList[name] = user
	}

	return repo, userRepo, userList
}

func TestTransferRepoNominalCase(t *testing.T) {
	//So my local environment on my personal computer for some reason is in golang 1.22, I decided to not do the upgrade immediatly.
	// starting with 1.24 the correct way is t.Context(), but context.TODO() should work
	//ctx := t.Context()
	ctx := context.TODO()

	repo, userRepo, userList := setup(t, ctx)

	//I feel it's nicer to not acces the map all the time, but it's personal preference
	alice := userList["Alice"]
	bob := userList["Bob"]
	charlie := userList["Charlie"]

	output, err := repo.GetByUser(ctx, users.User{ID: 1})
	if err != nil {
		t.Fatalf("error retrieving transfers: %s", err.Error())
	}
	if len(output) != 0 {
		t.Fatalf("incorrect transfers: %v", output)
	}

	//Insert three transfer, two of Alice and one of Bob, and see the result
	transfers := []Transfer{
		Transfer{SourceID: alice.ID, DestinationID: charlie.ID, Amount: 5},
		Transfer{SourceID: bob.ID, DestinationID: charlie.ID, Amount: 10},
		Transfer{SourceID: bob.ID, DestinationID: alice.ID, Amount: 5},
	}
	for _, transfer := range transfers {
		output, err := repo.CreateTransfer(ctx, transfer)
		if err != nil {
			t.Fatalf("error creating transfer: %s", err.Error())
		}
		if output.SourceID != transfer.SourceID {
			t.Fatalf("incorrect transfer: wrong source: %d instead of %d, output:%v", output.SourceID, transfer.SourceID, output)
		}
		if output.DestinationID != transfer.DestinationID {
			t.Fatalf("incorrect transfer: wrong destination: %d instead of %d, output: %v", output.DestinationID, transfer.DestinationID, output)
		}
		if output.Amount != transfer.Amount {
			t.Fatalf("incorrect transfer: wrong source: %d instead of %d, output: %v", output.Amount, transfer.Amount, output)
		}
	}

	output, err = repo.GetByUser(ctx, users.User{ID: alice.ID})
	if err != nil {
		t.Fatalf("error retrieving transfers: %s", err.Error())
	}
	// We expect 2 transfers
	if len(output) != 2 {
		t.Fatalf("incorrect transfers: %v", output)
	}

	//check balance

	//We get the users again for updated balance
	alice, err = userRepo.GetByUserName(ctx, "Alice")
	if err != nil {
		t.Fatalf("Alice not found: %s", err.Error())
	}
	if alice.Balance != 100 {
		t.Fatalf("incorrect alice balance: %v", alice)
	}

	bob, err = userRepo.GetByUserName(ctx, "Bob")
	if err != nil {
		t.Fatalf("Bob not found: %s", err.Error())
	}
	if bob.Balance != 85 {
		t.Fatalf("incorrect bob balance: %v", bob)
	}

	charlie, err = userRepo.GetByUserName(ctx, "Charlie")
	if err != nil {
		t.Fatalf("Charlie not found: %s", err.Error())
	}
	if charlie.Balance != 115 {
		t.Fatalf("incorrect charlie balance: %v", charlie)
	}

	//test trying an illegal transfer
	transfer := Transfer{
		SourceID: alice.ID,
		DestinationID: charlie.ID,
		Amount: 2000,
	}
	erroneousOutput, err := repo.CreateTransfer(ctx, transfer)
	if err == nil {
		t.Fatalf("Accepting illegal transfer: %v", erroneousOutput)
	}
	//TODO: check with a regexp the error message
	//We don't want exact error matches in tests, but checking the important information is here is good
}
