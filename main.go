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

	//initialize plaid
	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", "")
	configuration.AddDefaultHeader("PLAID-SECRET", "")
	configuration.UseEnvironment(plaid.Sandbox)
	client := plaid.NewAPIClient(configuration)

	print(client)

	mux := http.NewServeMux()
	mux.HandleFunc("/init", api.CreateLinkToken(client))
	fmt.Println("Server listening on port: 8080")

	if err := http.ListenAndServe(":8080", utils.LoggingMiddleware(utils.AuthMiddleware(mux, authClient))); err != nil {
		fmt.Println("Server failed to start: " + err.Error() + "!")
	}
}
