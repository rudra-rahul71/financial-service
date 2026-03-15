package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"github.com/joho/godotenv"
	"github.com/plaid/plaid-go/v40/plaid"
	"github.com/rudra-rahul71/financial-service/internal/handler"
	"github.com/rudra-rahul71/financial-service/internal/middleware"
	"github.com/rudra-rahul71/financial-service/internal/storage"
	"google.golang.org/api/option"
)

func main() {
	fmt.Print(`
███████╗   ██████╗     ███████╗ ███████╗ ███████╗ ██╗   ██╗ ███████╗ ███████╗    ██╗      ██╗ ██╗   ██╗ ███████╗
██╔════╝  ██╔═══██╗    ██╔════╝ ██╔════╝ ██╔══██╗ ██║   ██║ ██╔════╝ ██╔══██╗    ██║      ██║ ██║   ██║ ██╔════╝
██║  ███╗ ██║   ██║    ███████╗ █████╗   ██████╔╝ ██║   ██║ █████╗   ██████╔╝    ██║      ██║ ██║   ██║ █████╗  
██║   ██║ ██║   ██║    ╚════██║ ██╔══╝   ██╔══██╗ ╚██╗ ██╔╝ ██╔════╝ ██╔══██╗    ██║      ██║ ╚██╗ ██╔╝ ██╔══╝  
╚██████╔╝ ╚██████╔╝    ███████║ ███████╗ ██║  ██║  ╚████╔╝  ███████╗ ██║  ██║    ███████╗ ██║  ╚████╔╝  ███████╗ ██╗ ██╗ ██╗
 ╚═════╝   ╚═════╝     ╚══════╝ ╚══════╝ ╚═╝  ╚═╝   ╚═══╝   ╚══════╝ ╚═╝  ╚═╝    ╚══════╝ ╚═╝   ╚═══╝   ╚══════╝ ╚═╝ ╚═╝ ╚═╝
`)

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error parsing it. Relying on system environment variables.")
	}

	ctx := context.Background()

	//initialize firebase
	firebaseCredentials := os.Getenv("FIREBASE_CREDENTIALS")
	if firebaseCredentials == "" {
		firebaseCredentials = "configs/serviceAccountKey.json"
	}

	opt := option.WithCredentialsFile(firebaseCredentials)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v\n", err)
	}
	authClient, err := app.Auth(ctx)
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
	plaidClientID := os.Getenv("PLAID_CLIENT_ID")
	plaidSecret := os.Getenv("PLAID_SECRET")
	plaidEnvStr := os.Getenv("PLAID_ENV")

	if plaidClientID == "" || plaidSecret == "" {
		log.Fatal("PLAID_CLIENT_ID and PLAID_SECRET environment variables must be set")
	}

	plaidEnv := plaid.Production
	if plaidEnvStr == "sandbox" || plaidEnvStr == "development" {
		plaidEnv = plaid.Sandbox
	}

	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", plaidClientID)
	configuration.AddDefaultHeader("PLAID-SECRET", plaidSecret)
	configuration.UseEnvironment(plaidEnv)
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
