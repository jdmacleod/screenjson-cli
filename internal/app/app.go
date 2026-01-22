// Package app provides the application container and dependency injection.
package app

import (
	"context"
	"fmt"

	"screenjson/cli/internal/config"
	"screenjson/cli/internal/external/gotenberg"
	"screenjson/cli/internal/external/llm"
	"screenjson/cli/internal/external/tika"
	"screenjson/cli/internal/pipeline"
	"screenjson/cli/internal/queue"
	"screenjson/cli/internal/schema"
)

// App is the main application container.
type App struct {
	Config    *config.Config
	Builder   *pipeline.Builder
	Writer    *pipeline.Writer
	Validator *schema.Validator
	Queue     *queue.Queue
	
	// External services (lazy initialized)
	gotenbergClient *gotenberg.Client
	tikaClient      *tika.Client
	llmClient       *llm.Client
}

// New creates a new application instance.
func New(cfg *config.Config) (*App, error) {
	validator, err := schema.NewValidator()
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %w", err)
	}

	app := &App{
		Config:    cfg,
		Builder:   pipeline.NewBuilder(),
		Writer:    pipeline.NewWriter(),
		Validator: validator,
		Queue:     queue.New(cfg.Server.Workers),
	}

	// Register decoders and encoders
	app.registerCodecs()

	return app, nil
}

// registerCodecs registers all format codecs with the builder and writer.
func (a *App) registerCodecs() {
	// These will be registered when the format packages are imported
	// via init() functions. This is a placeholder for explicit registration.
}

// Gotenberg returns the Gotenberg client (lazy initialized).
func (a *App) Gotenberg() *gotenberg.Client {
	if a.gotenbergClient == nil {
		a.gotenbergClient = gotenberg.NewClient(
			a.Config.Gotenberg.URL,
			a.Config.Gotenberg.Timeout,
		)
	}
	return a.gotenbergClient
}

// Tika returns the Tika client (lazy initialized).
func (a *App) Tika() *tika.Client {
	if a.tikaClient == nil {
		a.tikaClient = tika.NewClient(
			a.Config.Tika.URL,
			a.Config.Tika.Timeout,
		)
	}
	return a.tikaClient
}

// LLM returns the LLM client (lazy initialized).
func (a *App) LLM() *llm.Client {
	if a.llmClient == nil {
		a.llmClient = llm.NewClient(
			a.Config.LLM.URL,
			a.Config.LLM.APIKey,
			a.Config.LLM.Model,
			a.Config.LLM.Timeout,
		)
	}
	return a.llmClient
}

// Start starts background services.
func (a *App) Start() {
	a.Queue.Start()
}

// Stop stops background services.
func (a *App) Stop() {
	a.Queue.Stop()
}

// HealthCheck performs a health check on all external services.
func (a *App) HealthCheck(ctx context.Context) map[string]error {
	results := make(map[string]error)

	// Check Gotenberg
	if a.Config.Gotenberg.URL != "" {
		results["gotenberg"] = a.Gotenberg().Health(ctx)
	}

	// Check Tika
	if a.Config.Tika.URL != "" {
		results["tika"] = a.Tika().Health(ctx)
	}

	return results
}
