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

	var user models.User
	if err := db.Where(&models.User{Username: claims.Username}).Preload("Posts").Find(&user).Error; err != nil {
		return c.Redirect("/post/my-post/create")
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
	var user models.User
	if err := db.First(&post, reqId).Error; err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	// if err := db.Model(&user).Where("Username = ", post.UserID).Association("Posts").Find(&post); err != nil {
	// 	fmt.Println(err)
	// }

	if err := db.Select("Username").Where("ID = ?", post.UserID).Find(&user).Error; err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	var authorized bool
	if user.Username != claims.Username {
		authorized = true
	}

	// Change date format to January 2, 2006
	t := post.CreatedAt
	formatDate := t.Format("January 2, 2006")

	return c.Render("post", fiber.Map{
		"id":         post.ID,
		"title":      post.Title,
		"subtitle":   post.Subtitle,
		"cover":      post.Cover,
		"author":     claims.Username,
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

	var post models.Post
	var user models.User
	if err := db.Where("ID = ?", reqId).First(&post).Error; err != nil {
		fmt.Println(err)
	}

	if err := db.Where("ID = ?", post.UserID).First(&user).Error; err != nil {
		fmt.Println(err)
	}

	if claims.Username != user.Username {
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

	type Input struct {
		Id       uint
		Title    string
		Subtitle string
		Cover    string
		Data     string
	}

	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Redirect("/")
	}

	input.Data = utils.HTML(input.Data)

	var post models.Post
	var user models.User
	if err := db.Where("ID = ?", input.Id).First(&post).Error; err != nil {
		fmt.Println(err)
	}
	// Check if user authorized
	if err := db.Where("ID = ?", post.UserID).First(&user).Error; err != nil {
		fmt.Println(err)
	}
	if user.Username != claims.Username {
		fmt.Println("executed2")
		return c.Redirect(fmt.Sprintf("/post/show/%d", input.Id))
	}
	if err := db.Model(&post).Where("ID = ?", input.Id).Updates(input).Error; err != nil {
		fmt.Println(err)
	}

	return c.Render("index", fiber.Map{})
}
