package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/plaid/plaid-go/v40/plaid"
	"github.com/rudra-rahul71/financial-service/utils"
)

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

		fmt.Printf("Document written with ID: %s\n", docRef.ID)
	}
}
