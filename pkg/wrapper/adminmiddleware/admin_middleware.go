package adminmiddleware

import "net/http"

func ChiRequireRoles(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler { return next }
}
