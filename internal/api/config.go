package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/example/api-service/internal/config"
	"github.com/example/api-service/internal/repository"
	"github.com/example/api-service/internal/api/handlers"
	"github.com/example/api-service/internal/repository/repoimpl"
)

type api struct {
	cfg      *config.APIConfig
	e        *gin.Engine
	testRepo repository.TestRepository
}

func NewAPI(e *gin.Engine, opts ...config.Func[config.APIConfig]) *api {
	cfg, err := config.NewAPIConfig(opts...)
	if err != nil {
		slog.Error("Failed to create API config", "error", err)
		cfg = &config.APIConfig{AppPort: "8080"} // fallback
	}

	api := &api{
		cfg: cfg,
		e:   e,
	}

	api.setupDefaultMiddlewares()
	api.initHealthCheck()
	api.initRepos()
	api.initRoutes()

	return api
}

func (a *api) Start() {
	defer a.Stop()
	err := a.e.Run(fmt.Sprintf(":%s", a.cfg.AppPort))
	if err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}

func (a *api) Stop() {
	slog.Info("API server is shutting down...")
}

func (a *api) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	a.e.ServeHTTP(w, req)
}

func (a *api) setupDefaultMiddlewares() {
	a.e.Use(gin.Recovery())
	a.e.Use(gin.Logger())
}

func (a *api) initHealthCheck() {
	healthHandler := handlers.NewHealthHandler()
	a.e.GET("/health", healthHandler.Health)
	a.e.GET("/ping", healthHandler.HealthSimple)
}

func (a *api) initRepos() {
	// Initialize test repository
	testRepo := repoimpl.NewTestRepository()
	a.testRepo = testRepo
}


func (a *api) initRoutes() {
	// Test routes group
	testGroup := a.e.Group("/test")
	{
		testGroup.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})
	}
}
