package logger

import gologger "github.com/driftappdev/libpackage/gologger"

func New(service string) *gologger.Logger {
	return gologger.Default().With(gologger.F("service", service))
}
