package adminmiddleware

import (
	"net/http"

	adminshield "github.com/driftappdev/libpackage/filemods/middleware/adminshield/admin-middleware"
)

func ChiRequireRoles(roles ...string) func(http.Handler) http.Handler {
	return adminshield.ChiRequireRoles(roles...)
}
