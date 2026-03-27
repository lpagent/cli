package appctx

import (
	"context"

	"github.com/lpagent/cli/internal/client"
	"github.com/lpagent/cli/internal/config"
)

type ctxKey struct{}

type App struct {
	Config  *config.Config
	Client  *client.Client
	Verbose bool
	Format  string
}

func NewApp(cfg *config.Config, apiKeyFlag string, verbose bool, format string) (*App, error) {
	apiKey, err := cfg.GetAPIKey(apiKeyFlag)
	if err != nil {
		return nil, err
	}

	if format == "" {
		format = cfg.OutputFormat
	}

	return &App{
		Config:  cfg,
		Client:  client.New(cfg.BaseURL, apiKey, verbose),
		Verbose: verbose,
		Format:  format,
	}, nil
}

func WithApp(ctx context.Context, app *App) context.Context {
	return context.WithValue(ctx, ctxKey{}, app)
}

func FromContext(ctx context.Context) *App {
	app, _ := ctx.Value(ctxKey{}).(*App)
	return app
}
