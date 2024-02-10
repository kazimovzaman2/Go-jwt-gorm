package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kazimovzaman2/Go-chatapp/config"
	"github.com/kazimovzaman2/Go-chatapp/database"
	"github.com/kazimovzaman2/Go-chatapp/model"
	"github.com/kazimovzaman2/Go-chatapp/utils"
	"golang.org/x/crypto/bcrypt"
)

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateTokens(user model.User) (string, string, error) {
	accessToken, err := utils.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := utils.GenerateRefreshToken(user)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func Login(c *fiber.Ctx) error {
	db := database.DB
	var user model.User

	var input model.LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Status:  "error",
			Message: "Invalid input",
			Errors:  err.Error(),
		})
	}

	if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
			Status:  "error",
			Message: "User not found",
			Errors:  err.Error(),
		})
	}

	if !CheckPasswordHash(input.Password, user.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Status:  "error",
			Message: "The password is incorrect",
			Errors:  "Invalid password",
		})
	}

	accessToken, refreshToken, err := generateTokens(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Status:  "error",
			Message: "Could not login",
			Errors:  err.Error(),
		})
	}

	successResp := model.SuccessResponse{
		Status:  "success",
		Message: "Logged in",
		Data: fiber.Map{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	}

	return c.Status(fiber.StatusOK).JSON(successResp)
}

func RefreshToken(c *fiber.Ctx) error {
	type RefreshTokenInput struct {
		RefreshToken string `json:"refresh_token"`
	}

	var input RefreshTokenInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Status:  "error",
			Message: "Invalid input",
			Errors:  err.Error(),
		})
	}

	token, err := jwt.Parse(input.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWTSecret), nil
	})
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Status:  "error",
			Message: "Invalid token",
			Errors:  err.Error(),
		})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Status:  "error",
			Message: "The token is invalid",
			Errors:  "Invalid token",
		})
	}

	db := database.DB
	userID := claims["id"].(float64)
	var user model.User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
			Status:  "error",
			Message: "User not found",
			Errors:  err.Error(),
		})
	}

	accessToken, refreshToken, err := generateTokens(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Status:  "error",
			Message: "Could not refresh token",
			Errors:  err.Error(),
		})
	}

	successResp := model.SuccessResponse{
		Status:  "success",
		Message: "Token refreshed",
		Data: fiber.Map{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	}

	return c.Status(fiber.StatusOK).JSON(successResp)
}