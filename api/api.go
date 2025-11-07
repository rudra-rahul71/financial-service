package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/plaid/plaid-go/v40/plaid"
	"github.com/rudra-rahul71/financial-service/utils"
)

func CreateLinkToken(client *plaid.APIClient) http.HandlerFunc {
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
		resp, _, err := client.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*request).Execute()

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
