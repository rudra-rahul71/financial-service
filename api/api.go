package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/plaid/plaid-go/v40/plaid"
	"github.com/rudra-rahul71/financial-service/utils"
)

type UserDocument struct {
	Accounts []plaid.ItemPublicTokenExchangeResponse `firestore:"accounts"`
}

func CreateLinkToken(plaidClient *plaid.APIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := utils.GetIDToken(r.Context())
		ctx := context.Background()

		user := plaid.LinkTokenCreateRequestUser{
			ClientUserId: token.UID,
		}
		request := plaid.NewLinkTokenCreateRequest(
			"Plaid Test",
			"en",
			[]plaid.CountryCode{plaid.COUNTRYCODE_US},
		)
		request.SetUser(user)
		request.SetProducts([]plaid.Products{plaid.PRODUCTS_TRANSACTIONS})
		resp, _, err := plaidClient.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*request).Execute()

		if err != nil {
			var plaidErr plaid.GenericOpenAPIError
			if errors.As(err, &plaidErr) {
				fmt.Println("Plaid API Error: " + string(plaidErr.Body()))
			} else {
				fmt.Println("An unexpected error occurred: " + err.Error())
			}
			http.Error(w, "Failed to create link token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err2 := json.NewEncoder(w).Encode(resp)
		if err2 != nil {
			http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		}
	}
}

func ExchangePublicToken(client *plaid.APIClient, firestoreClient *firestore.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		publicToken := r.PathValue("publicToken")

		token := utils.GetIDToken(r.Context())
		ctx := context.Background()

		request := plaid.NewItemPublicTokenExchangeRequest(publicToken)

		resp, _, err := client.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(*request).Execute()

		if err != nil {
			http.Error(w, "Error exchanging token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		collection := firestoreClient.Collection("users")

		docRef := collection.Doc(token.UID)

		updateData := map[string]interface{}{
			"accounts": firestore.ArrayUnion(resp),
		}

		_, err2 := docRef.Set(ctx, updateData, firestore.MergeAll)
		if err2 != nil {
			http.Error(w, "error adding document: %v", http.StatusInternalServerError)
		}
	}
}

func SearchAccounts(client *plaid.APIClient, firestoreClient *firestore.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := utils.GetIDToken(r.Context())

		collection := firestoreClient.Collection("users")
		docRef := collection.Doc(token.UID)

		docSnap, err := docRef.Get(r.Context())
		if err != nil {
			http.Error(w, "Error getting document: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var userDoc UserDocument
		if err := docSnap.DataTo(&userDoc); err != nil {
			http.Error(w, "Error decoding document: "+err.Error(), http.StatusInternalServerError)
			return
		}

		resp := []plaid.TransactionsGetResponse{}
		for _, account := range userDoc.Accounts {
			trans, err := GetTransactions(client, account.AccessToken)
			if err != nil {
				http.Error(w, "Error getting transactions: "+err.Error(), http.StatusInternalServerError)
			}
			resp = append(resp, *trans)
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		}
	}
}

func GetTransactions(client *plaid.APIClient, accessToken string) (*plaid.TransactionsGetResponse, error) {
	ctx := context.Background()

	layout := "2006-01-02"

	// Get the current time and a time 30 days ago
	endTime := time.Now()
	year, month, _ := endTime.Date()

	startTime := time.Date(year, month, 1, 0, 0, 0, 0, endTime.Location())

	// Format the time objects into the required string format
	endDate := endTime.Format(layout)
	startDate := startTime.Format(layout)

	request := plaid.NewTransactionsGetRequest(accessToken, startDate, endDate)

	resp, _, err := client.PlaidApi.TransactionsGet(ctx).TransactionsGetRequest(*request).Execute()

	if err != nil {
		return nil, err
	}

	return &resp, nil
}
