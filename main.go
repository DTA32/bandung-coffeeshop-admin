// Command bandung-coffeeshop-admin serves BDGCafe mission control
// (mission-control.bdgcafe.com). It has no auth of its own: run it only behind
// the Envoy basic-auth SecurityPolicy (see flux-gitops).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "time/tzdata" // DISPLAY_TZ without OS tzdata

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/initializers"
	"github.com/dta32/bandung-coffeeshop-admin/repository"
	"github.com/dta32/bandung-coffeeshop-admin/routes"
	"github.com/dta32/bandung-coffeeshop-admin/static"
	"github.com/dta32/bandung-coffeeshop-admin/storage"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	initializers.LoadEnv()

	cfg, err := initializers.LoadConfig()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	templates.TZ = cfg.DisplayTZ
	templates.ImageBase = cfg.ImageBaseURL

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := initializers.ConnectDB(ctx, cfg)
	if err != nil {
		slog.Error("db pool", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	var images *storage.R2
	if cfg.HasR2() {
		images = storage.NewR2(cfg.R2AccountID, cfg.R2AccessKeyID, cfg.R2SecretAccessKey, cfg.R2Bucket)
	} else {
		slog.Warn("R2_* not set: image uploads disabled")
	}
	templates.ImageUploads = images != nil

	h := handlers.New(controllers.New(repository.New(pool), images, cfg.ImageBaseURL), cfg)
	srv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           routes.New(h, static.FS),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		slog.Info("listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown", "err", err)
	}
}
