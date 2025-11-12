package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/plaid/plaid-go/v40/plaid"
	"github.com/rudra-rahul71/financial-service/api"
	"github.com/rudra-rahul71/financial-service/utils"
	"google.golang.org/api/option"
)

var authClient *auth.Client

func main() {
	fmt.Println(`
		███████╗   ██████╗     ███████╗ ███████╗ ███████╗ ██╗   ██╗ ███████╗ ███████╗    ██╗      ██╗ ██╗   ██╗ ███████╗
		██╔════╝  ██╔═══██╗    ██╔════╝ ██╔════╝ ██╔══██╗ ██║   ██║ ██╔════╝ ██╔══██╗    ██║      ██║ ██║   ██║ ██╔════╝
		██║  ███╗ ██║   ██║    ███████╗ █████╗   ██████╔╝ ██║   ██║ █████╗   ██████╔╝    ██║      ██║ ██║   ██║ █████╗  
		██║   ██║ ██║   ██║    ╚════██║ ██╔══╝   ██╔══██╗ ╚██╗ ██╔╝ ██╔══╝   ██╔══██╗    ██║      ██║ ╚██╗ ██╔╝ ██╔══╝  
		╚██████╔╝ ╚██████╔╝    ███████║ ███████╗ ██║  ██║  ╚████╔╝  ███████╗ ██║  ██║    ███████╗ ██║  ╚████╔╝  ███████╗ ██╗ ██╗ ██╗
		 ╚═════╝   ╚═════╝     ╚══════╝ ╚══════╝ ╚═╝  ╚═╝   ╚═══╝   ╚══════╝ ╚═╝  ╚═╝    ╚══════╝ ╚═╝   ╚═══╝   ╚══════╝ ╚═╝ ╚═╝ ╚═╝
	`)

	ctx := context.Background()

	//initialize firebase
	opt := option.WithCredentialsFile("configs/serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v\n", err)
	}
	authClient, err = app.Auth(ctx)
	if err != nil {
		log.Fatalf("Error getting Firebase Auth client: %v\n", err)
	}
	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("error getting Firestore client: %v", err)
	}

	//initialize plaid
	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", "")
	configuration.AddDefaultHeader("PLAID-SECRET", "")
	configuration.UseEnvironment(plaid.Sandbox)
	plaidClient := plaid.NewAPIClient(configuration)

	mux := http.NewServeMux()
	mux.HandleFunc("/init", api.CreateLinkToken(plaidClient))
	mux.HandleFunc("/create/{publicToken}", api.ExchangePublicToken(plaidClient, firestoreClient))
	mux.HandleFunc("/search/{days}", api.SearchAccounts(plaidClient, firestoreClient))
	fmt.Println("Server listening on port: 8080")

	if err := http.ListenAndServe(":8080", utils.LoggingMiddleware(utils.AuthMiddleware(mux, authClient))); err != nil {
		fmt.Println("Server failed to start: " + err.Error() + "!")
	}
}
