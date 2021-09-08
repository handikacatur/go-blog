package handlers

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/handikacatur/go-blog/database"
	"github.com/handikacatur/go-blog/models"
	"github.com/handikacatur/go-blog/utils"
)

func GetMyPosts(c *fiber.Ctx) error {
	db := database.DBConn

	tokenString := c.Cookies("token")

	claims := models.CustomClaim{}

	var err error

	if len(tokenString) > 0 {
		claims, err = claims.GetClaim(tokenString)

		if err != nil {
			ClearCookie(c)
			return c.Redirect("/auth/login")
		}
	}

	var user models.User
	if err := db.Where(&models.User{Username: claims.Username}).Preload("Posts").Find(&user).Error; err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	type PostData struct {
		Id       uint
		Title    string
		Subtitle string
		Date     string
	}

	postData := make([]PostData, len(user.Posts))

	for i := range postData {
		t := user.Posts[i].CreatedAt
		dt := t.Format("January 2, 2006")
		postData[i].Id = user.Posts[i].ID
		postData[i].Title = user.Posts[i].Title
		postData[i].Subtitle = user.Posts[i].Subtitle
		postData[i].Date = dt
	}

	return c.Render("my_post", fiber.Map{
		"username": claims.Username,
		"posts":    postData,
	})
}

func GetPost(c *fiber.Ctx) error {
	db := database.DBConn

	tokenString := c.Cookies("token")

	claims := models.CustomClaim{}

	claims, err := claims.GetClaim(tokenString)
	if err != nil {
		ClearCookie(c)
		return c.Redirect("/auth/login")
	}

	// Get parameter
	reqId := c.Params("id")

	var post models.Post
	if err := db.First(&post, reqId).Error; err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	// Change date format to January 2, 2006
	t := post.CreatedAt
	formatDate := t.Format("January 2, 2006")

	return c.Render("post", fiber.Map{
		"title":    post.Title,
		"subtitle": post.Subtitle,
		"cover":    post.Cover,
		"author":   claims.Username,
		"date":     formatDate,
		"data":     template.HTML(post.Data),
	})
}

func GetCreatePost(c *fiber.Ctx) error {

	return c.Render("create_post", fiber.Map{
		"title": "Hello world",
	})
}

func CreatePost(c *fiber.Ctx) error {
	tokenString := c.Cookies("token")

	claims := models.CustomClaim{}

	var err error
	if len(tokenString) > 0 {
		claims, err = claims.GetClaim(tokenString)
		if err != nil {
			ClearCookie(c)
			return c.Redirect("/auth/login")
		}
	}

	var input models.Post

	if err := c.BodyParser(&input); err != nil {
		return err
	}

	file, err := c.FormFile("cover")
	if err != nil {
		return c.SendStatus(fiber.ErrBadRequest.Code)
	}

	// Check file content type
	if !strings.Contains(file.Header["Content-Type"][0], "image/") {
		return c.SendStatus(fiber.ErrBadRequest.Code)
	}

	user, _ := GetUserByUsername(claims.Username)

	// Rename and upload image file
	date := strings.ReplaceAll(time.Now().Format("01-02-20006"), "-", "")
	ids := fmt.Sprintf("%d%d", user.ID, time.Now().Unix())
	file.Filename = fmt.Sprintf("%s-%s-%s.%s", date, ids, user.Username, file.Header["Content-Type"][0][6:])

	// Save file to /public/assets/img/covers
	if err := c.SaveFile(file, fmt.Sprintf("./public/assets/img/covers/%s", file.Filename)); err != nil {
		return err
	}

	// Convert clean data to HMTL
	input.Data = utils.HTML(string(input.Data))

	post := new(models.Post)
	post.Title = input.Title
	post.Subtitle = input.Subtitle
	post.Cover = file.Filename
	post.Data = input.Data
	post.UserID = user.ID

	db := database.DBConn

	db.Create(post)
	return c.Redirect("/")
}
