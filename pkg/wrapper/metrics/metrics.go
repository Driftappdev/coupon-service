package metrics

import (
	"net/http"

	gometrics "github.com/driftappdev/libpackage/gometrics"
)

func Handler() http.HandlerFunc {
	return gometrics.Handler()
}
