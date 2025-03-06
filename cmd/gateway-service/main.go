// @title API Gateway
// @version 1.0
// @description API Gateway
// @host localhost:8080
// @BasePath /
// @schemes http
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/GlebPoroshin/geochat-gateway-service/cmd/gateway-service/docs"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/swagger"
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

	// Gateway Swagger UI
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Специальный маршрут для Swagger auth-service
	// Важно: этот маршрут должен быть определен ДО общего маршрута /auth/*
	app.All("/auth/swagger/*", func(context *fiber.Ctx) error {
		path := context.Params("*")
		log.Printf("Request path: %s", context.Path())
		log.Printf("Checking path: %s", context.Path())
		
		// Проверяем, является ли запрос запросом к Swagger
		if strings.HasPrefix(context.Path(), "/auth/swagger/") {
			log.Printf("Path %s is a Swagger path, allowing access", context.Path())
			
			// Если запрашивается index.html, отправляем модифицированный HTML с правильным путем к doc.json
			if path == "index.html" || path == "" || path == "/" {
				// Проксируем запрос к auth-service для получения index.html
				originalURL := fmt.Sprintf("%s/swagger/index.html", authServiceURL)
				
				// Получаем содержимое через прокси
				if err := proxy.Do(context, originalURL); err != nil {
					return err
				}
				
				// Модифицируем URL для doc.json, чтобы он указывал на правильный путь
				body := string(context.Response().Body())
				body = strings.Replace(body, `"url":"/swagger/doc.json"`, `"url":"/auth/swagger/doc.json"`, 1)
				
				// Устанавливаем модифицированное содержимое
				context.Response().SetBody([]byte(body))
				return nil
			}
			
			// Для других файлов Swagger (doc.json, css, js и т.д.)
			log.Printf("Proxying Swagger request to auth service: %s", path)
			return proxy.Do(context, fmt.Sprintf("%s/swagger/%s", authServiceURL, path))
		}
		
		return context.Next()
	})

	// Пример маршрута с аннотациями Swagger:
	// Proxy запроса к Auth Service
	// @Summary Проксирование запроса к Auth Service
	// @Description Проксирует запрос, начинающийся с /auth, в Auth Service
	// @Tags gateway
	// @Accept json
	// @Produce json
	// @Param path path string true "Путь запроса"
	// @Success 200 {object} map[string]interface{} "Успешный ответ"
	// @Failure 400 {object} map[string]interface{} "Ошибка в запросе"
	// @Router /auth/{path} [get]
	// @Router /auth/{path} [post]
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
