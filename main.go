package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	// "time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
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

	mux := http.NewServeMux()
	mux.HandleFunc("/init", CreateLinkToken())
	fmt.Println("Server listening on port: 8080")

	if err := http.ListenAndServe(":8080", utils.LoggingMiddleware(utils.AuthMiddleware(mux, authClient))); err != nil {
		fmt.Println("Server failed to start: " + err.Error() + "!")
	}
}

func CreateLinkToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "Successfully created a link token!")
	}
}

// Helper to get the token from the context within a handler
// func GetUserFromContext(ctx context.Context) (*auth.Token, bool) {
//     type contextKey string
//     const userContextKey contextKey = "firebaseUser"
//     token, ok := ctx.Value(userContextKey).(*auth.Token)
//     return token, ok
// }
