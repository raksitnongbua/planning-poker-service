package protocol

import "github.com/gofiber/fiber/v2"

const scalarHTML = `<!doctype html>
<html>
<head>
  <title>Planning Poker API</title>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<body>
  <script id="api-reference" data-url="/openapi.yaml"></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`

func docsHandler(c *fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return c.SendString(scalarHTML)
}
