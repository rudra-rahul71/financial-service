package storage

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/plaid/plaid-go/v40/plaid"
)

type UserDocument struct {
	Accounts []plaid.ItemPublicTokenExchangeResponse `firestore:"accounts"`
}

type Service struct {
	client *firestore.Client
}

func NewService(client *firestore.Client) *Service {
	return &Service{client: client}
}

func (s *Service) AddAccount(ctx context.Context, uid string, exchangeResp plaid.ItemPublicTokenExchangeResponse) error {
	collection := s.client.Collection("users")
	docRef := collection.Doc(uid)

	updateData := map[string]interface{}{
		"accounts": firestore.ArrayUnion(exchangeResp),
	}

	_, err := docRef.Set(ctx, updateData, firestore.MergeAll)
	return err
}

func (s *Service) GetUserAccounts(ctx context.Context, uid string) ([]plaid.ItemPublicTokenExchangeResponse, error) {
	collection := s.client.Collection("users")
	docRef := collection.Doc(uid)

	docSnap, err := docRef.Get(ctx)
	if err != nil {
		return nil, err
	}

	var userDoc UserDocument
	if err := docSnap.DataTo(&userDoc); err != nil {
		return nil, err
	}

	return userDoc.Accounts, nil
}
