package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/kazimovzaman2/Go-chatapp/config"
	"github.com/kazimovzaman2/Go-chatapp/handler"
	"github.com/kazimovzaman2/Go-chatapp/middleware"
)

func SetupRoutes(app *fiber.App) {
	protected := middleware.NewAuthMiddleware(config.JWTSecret)

	api := app.Group("/api", logger.New())
	api.Get("/hello/", handler.Hello)

	auth := api.Group("/auth")
	auth.Post("/login/", handler.Login)

	users := api.Group("/users")
	users.Get("/", handler.GetAllUsers)
	users.Post("/", handler.CreateUser)
	users.Get("/me/", protected, handler.GetMe)
	users.Get("/:id/", handler.GetUser)
}
