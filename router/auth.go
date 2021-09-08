package router

import (
	"github.com/handikacatur/go-blog/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetAuth(app *fiber.App) {
	auth := app.Group("/auth")

	// Login routes
	auth.Get("/login", handlers.GetLogin)
	auth.Post("/login", handlers.Login)

	// Sign-up routes
	auth.Get("/sign-up", handlers.GetSignup)
	auth.Post("/sign-up", handlers.CreateUser)

	// Logout route
	auth.Get("/logout", handlers.Logout)
}
