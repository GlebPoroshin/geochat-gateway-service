package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

var (
	authServiceURL         = getEnv("AUTH_SERVICE_URL")
	userServiceURL         = getEnv("USER_SERVICE_URL")
	chatServiceURL         = getEnv("CHAT_SERVICE_URL")
	eventServiceURL        = getEnv("EVENT_SERVICE_URL")
	notificationServiceURL = getEnv("NOTIFICATION_SERVICE_URL")
	presenceServiceURL     = getEnv("PRESENCE_SERVICE_URL")
)

func main() {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(context *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error

			if errors.As(err, &e) {
				code = e.Code
			}

			return context.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	app.Use(jwtMiddleware())

	app.Get("/health", func(context *fiber.Ctx) error {
		return context.JSON(fiber.Map{
			"status": "ok",
		})
	})

	// Auth Service Routes
	app.All("/auth/*", func(context *fiber.Ctx) error {
		path := context.Params("*")
		return proxy.Do(context, fmt.Sprintf("%s/auth/%s", authServiceURL, path))
	})

	// User Service Routes
	app.All("/users/*", func(context *fiber.Ctx) error {
		path := context.Params("*")
		setUserId(context)
		return proxy.Do(context, fmt.Sprintf("%s/users/%s", userServiceURL, path))
	})

	app.All("/chats/*", func(context *fiber.Ctx) error {
		path := context.Params("*")
		setUserId(context)
		return proxy.Do(context, fmt.Sprintf("%s/chats/%s", chatServiceURL, path))
	})

	app.All("/events/*", func(context *fiber.Ctx) error {
		path := context.Params("*")
		setUserId(context)
		return proxy.Do(context, fmt.Sprintf("%s/events/%s", eventServiceURL, path))
	})

	app.All("/notifications/*", func(context *fiber.Ctx) error {
		path := context.Params("*")
		setUserId(context)
		return proxy.Do(context, fmt.Sprintf("%s/notifications/%s", notificationServiceURL, path))
	})

	app.All("/presence/*", func(context *fiber.Ctx) error {
		path := context.Params("*")
		setUserId(context)
		return proxy.Do(context, fmt.Sprintf("%s/presence/%s", presenceServiceURL, path))
	})

	port := getEnv("PORT")
	log.Printf("Gateway service starting on port %s", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
}

func getEnv(key string) string {
	return os.Getenv(key)
}

func setUserId(c *fiber.Ctx) {
	if userId := c.Locals("userId"); userId != nil {
		c.Request().Header.Set("User-Id", userId.(string))
	}
}
