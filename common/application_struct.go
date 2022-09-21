package common

import (
	"github.com/Play-Economy37/Play.Common/configuration"
	"github.com/Play-Economy37/Play.Common/logger"
	"go.opentelemetry.io/otel/trace"
)

type CommonApplication struct {
	Config configuration.Config
	Logger *logger.Logger
	Tracer trace.Tracer
}
