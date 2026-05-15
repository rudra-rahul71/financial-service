package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/plaid/plaid-go/v40/plaid"
	"github.com/rudra-rahul71/financial-service/internal/middleware"
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

type SyncRequest struct {
	Cursors map[string]string `json:"cursors"`
}

type SyncResponse struct {
	Item       plaid.Item                 `json:"item"`
	Accounts   []plaid.AccountBase        `json:"accounts"`
	Added      []plaid.Transaction        `json:"added"`
	Modified   []plaid.Transaction        `json:"modified"`
	Removed    []plaid.RemovedTransaction `json:"removed"`
	NextCursor string                     `json:"next_cursor"`
}

func SyncAccounts(client *plaid.APIClient, store *storage.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := middleware.GetIDToken(r.Context())

		var reqData SyncRequest
		if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		accounts, err := store.GetUserAccounts(r.Context(), token.UID)
		if err != nil {
			http.Error(w, "Error getting document: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var syncResponses []SyncResponse

		for _, account := range accounts {
			accountsReq := plaid.NewAccountsGetRequest(account.AccessToken)
			accountsResp, _, err := client.PlaidApi.AccountsGet(r.Context()).AccountsGetRequest(*accountsReq).Execute()
			if err != nil {
				http.Error(w, "Error getting accounts: "+err.Error(), http.StatusInternalServerError)
				return
			}

			cursor := reqData.Cursors[account.ItemId]

			var added []plaid.Transaction
			var modified []plaid.Transaction
			var removed []plaid.RemovedTransaction
			hasMore := true
			nextCursor := cursor

			for hasMore {
				syncReq := plaid.NewTransactionsSyncRequest(account.AccessToken)
				if nextCursor != "" {
					syncReq.SetCursor(nextCursor)
				}
				syncResp, _, err := client.PlaidApi.TransactionsSync(r.Context()).TransactionsSyncRequest(*syncReq).Execute()
				if err != nil {
					http.Error(w, "Error syncing transactions: "+err.Error(), http.StatusInternalServerError)
					return
				}

				added = append(added, syncResp.GetAdded()...)
				modified = append(modified, syncResp.GetModified()...)
				removed = append(removed, syncResp.GetRemoved()...)

				hasMore = syncResp.GetHasMore()
				nextCursor = syncResp.GetNextCursor()
			}

			syncResponses = append(syncResponses, SyncResponse{
				Item:       accountsResp.GetItem(),
				Accounts:   accountsResp.GetAccounts(),
				Added:      added,
				Modified:   modified,
				Removed:    removed,
				NextCursor: nextCursor,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(syncResponses)
		if err != nil {
			http.Error(w, "Error encoding JSON response: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
