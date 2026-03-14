package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/plaid/plaid-go/v40/plaid"
	"github.com/rudra-rahul71/financial-service/internal/handler"
	"github.com/rudra-rahul71/financial-service/internal/middleware"
	"github.com/rudra-rahul71/financial-service/internal/storage"
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

	// Initialize storage service
	store := storage.NewService(firestoreClient)

	//initialize plaid
	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", "68dc0035a333020024b7bc15")
	configuration.AddDefaultHeader("PLAID-SECRET", "c2a6aae42a0dbd3c6ab7799673fe41")
	configuration.UseEnvironment(plaid.Production)
	plaidClient := plaid.NewAPIClient(configuration)

	mux := http.NewServeMux()
	mux.HandleFunc("/init", handler.CreateLinkToken(plaidClient))
	mux.HandleFunc("/create/{publicToken}", handler.ExchangePublicToken(plaidClient, store))
	mux.HandleFunc("/search/{days}", handler.SearchAccounts(plaidClient, store))
	fmt.Println("Server listening on port: 8080")

	if err := http.ListenAndServe(":8080", middleware.LoggingMiddleware(middleware.AuthMiddleware(mux, authClient))); err != nil {
		fmt.Println("Server failed to start: " + err.Error() + "!")
	}
}
