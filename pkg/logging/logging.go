package logging

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

type Config struct {
	Level      string
	Structured bool
}

func New(ctx context.Context, cfg Config) (logr.Logger, context.Context, error) {
	zapCfg := zap.NewProductionConfig()
	zapCfg.OutputPaths = []string{"stdout"}

	switch cfg.Level {
	case "debug":
		zapCfg.Level.SetLevel(zap.DebugLevel)
	case "info":
		zapCfg.Level.SetLevel(zap.InfoLevel)
	case "warn":
		zapCfg.Level.SetLevel(zap.WarnLevel)
	case "error":
		zapCfg.Level.SetLevel(zap.ErrorLevel)
	default:
		return logr.Logger{}, nil, errors.New("unknown log level")
	}

	if cfg.Structured {
		zapCfg.Encoding = "json"
	} else {
		zapCfg.Encoding = "console"
	}

	zapLogger, err := zapCfg.Build()
	if err != nil {
		return logr.Logger{}, nil, fmt.Errorf("failed to build zap logger, %w", err)
	}

	logger := zapr.NewLogger(zapLogger)
	ctx = logr.NewContext(ctx, logger)
	return logger, ctx, nil
}
