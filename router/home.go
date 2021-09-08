package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/handikacatur/go-blog/handlers"
)

func SetHome(app *fiber.App) {
	home := app.Group("/")
	home.Get("/", handlers.GetHome)
}
