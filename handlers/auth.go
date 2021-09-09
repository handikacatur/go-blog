package handlers

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/handikacatur/go-blog/database"
	"github.com/handikacatur/go-blog/models"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)

	return string(bytes), err
}

func GetUserByUsername(u string) (*models.User, error) {
	db := database.DBConn

	var user models.User
	if err := db.Where(&models.User{Username: u}).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func getJwtToken(user models.User) fiber.Cookie {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = user.Username
	claims["exp"] = time.Now().Add(time.Hour * 730).Unix()

	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return fiber.Cookie{Name: "jwt", Value: "error"}
	}

	cookie := new(fiber.Cookie)
	cookie.Name = "token"
	cookie.Value = t
	cookie.Expires = time.Now().Add(time.Hour * 24)

	return *cookie
}

func checkCookie(c *fiber.Ctx) bool {
	cookie := c.Cookies("token")

	return cookie != ""
}

func ClearCookie(c *fiber.Ctx) {
	cookie := new(fiber.Cookie)
	cookie.Name = "token"
	cookie.Value = ""

	c.Cookie(cookie)
}

// @desc	Get login page
// @route	GET /auth/login
func GetLogin(c *fiber.Ctx) error {
	tokenString := c.Cookies("token")

	if len(tokenString) > 0 {
		claims := models.CustomClaim{}
		claims, err := claims.GetClaim(tokenString)
		if err == nil {
			return c.Redirect("/")
		}
	}

	return c.Render("login", fiber.Map{})
}

// @desc	Login request
// @route	POST /auth/login
func Login(c *fiber.Ctx) error {
	input := new(models.User)
	inputValues := fiber.Map{}

	// Get request body
	if err := c.BodyParser(input); err != nil {
		return err
	}

	if input.Username == "" || input.Password == "" {
		inputValues["status"] = "error"
		switch {
		case input.Username == "":
			inputValues["username"] = "Please fill this field"
			inputValues["passwordValue"] = input.Password
		case input.Password == "":
			inputValues["password"] = "Please fill this field"
			inputValues["usernameValue"] = input.Username
		}

		return c.Render("login", inputValues)
	}

	user, err := GetUserByUsername(input.Username)
	if err != nil {
		return c.Render("login", fiber.Map{
			"status":        "error",
			"username":      "Username not found",
			"usernameValue": input.Username,
		})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return c.Render("login", fiber.Map{
			"status":        "error",
			"password":      "Wrong password",
			"usernameValue": input.Username,
		})
	}

	cookie := getJwtToken(*user)

	c.Cookie(&cookie)

	// authBearer := "Bearer " + cookie.Value

	c.Set("Authorization", cookie.Value)
	return c.Status(200).Redirect("/")
}

// @desc 	Logout request
// @route	GET /auth/logout
func Logout(c *fiber.Ctx) error {
	ClearCookie(c)

	return c.Redirect("/auth/login")
}

// @desc	Get register page
// @route	GET /auth/sign-up
func GetSignup(c *fiber.Ctx) error {
	if checkCookie(c) {
		return c.Redirect("/")
	}

	return c.Render("sign-up", fiber.Map{
		"success": true,
	})
}

// @desc 	Register user
// @route	POST /auth/sign-up
func CreateUser(c *fiber.Ctx) error {
	db := database.DBConn

	type inputField struct {
		Username        string `form:"username"`
		Password        string `form:"password"`
		ConfirmPassword string `form:"confirm-password"`
	}

	var input inputField
	user := new(models.User)
	inputValues := fiber.Map{}

	if err := c.BodyParser(&input); err != nil {
		return err
	}

	if input.Username == "" || input.Password == "" || input.ConfirmPassword == "" {
		inputValues["status"] = "error"
		switch {
		case input.Username == "":
			inputValues["username"] = "Please fill this field"
			inputValues["passwordValue"] = input.Password
		case input.Password == "" || input.ConfirmPassword == "":
			inputValues["password"] = "Please fill this field"
			inputValues["usernameValue"] = input.Username
		}

		return c.Render("sign-up", inputValues)
	}

	if input.Password != input.ConfirmPassword {
		return c.Render("sign-up", fiber.Map{
			"status":        "error",
			"password":      "Password didn't match",
			"usernameValue": input.Username,
		})
	}

	_, err := GetUserByUsername(input.Username)
	if err == nil {
		return c.Render("sign-up", fiber.Map{
			"status":   "error",
			"username": "Username already registered",
		})
	}

	hash, err := hashPassword(input.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Couldn't hash password", "data": err})
	}

	user.Username = input.Username
	user.Password = input.Password

	user.Password = hash

	db.Create(user)

	cookie := getJwtToken(*user)

	c.Cookie(&cookie)

	return c.Status(200).Redirect("/")
}
