package protocol

import (
	"net/http"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/raksitnongbua/planning-poker-service/configs"
	"github.com/raksitnongbua/planning-poker-service/internal/core/handler/health"
	"github.com/raksitnongbua/planning-poker-service/internal/core/handler/room"
	roomsocket "github.com/raksitnongbua/planning-poker-service/internal/core/handler/room_socket"
	"github.com/raksitnongbua/planning-poker-service/internal/core/handler/user"
	"github.com/raksitnongbua/planning-poker-service/pkg/logger"
)

func ServeREST() {
	app := fiber.New()
	app.Use(cors.New())
	app.Get("/health", health.HealthCheckHandler)
	app.Static("/openapi.yaml", "./openapi.yaml")
	app.Get("/docs", docsHandler)

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/room/:uid/:id", websocket.New(roomsocket.SocketRoomHandler))

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
