package websocketauth

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/raksitnongbua/planning-poker-service/constants"
	"github.com/raksitnongbua/planning-poker-service/internal/core/usecase/profile"
)

// ExtractAuthenticatedUID extracts and verifies the user ID from authentication cookies.
// Checks NextAuth session cookie first (authenticated users), falls back to guest UID cookie.
// Returns the authenticated UID and an error if no valid authentication is found.
func ExtractAuthenticatedUID(c *fiber.Ctx) (string, error) {
	// Try NextAuth session cookie first (authenticated users)
	var sessionCookie string
	if c.Secure() {
		sessionCookie = c.Cookies(constants.SecureSessionCookie)
	} else {
		sessionCookie = c.Cookies(constants.SessionCookie)
	}

	if sessionCookie != "" {
		p, err := profile.GetProfile(sessionCookie)
		if err == nil && p != nil && p.UID != "" {
			return p.UID, nil
		}
		// If decryption fails, fall through to check guest cookie
	}

	// Fall back to guest UID cookie
	guestUID := c.Cookies("CPPUniID")
	if guestUID != "" {
		return guestUID, nil
	}

	return "", errors.New("no valid authentication found")
}
