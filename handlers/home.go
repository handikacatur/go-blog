package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/handikacatur/go-blog/database"
	"github.com/handikacatur/go-blog/models"
)

func GetHome(c *fiber.Ctx) error {
	db := database.DBConn

	tokenString := c.Cookies("token")

	claims := models.CustomClaim{}

	if len(tokenString) > 0 {
		claims, _ = claims.GetClaim(tokenString)
	}

	// var user modelse.User
	type result struct {
		Id        string
		Title     string
		Subtitle  string
		CreatedAt time.Time
		Username  string
		Date      string
	}

	var posts []result

	if err := db.Table("posts").Select("posts.id, posts.title, posts.subtitle, posts.created_at, users.username").Joins("left join users on posts.user_id = users.id").Order("posts.created_at desc").Scan(&posts).Error; err != nil {
		fmt.Println(err)
	}

	for i := range posts {
		posts[i].Date = posts[i].CreatedAt.Format("January 2, 2006")
	}

	return c.Render("index", fiber.Map{
		"username": claims.Username,
		"posts":    posts,
	})
}
