package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go-vue/internal/handlers/arbitration"
	"go-vue/internal/handlers/market"
	"go-vue/internal/handlers/portfolio"
	"go-vue/internal/handlers/sniping"
	"go-vue/internal/handlers/solana"
	"go-vue/internal/handlers/system"
	"go-vue/internal/handlers/telegram"
	"go-vue/internal/handlers/token"
	"go-vue/internal/middleware"
	"go-vue/internal/services"
	"go-vue/internal/websocket"
	"go-vue/pkg/config"

	"github.com/gin-gonic/gin"
)

type Server struct {
	config     *config.Config
	router     *gin.Engine
	httpServer *http.Server
	wsHub      *websocket.Hub
}

func New(cfg *config.Config) *Server {
	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize services
	services.InitializeServices()

	router := gin.Default()

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	server := &Server{
		config: cfg,
		router: router,
		wsHub:  wsHub,
	}

	server.setupMiddleware()
	server.setupRoutes()

	return server
}

func (s *Server) setupMiddleware() {
	// CORS middleware
	s.router.Use(middleware.CORS())

	// Logging middleware
	s.router.Use(gin.Logger())

	// Recovery middleware
	s.router.Use(gin.Recovery())

	// Rate limiting middleware
	s.router.Use(middleware.RateLimit())
}

func (s *Server) setupRoutes() {
	// Serve static files
	s.router.Static("/static", "./frontend/dist")
	s.router.StaticFile("/", "./frontend/dist/index.html")

	// API routes
	api := s.router.Group("/api")
	{
		// System endpoints
		system.RegisterRoutes(api)

		// WebSocket endpoint
		api.GET("/ws", s.handleWebSocket)

		// Feature-based route groups
		telegram.RegisterRoutes(api)
		market.RegisterRoutes(api)
		solana.RegisterRoutes(api)
		portfolio.RegisterRoutes(api)
		arbitration.RegisterRoutes(api)
		sniping.RegisterRoutes(api)
		token.RegisterRoutes(api)
	}
}

func (s *Server) handleWebSocket(c *gin.Context) {
	websocket.HandleWebSocket(s.wsHub, c)
}

func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: s.router,
	}

	log.Printf("Server starting on port %d", s.config.Port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctx)
}
