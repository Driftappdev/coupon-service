package middleware

import "net/http"

func Context() func(http.Handler) http.Handler {
	return WithRequestID
}
