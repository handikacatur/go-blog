package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/handikacatur/go-blog/models"
)

func GetHome(c *fiber.Ctx) error {
	tokenString := c.Cookies("token")

	claims := models.CustomClaim{}

	if len(tokenString) > 0 {

		claims, _ = claims.GetClaim(tokenString)
	}

	return c.Render("index", fiber.Map{
		"username": claims.Username,
	})
}
