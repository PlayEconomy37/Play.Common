package common

import (
	"github.com/PlayEconomy37/Play.Common/configuration"
	"github.com/PlayEconomy37/Play.Common/logger"
	"go.opentelemetry.io/otel/trace"
)

type CommonApplication struct {
	Config configuration.Config
	Logger *logger.Logger
	Tracer trace.Tracer
}
