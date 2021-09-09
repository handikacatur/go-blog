package handlers

import (
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/handikacatur/go-blog/database"
	"github.com/handikacatur/go-blog/models"
	"github.com/handikacatur/go-blog/utils"
)

type Data struct {
	Text string
	File string
}

type Block struct {
	Type string
	Data Data
}

type CleanData struct {
	Blocks []Block
}

func htmlToClean(data string) CleanData {
	splitData := strings.Split(data, "\n")
	cleanData := new(CleanData)
	for _, data := range splitData {
		if strings.Contains(data, "<p>") {
			newData := Data{Text: data[3 : len(data)-4]}
			newBlock := Block{Type: "paragraph", Data: newData}
			cleanData.Blocks = append(cleanData.Blocks, newBlock)
		} else if strings.Contains(data, "<img") {
			src := strings.Split(data, " ")[1]
			newData := Data{File: src[5 : len(src)-1]}
			newBlock := Block{Type: "image", Data: newData}
			cleanData.Blocks = append(cleanData.Blocks, newBlock)
		}
	}

	return *cleanData
}

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

	type result struct {
		ID        string
		Title     string
		Subtitle  string
		CreatedAt time.Time
		Username  string
		Date      string
	}
	var posts []result
	if err := db.Table("posts").Select("posts.id, posts.title, posts.subtitle, posts.created_at, users.username").Joins("left join users on posts.user_id = users.id").Where("users.username = ?", claims.Username).Scan(&posts).Error; err != nil {
		return err
	}

	for i := range posts {
		posts[i].Date = posts[i].CreatedAt.Format("January 2, 2006")
	}

	return c.Render("my_post", fiber.Map{
		"username": claims.Username,
		"posts":    posts,
	})
}

func GetPost(c *fiber.Ctx) error {
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

	// Get parameter
	reqId := c.Params("id")

	type result struct {
		ID        string
		Title     string
		Subtitle  string
		Cover     string
		Username  string
		CreatedAt time.Time
		Data      string
	}
	var post result
	err = db.Table("posts").Select("posts.id, posts.title, posts.subtitle, posts.cover, users.username, posts.created_at, posts.data").Joins("left join users on posts.user_id = users.id").First(&post, reqId).Error
	if err != nil {
		return err
	}

	// Check if the user is authorized
	var authorized bool
	if post.Username == claims.Username {
		authorized = true
	}

	// Change date format to January 2, 2006
	formatDate := post.CreatedAt.Format("January 2, 2006")

	return c.Render("post", fiber.Map{
		"id":         post.ID,
		"title":      post.Title,
		"subtitle":   post.Subtitle,
		"cover":      post.Cover,
		"author":     post.Username,
		"date":       formatDate,
		"data":       template.HTML(post.Data),
		"authorized": authorized,
	})
}

func GetCreatePost(c *fiber.Ctx) error {
	return c.Render("create_post", fiber.Map{})
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

func GetUpdatePost(c *fiber.Ctx) error {
	db := database.DBConn

	tokenString := c.Cookies("token")

	claims := models.CustomClaim{}

	claims, err := claims.GetClaim(tokenString)
	if err != nil {
		ClearCookie(c)
		return c.Redirect("/auth/login")
	}

	// Get request parameter
	reqId := c.Params("id")

	type result struct {
		ID       string
		Title    string
		Subtitle string
		Cover    string
		Data     string
		Username string
	}
	var post result
	if err := db.Table("posts").Select("posts.id, posts.title, posts.cover, posts.data, users.username").Joins("left join users on posts.user_id = users.id").Where("posts.id = ?", reqId).First(&post).Error; err != nil {
		return err
	}

	if claims.Username != post.Username {
		return c.Redirect(fmt.Sprintf("/post/show/%s", reqId))
	}

	cleanData := htmlToClean(post.Data)

	return c.Render("create_post", fiber.Map{
		"id":       post.ID,
		"title":    post.Title,
		"subtitle": post.Subtitle,
		"cover":    post.Cover,
		"data":     cleanData,
	})
}

func UpdatePost(c *fiber.Ctx) error {
	db := database.DBConn

	tokenString := c.Cookies("token")

	claims := models.CustomClaim{}

	claims, err := claims.GetClaim(tokenString)
	if err != nil {
		ClearCookie(c)
		return c.Redirect("/auth/login")
	}

	file, _ := c.FormFile("cover")

	type Input struct {
		ID       uint
		Title    string
		Subtitle string
		Cover    string
		Data     string
	}
	var input Input

	if err := c.BodyParser(&input); err != nil {
		return err
	}

	input.Data = utils.HTML(input.Data)

	// Get id and username from requested post
	type result struct {
		ID       string
		Username string
		Cover    string
	}
	var post result
	if err := db.Table("posts").Select("posts.id, users.username, posts.cover").Joins("left join users on posts.user_id = users.id").Where("posts.id = ?", input.ID).First(&post).Error; err != nil {
		return err
	}

	// Check file content type
	if file != nil {
		if strings.Contains(file.Header["Content-Type"][0], "image/") {
			user, _ := GetUserByUsername(claims.Username)

			// Rename and upload image file
			date := strings.ReplaceAll(time.Now().Format("01-02-20006"), "-", "")
			ids := fmt.Sprintf("%d%d", user.ID, time.Now().Unix())
			file.Filename = fmt.Sprintf("%s-%s-%s.%s", date, ids, user.Username, file.Header["Content-Type"][0][6:])
			// Delete last cover
			os.Remove(fmt.Sprintf("./public/assets/img/covers/%s", post.Cover))
			// Save file to /public/assets/img/covers
			if err := c.SaveFile(file, fmt.Sprintf("./public/assets/img/covers/%s", file.Filename)); err != nil {
				return err
			}
			input.Cover = file.Filename
		}
	}

	// Check if client is authorized
	if post.Username != claims.Username {
		return c.Redirect(fmt.Sprintf("/post/show/%s", post.ID))
	}

	// Update the post
	if err := db.Model(&models.Post{}).Where("ID = ?", post.ID).Updates(input).Error; err != nil {
		return err
	}

	return c.SendStatus(200)
}

func DeletePost(c *fiber.Ctx) error {
	db := database.DBConn

	tokenString := c.Cookies("token")

	claims := models.CustomClaim{}

	claims, err := claims.GetClaim(tokenString)
	if err != nil {
		ClearCookie(c)
		return c.Redirect("/auth/login")
	}

	reqId := c.Params("id")

	type result struct {
		ID       string
		Username string
		Cover    string
	}
	var post result
	if err := db.Table("posts").Select("posts.id, users.username, posts.cover").Joins("left join users on posts.user_id = users.id").Where("posts.id = ?", reqId).Scan(&post).Error; err != nil {
		return err
	}

	if post.Username != claims.Username {
		fmt.Println("executed")
		return c.Redirect(fmt.Sprintf("/post/show/%s", reqId))
	}

	db.Unscoped().Delete(&models.Post{}, reqId)
	os.Remove(fmt.Sprintf("./public/assets/img/covers/%s", post.Cover))

	return c.Redirect("/post/my-post")
}
