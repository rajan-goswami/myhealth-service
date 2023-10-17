// Package v1 implements routing paths. Each services in own file.
package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "myhealth-service/docs"

	// Swagger docs.
	"myhealth-service/config"
	"myhealth-service/internal/usecase"
	"myhealth-service/middleware"

	"myhealth-service/pkg/logger"
)

// NewRouter -.
// Swagger spec:
//
//	@title			My Health Service APIs
//	@description	REST APIs exposed by my health services platform
//	@version		1.0
//	@host			localhost:7081
//	@BasePath		/v1
func NewRouter(handler *gin.Engine, cfg *config.Config, l logger.Interface, g usecase.GlucoseTracking) {
	// Options
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	// Swagger
	// swaggerHandler := ginSwagger.DisablingWrapHandler(swaggerFiles.Handler, "DISABLE_SWAGGER_HTTP_HANDLER")
	swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)
	handler.GET("/swagger/*any", swaggerHandler)

	// K8s probe
	handler.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Prometheus metrics
	handler.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Routers
	h := handler.Group("/v1")
	h.Use(middleware.AuthMiddleware(cfg.Security.APIKey))

	u := h.Group("/users/:id")
	{
		newGlucoseDataController(u, g, l)
	}
}
