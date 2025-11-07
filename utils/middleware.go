package utils

import (
	"context"
	"log"
	"net/http"
	"time"

	"firebase.google.com/go/auth"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Printf(
			"Incoming request -> Method: %s | Path: %s | From: %s",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
		)

		next.ServeHTTP(w, r)

		log.Printf(
			"Request completed in %v",
			time.Since(start),
		)
	})
}

func AuthMiddleware(next http.Handler, authClient *auth.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: Missing Authorization header", http.StatusUnauthorized)
			return
		}

		idToken := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			idToken = authHeader[7:]
		}

		if idToken == "" {
			http.Error(w, "Unauthorized: Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		token, err := authClient.VerifyIDToken(r.Context(), idToken)
		if err != nil {
			log.Printf("Error verifying ID token: %v", err)
			http.Error(w, "Unauthorized: Invalid or expired token", http.StatusUnauthorized)
			return
		}

		type contextKey string
		const userContextKey contextKey = "firebaseUser"

		ctx := context.WithValue(r.Context(), userContextKey, token)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
