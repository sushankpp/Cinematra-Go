package handlers

import (
	"cinematra-backend/config"
	"cinematra-backend/models"
	"cinematra-backend/utils"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/gofiber/fiber/v3"
)

func Register(c fiber.Ctx) error {
	var body struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		Password  string `json:"password"`
	}

	if err := c.Bind().JSON(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	email := strings.TrimSpace(body.Email)
	password := strings.TrimSpace(body.Password)
	if email == "" || password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email and password are required"})
	}

	var existingUser models.User
	err := config.DB.QueryRow(context.Background(), "select id from users where lower(trim(email)) = lower(trim($1))", email).Scan(&existingUser.ID)
	if err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user with this email already exists"})
	}
	if !isNoRowsErr(err) {
		fmt.Println("Register lookup error for email:", email, "err:", err)
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}
	var user models.User
	err = config.DB.QueryRow(context.Background(),
		"INSERT INTO users (first_name, last_name, email, phone, password, provider) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, first_name, last_name, email, phone, password, picture, provider, created_at, updated_at",
		body.FirstName, body.LastName, email, body.Phone, hash, "email").
		Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Phone, &user.Password, &user.Picture, &user.Provider, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		fmt.Println("Register user insert failed for email:", email, "err:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	isProduction := os.Getenv("ENV") == "production"
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		HTTPOnly: true,
		Secure:   isProduction,
		SameSite: "None",
		MaxAge:   60 * 60 * 24 * 7,
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user": user,
	})
}

func Login(c fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.Bind().JSON(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	email := strings.TrimSpace(body.Email)
	password := strings.TrimSpace(body.Password)
	if email == "" || password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email and password are required"})
	}

	var user models.User
	err := config.DB.QueryRow(context.Background(), "select id, first_name, last_name, email, phone, password, picture, provider, created_at, updated_at from users where lower(trim(email)) = lower(trim($1))", email).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Phone, &user.Password, &user.Picture, &user.Provider, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if isNoRowsErr(err) {
			fmt.Println("Login failed: no user found for email:", email)
		} else {
			fmt.Println("Login DB error for email:", email, "err:", err)
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	if user.Password == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "This account uses Google login"})
	}

	if !utils.CheckPassword(password, *user.Password) {
		fmt.Println("Login failed: password mismatch for email:", email)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	isProduction := os.Getenv("ENV") == "production"
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		HTTPOnly: true,
		Secure:   isProduction,
		SameSite: "None",
		MaxAge:   60 * 60 * 24 * 7,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user": user,
	})
}

func getGoogleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/auth/google/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

func generateRandomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func GoogleLogin(c fiber.Ctx) error {
	state := generateRandomState()
	isProduction := os.Getenv("ENV") == "production"
	c.Cookie(&fiber.Cookie{Name: "oauth_state", Value: state, HTTPOnly: true, Secure: isProduction, SameSite: "None", MaxAge: 86400})
	url := getGoogleOAuthConfig().AuthCodeURL(state)
	return c.Redirect().To(url)
}

func GoogleCallback(c fiber.Ctx) error {
	state := c.Query("state")
	cookieState := c.Cookies("oauth_state")
	if state == "" || state != cookieState {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid State",
		})
	}
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No code provided",
		})
	}
	oauthConfig := getGoogleOAuthConfig()

	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to exchange code for token",
		})
	}

	client := oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user info",
		})
	}
	defer resp.Body.Close()

	var googleUser struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"given_name"`
		LastName  string `json:"family_name"`
		Picture   string `json:"picture"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read user info",
		})
	}

	if err = json.Unmarshal(body, &googleUser); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse user info",
		})
	}
	var user models.User
	googleEmail := strings.TrimSpace(googleUser.Email)
	err = config.DB.QueryRow(context.Background(), "select id, first_name, last_name, email, phone, password, picture, provider, created_at, updated_at from users where lower(trim(email)) = lower(trim($1))", googleEmail).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Phone, &user.Password, &user.Picture, &user.Provider, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if !isNoRowsErr(err) {
			fmt.Println("Google lookup DB error for email:", googleEmail, "err:", err)
		}
		err = config.DB.QueryRow(context.Background(), "Insert into users (first_name, last_name, email, phone, password, picture, provider) values ( $1, $2, $3, $4, $5, $6, $7) returning id, first_name, last_name, email, phone, password, picture, provider, created_at, updated_at", googleUser.FirstName, googleUser.LastName, googleEmail, nil, nil, googleUser.Picture, "google").Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Phone, &user.Password, &user.Picture, &user.Provider, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			fmt.Println("Google user insert failed for email:", googleEmail, "err:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create user",
			})
		}
	}

	jwtToken, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	isProduction := os.Getenv("ENV") == "production"
	c.Cookie(&fiber.Cookie{Name: "oauth_state", Value: "", MaxAge: -1})

	c.Cookie(&fiber.Cookie{Name: "token", Value: jwtToken, HTTPOnly: true, Secure: isProduction, SameSite: "None", MaxAge: 60 * 60 * 24 * 7})

	return c.Redirect().To("http://localhost:3000/auth/callback")

}

func isNoRowsErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "no rows")
}

func Logout(c fiber.Ctx) error {
	isProduction := os.Getenv("ENV") == "production"
	c.Cookie(&fiber.Cookie{Name: "token", Value: "", MaxAge: -1, HTTPOnly: true, Secure: isProduction, SameSite: "None"})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

func Me(c fiber.Ctx) error {
	token := c.Cookies("token")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	userID, err := utils.ValidateJWT(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized user token doesn't match",
		})
	}

	var user models.User
	var onboardingCompleted bool

	query := `
        select u.id, u.first_name, u.last_name, u.email, u.phone, u.password,
               u.picture, u.provider, u.created_at, u.updated_at,
               coalesce(up.onboarding_complete, false) as onboarding_complete
        from users u
        left join user_preferences up on up.user_id = u.id
        where u.id = $1
    `

	err = config.DB.QueryRow(context.Background(), query, userID["user_id"]).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Phone,
		&user.Password, &user.Picture, &user.Provider, &user.CreatedAt, &user.UpdatedAt,
		&onboardingCompleted,
	)
	if err != nil {
		fmt.Println("Me DB error:", err)
		fmt.Println("UserID from token:", userID["user_id"])
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user",
		})
	}

	isProduction := os.Getenv("ENV") == "production"
	c.Cookie(&fiber.Cookie{Name: "token",
		Value:    token,
		HTTPOnly: true,
		Secure:   isProduction,
		SameSite: "None",
		MaxAge:   60 * 60 * 24 * 7,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user":                 user,
		"onboarding_completed": onboardingCompleted,
	})
}

func UpdateProfile(c fiber.Ctx) error {

	token := c.Cookies("token")

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	claims, err := utils.ValidateJWT(token)

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized user token doesn't match",
		})
	}

	userId, ok := claims["user_id"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized user token doesn't match",
		})
	}

	// IMPORTANT: use pointers for partial updates
	var body struct {
		FirstName *string `json:"first_name"`
		LastName  *string `json:"last_name"`
		Phone     *string `json:"phone"`
		Password  *string `json:"password"`
		Picture   *string `json:"picture"`
	}

	if err := c.Bind().JSON(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Handle password safely (only if provided)
	var hashedPassword *string

	if body.Password != nil && *body.Password != "" {

		hash, err := utils.HashPassword(*body.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to hash password",
			})
		}

		hashedPassword = &hash
	}

	// If empty string is passed, treat it as NULL (ignore update)
	if body.FirstName != nil && *body.FirstName == "" {
		body.FirstName = nil
	}
	if body.LastName != nil && *body.LastName == "" {
		body.LastName = nil
	}
	if body.Phone != nil && *body.Phone == "" {
		body.Phone = nil
	}
	if body.Picture != nil && *body.Picture == "" {
		body.Picture = nil
	}

	var user models.User

	err = config.DB.QueryRow(
		context.Background(),
		`
    UPDATE users
    SET
        first_name = COALESCE($1, first_name),
        last_name  = COALESCE($2, last_name),
        phone      = COALESCE($3, phone),
        password   = COALESCE($4, password),
        picture    = COALESCE($5, picture),
		updated_at = NOW()
    WHERE id = $6
    RETURNING
        id,
        first_name,
        last_name,
        email,
        phone,
        password,
        picture,
        provider,
        provider_id,
        created_at,
        updated_at
    `,
		body.FirstName,
		body.LastName,
		body.Phone,
		hashedPassword,
		body.Picture,
		userId,
	).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Phone,
		&user.Password,
		&user.Picture,
		&user.Provider,
		&user.ProviderID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		fmt.Println("UpdateProfile DB error:", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to update user",
			"detail": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user": user,
	})
}
