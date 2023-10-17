// Package app configures and runs application.
package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"myhealth-service/config"
	v1 "myhealth-service/internal/controller/http/v1"
	"myhealth-service/internal/usecase"
	"myhealth-service/internal/usecase/repo"
	"myhealth-service/pkg/httpserver"
	"myhealth-service/pkg/logger"
	"myhealth-service/pkg/postgres"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	// Database connection
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	// run migration
	if err := postgres.AutoMigrate(pg.Database()); err != nil {
		l.Fatal(fmt.Errorf("postgres migrate error: %w", err))
		panic(err.Error())
	}

	// Use case
	glucoseTrackingUseCase := usecase.NewGlucoseTracking(
		repo.NewGlucoseDataRepo(pg.Database(), l),
	)

	// HTTP Server
	handler := gin.New()
	v1.NewRouter(handler, cfg, l, glucoseTrackingUseCase)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// Graceful shutdown
	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}
