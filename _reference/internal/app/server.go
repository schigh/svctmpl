package app

import (
	"context"
	"net/http"

	"github.com/example/myservice/internal/repository"
	"github.com/example/myservice/internal/service"
	httphandler "github.com/example/myservice/internal/transport/http"
)

// initHTTPServer wires the dependency chain (repository -> service -> handler)
// and creates the *http.Server.
func (a *App) initHTTPServer() error {
	repo := repository.NewResourceRepository(a.DB)
	svc := service.NewResourceService(repo)
	handler := httphandler.NewHandler(svc, a.DB, a.Logger, a.Config.OTel.Enabled)

	a.HTTPServer = &http.Server{
		Addr:         a.Config.HTTP.Addr(),
		Handler:      handler,
		ReadTimeout:  a.Config.HTTP.ReadTimeout,
		WriteTimeout: a.Config.HTTP.WriteTimeout,
		IdleTimeout:  a.Config.HTTP.IdleTimeout,
	}

	a.onShutdown(func(ctx context.Context) error {
		return a.HTTPServer.Shutdown(ctx)
	})

	return nil
}
