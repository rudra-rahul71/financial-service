package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/plaid/plaid-go/v40/plaid"
	"github.com/rudra-rahul71/financial-service/internal/middleware"
	"github.com/rudra-rahul71/financial-service/internal/plaidclient"
	"github.com/rudra-rahul71/financial-service/internal/storage"
)

func CreateLinkToken(plaidClient *plaid.APIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := middleware.GetIDToken(r.Context())
		ctx := r.Context()

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
			return
		}
	}
}

func ExchangePublicToken(client *plaid.APIClient, store *storage.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		publicToken := r.PathValue("publicToken")

		token := middleware.GetIDToken(r.Context())
		ctx := r.Context()

		request := plaid.NewItemPublicTokenExchangeRequest(publicToken)

		resp, _, err := client.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(*request).Execute()

		if err != nil {
			http.Error(w, "Error exchanging token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		err = store.AddAccount(ctx, token.UID, resp)
		if err != nil {
			http.Error(w, "error adding document: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func SearchAccounts(client *plaid.APIClient, store *storage.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := middleware.GetIDToken(r.Context())

		accounts, err := store.GetUserAccounts(r.Context(), token.UID)
		if err != nil {
			http.Error(w, "Error getting document: "+err.Error(), http.StatusInternalServerError)
			return
		}

		days := r.PathValue("days")
		i, err := strconv.Atoi(days)
		if err != nil {
			http.Error(w, "Invalid days parameter: "+err.Error(), http.StatusBadRequest)
			return
		}

		resp := []plaid.TransactionsGetResponse{}
		for _, account := range accounts {
			trans, err := plaidclient.GetTransactions(r.Context(), client, account.AccessToken, i)
			if err != nil {
				http.Error(w, "Error getting transactions: "+err.Error(), http.StatusInternalServerError)
				return
			}
			resp = append(resp, *trans)
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
			return
		}
	}
}
