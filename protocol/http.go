package protocol

import (
	"net/http"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"github.com/raksitnongbua/planning-poker-service/configs"
	websocketauth "github.com/raksitnongbua/planning-poker-service/internal/core/auth/websocket"
	"github.com/raksitnongbua/planning-poker-service/internal/core/handler/health"
	"github.com/raksitnongbua/planning-poker-service/internal/core/handler/room"
	roomsocket "github.com/raksitnongbua/planning-poker-service/internal/core/handler/room_socket"
	"github.com/raksitnongbua/planning-poker-service/internal/core/handler/user"
	"github.com/raksitnongbua/planning-poker-service/pkg/logger"
)

func ServeREST() {
	app := fiber.New(fiber.Config{
		BodyLimit: 1 * 1024 * 1024, // 1MB max request body (security: prevent memory exhaustion)
	})

	// CORS configuration with explicit allowed origins (security: prevent CSRF)
	app.Use(cors.New(cors.Config{
		AllowOrigins:     configs.Conf.AllowedOrigins,
		AllowMethods:     "GET,POST,DELETE",
		AllowHeaders:     "Content-Type,Cookie",
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}))

	// Rate limiting for HTTP endpoints (security: prevent resource exhaustion)
	// 200 requests per minute per IP
	app.Use(limiter.New(limiter.Config{
		Max:        200,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // Rate limit by client IP
		},
		LimitReached: func(c *fiber.Ctx) error {
			logger.Warn("rate limit exceeded", "ip", c.IP(), "path", c.Path())
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded. Please try again later.",
			})
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		Storage:                nil, // Uses in-memory storage
	}))

	app.Get("/health", health.HealthCheckHandler)
	app.Static("/openapi.yaml", "./openapi.yaml")
	app.Get("/docs", docsHandler)

	// WebSocket authentication middleware - Phase 1: log-only mode (security: monitoring cookie auth)
	// Phase 2 (rejection) will be enabled after 48h metrics confirm cookie transmission works
	app.Use("/ws", func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}

		// Try to authenticate from cookies (NextAuth session or guest UID)
		authenticatedUID, err := websocketauth.ExtractAuthenticatedUID(c)
		if err != nil {
			// Phase 1: Log but allow connection (fallback to URL param in handler)
			logger.Warn("websocket auth from cookie failed - falling back to URL param",
				"error", err, "path", c.Path(), "remote_addr", c.IP())
		} else {
			// Success: cookie auth worked
			logger.Info("websocket auth from cookie successful", "uid", authenticatedUID, "remote_addr", c.IP())
			c.Locals("authenticated_uid", authenticatedUID)
		}

		c.Locals("allowed", true)
		return c.Next()
	})

	// WebSocket configuration with security limits
	app.Get("/ws/room/:uid/:id", websocket.New(roomsocket.SocketRoomHandler, websocket.Config{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}))

	api := app.Group("/api")
	v1 := api.Group("v1")
	v1.Get("/", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).SendString("Api v1 is ready!")
	})
	v1.Get("/guest/sign-in", user.SignInWithGuestHandler)
	v1.Post("/new-room", room.CreateNewRoomHandler)
	v1.Get("/room/recent-rooms/:id", room.GetRecentRoomsHandler)
	v1.Delete("/rooms/expired", room.CleanupExpiredRoomsHandler)
	v1.Delete("/rooms/:roomId/members/:memberId", room.KickMemberHandler)

	logger.Info("server starting", "port", "8080", "env", configs.Conf.AppEnv)
	app.Listen(":8080")
}
