package plaidclient

import (
	"context"
	"time"

	"github.com/plaid/plaid-go/v40/plaid"
)

// GetTransactions fetches transactions for the last N days
func GetTransactions(client *plaid.APIClient, accessToken string, days int) (*plaid.TransactionsGetResponse, error) {
	ctx := context.Background()

	layout := "2006-01-02"

	// Get the current time and a time 'days' ago
	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(days) * 24 * time.Hour)

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
