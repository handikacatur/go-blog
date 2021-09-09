package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/handikacatur/go-blog/handlers"
	"github.com/handikacatur/go-blog/middleware"
)

func SetPost(app *fiber.App) {
	post := app.Group("/post")

	post.Get("/show/:id", middleware.Protected(), handlers.GetPost)
	post.Get("/my-post", middleware.Protected(), handlers.GetMyPosts)
	post.Post("/my-post", middleware.Protected(), handlers.CreatePost)
	post.Put("/my-post/", middleware.Protected(), handlers.UpdatePost)
	post.Get("/my-post/create", middleware.Protected(), handlers.GetCreatePost)
	post.Get("/my-post/:id/edit", middleware.Protected(), handlers.GetUpdatePost)
}
