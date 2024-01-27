package protocol

import (
	"net/http"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/raksitnongbua/planning-poker-service/internal/core"
)

func ServeREST() {
	app := fiber.New()
	app.Use(cors.New())
	app.Get("/health", core.HealthCheckHandler)
	api := app.Group("/api")

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/room/:uid/:id", websocket.New(core.SocketRoomHandler))

	v1 := api.Group("v1")
	v1.Get("/", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).SendString("Api v1 is ready!")
	})
	v1.Get("/guest/sign-in", core.SignInWithGuestHandler)
	v1.Post("/new-room", core.CreateNewRoomHandler)
	v1.Get("/room/recent-rooms/:id", core.GetRecentRoomsHandler)

	app.Listen(":3001")
}
