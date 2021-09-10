package main

import (
	"fmt"
	"log"
	"os"

	"github.com/handikacatur/go-blog/database"
	"github.com/handikacatur/go-blog/models"
	"github.com/handikacatur/go-blog/router"
	"github.com/joho/godotenv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func initDB() {
	var err error
	database.DBConn, err = gorm.Open("sqlite3", "go_blog.db")
	if err != nil {
		panic("failed to connect database")
	}
	fmt.Println("Connection Opened to Database")
	database.DBConn.AutoMigrate(&models.Post{})
	database.DBConn.AutoMigrate(&models.User{})
	fmt.Println("Database Migrated")
}

func main() {
	// Create covers directory if not exist
	_, err := os.Stat("./public/assets/img/covers")
	if os.IsNotExist(err) {
		os.Mkdir("./public/assets/img/covers", 0775)
	}

	// Instantiate html template engine
	engine := html.New("./views", ".html")

	// Load env vars
	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Can not load env variables")
	}

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	initDB()

	app.Use(func(c *fiber.Ctx) error {
		token := "Bearer " + c.Cookies("token")
		c.Request().Header.Add("Authorization", token)
		return c.Next()
	})

	// Serve static
	app.Static("/", "./public")

	// Instantiate routers
	router.SetHome(app)
	router.SetAuth(app)
	router.SetPost(app)

	app.Use(func(c *fiber.Ctx) error {
		return c.Render("404", fiber.Map{})
	})

	log.Fatal(app.Listen(":" + os.Getenv("PORT")))
}
