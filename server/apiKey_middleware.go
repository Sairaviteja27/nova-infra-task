package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func APIKeyAuth(db *mongo.Database) func(http.Handler) http.Handler {
	col := db.Collection("api_keys")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			apiKey := r.Header.Get("api-key")
			apiKeyStr := strings.TrimSpace(apiKey)
			filter := bson.M{
				"key":     apiKeyStr,
				"revoked": false,
			}
			if apiKeyStr == "" {
				http.Error(w, "missing api key", http.StatusUnauthorized)
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()

			err := col.FindOne(ctx, filter).Err()
			if err != nil {
				if err == mongo.ErrNoDocuments {
					http.Error(w, "invalid api key", http.StatusUnauthorized)
					return
				}
				http.Error(w, "auth error", http.StatusInternalServerError)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
