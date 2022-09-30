package common

import (
	"sync"

	"github.com/PlayEconomy37/Play.Common/configuration"
	"github.com/PlayEconomy37/Play.Common/logger"
	"go.opentelemetry.io/otel/trace"
)

// App is a Common application struct for microservices
type App struct {
	Config    configuration.Config
	Logger    *logger.Logger
	Tracer    trace.Tracer
	WaitGroup sync.WaitGroup // Used to coordinate the graceful shutdown and our background goroutines
}
