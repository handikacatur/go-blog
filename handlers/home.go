package handlers

import (
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

	var post models.Post
	if err := db.Find(&post).Error; err != nil {
		return err
	}

	var user models.User
	if err := db.Where(&models.User{Username: claims.Username}).Preload("Posts").Find(&user).Error; err != nil {
		return c.Redirect("/post/my-post/create")
	}

	// type PostData struct {
	// 	Id       uint
	// 	Title    string
	// 	Subtitle string
	// 	Date     string
	// }

	// postData := make([]PostData, len(post.))

	// for i := range postData {
	// 	t := user.Posts[i].CreatedAt
	// 	dt := t.Format("January 2, 2006")
	// 	postData[i].Id = user.Posts[i].ID
	// 	postData[i].Title = user.Posts[i].Title
	// 	postData[i].Subtitle = user.Posts[i].Subtitle
	// 	postData[i].Date = dt
	// }

	// fmt.Println(len(post))

	return c.Render("index", fiber.Map{
		"username": claims.Username,
		"posts":    "hello wor",
	})
}
