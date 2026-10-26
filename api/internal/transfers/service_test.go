package transfers

import (
	"context"
	"testing"
	"encoding/json"
)

func TestTransferService_CreateTransfer(t *testing.T) {
	//So my local environment on my personal computer for some reason is in golang 1.22, I decided to not do the upgrade immediatly.
	// starting with 1.24 the correct way is t.Context(), but context.TODO() should work
	//ctx := t.Context()
	ctx := context.TODO()

	repo, userRepo, userList := setup(t, ctx)

	alice := userList["Alice"]
	bob := userList["Bob"]

	service, err := NewService(repo, userRepo)
	if err != nil {
		t.Fatalf("service creation failed: %s", err.Error())
	}

	_, code, _ := service.CreateTransfer(ctx, *alice, *alice, 10)
	if code != CODE_TRANSFER_ITSELF {
		t.Fatalf("incorrect error code when transfering to oneself: %s", code)
	}

	_, code, _ = service.CreateTransfer(ctx, *alice, *bob, 10000)
	if code != CODE_INSUFFICIENT_FUND {
		t.Fatalf("incorrect error code when not having enough fund: %s", code)
	}

	_, code, _ = service.CreateTransfer(ctx, *alice, *bob, 0)
	if code != CODE_NEGATIVE_AMOUNT {
		t.Fatalf("incorrect error code when trying a negative amount: %s", code)
	}

	_, _, err = service.CreateTransfer(ctx, *alice, *bob, 10)
	if err != nil {
		t.Fatalf("failed on a valid transfer: %s", err.Error())
	}
}

func TestTransferService_GetTransfer(t *testing.T) {
	//So my local environment on my personal computer for some reason is in golang 1.22, I decided to not do the upgrade immediatly.
	// starting with 1.24 the correct way is t.Context(), but context.TODO() should work
	//ctx := t.Context()
	ctx := context.TODO()

	repo, userRepo, userList := setup(t, ctx)

	alice := userList["Alice"]
	bob := userList["Bob"]

	service, err := NewService(repo, userRepo)
	if err != nil {
		t.Fatalf("service creation failed: %s", err.Error())
	}

	//this is to setup a non-empty list of transaction
	_, _, err = service.CreateTransfer(ctx, *alice, *bob, 10)
	if err != nil {
		t.Fatalf("failed on a valid transfer: %s", err.Error())
	}

	transfers, err := service.GetTransfers(ctx,* alice)
	if err != nil {
		t.Fatalf("failed retrieving transfers: %s", err.Error())
	}
	if len(transfers) > 1 {
		t.Fatalf("too many transfers found: %d", len(transfers))
	}

	if transfers[0].SourceName != alice.Username || transfers[0].DestinationName != bob.Username || transfers[0].Amount != 10 {
		//We convert the transfers to json for easier debug
		result, err := json.Marshal(transfers)
		if err != nil {
			t.Fatalf("transfers no valid json: %s", err.Error())
		}
		t.Fatalf("wrong transfer(s) returned: %s", result)
	}

}
