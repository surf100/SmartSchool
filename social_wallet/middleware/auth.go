package middleware

import "net/http"

func MockAPIKeyAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// const expectedAPIKey = "" 

// func APIKeyAuth(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		apiKey := r.Header.Get("x-api-key")
// 		if apiKey != expectedAPIKey {
// 			http.Error(w, "unauthorized", http.StatusUnauthorized)
// 			return
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }
