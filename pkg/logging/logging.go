package logging

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type Config struct {
	Level      string
	Structured bool
}

type contextKey struct{}

func New(ctx context.Context, cfg Config) (*zap.SugaredLogger, context.Context, error) {
	// Start with the production configuration
	zapCfg := zap.NewProductionConfig()

	// Set output
	zapCfg.OutputPaths = []string{"stdout"}

	// Set Level
	lvl, err := zap.ParseAtomicLevel(cfg.Level)
	if err != nil {
		return nil, nil, fmt.Errorf("unknown log level, %w", err)
	}
	zapCfg.Level = lvl

	// Set structured logging
	if cfg.Structured {
		zapCfg.Encoding = "json"
	} else {
		zapCfg.Encoding = "console"
	}

	// Build logger
	zapLogger, err := zapCfg.Build()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build zap logger, %w", err)
	}

	// Create suggared logger
	logger := zapLogger.Sugar()

	// Add suggared logger to the context
	ctx = NewContext(ctx, logger)

	return logger, ctx, nil
}

func NewContext(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

func FromContextOrDiscard(ctx context.Context) *zap.SugaredLogger {
	if v, ok := ctx.Value(contextKey{}).(*zap.SugaredLogger); ok {
		return v
	}

	return zap.NewNop().Sugar()
}
