package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// Список публичных маршрутов, доступных без авторизации
var publicRoutes = []string{
	"/health",
	"/auth/register",
	"/auth/verify-registration",
	"/auth/login",
	"/auth/password-reset",
	"/auth/verify-reset-code",
	"/auth/reset-password",
}

// Проверяет, является ли маршрут публичным
func isPublicRoute(path string) bool {
	log.Printf("Checking path: %s", path)
	for _, route := range publicRoutes {
		if strings.HasPrefix(path, route) {
			log.Printf("Path %s matches public route %s", path, route)
			return true
		}
	}
	log.Printf("Path %s is not public", path)
	return false
}

// Проверить наличие токена
func jwtMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		log.Printf("Request path: %s", path)

		// Пропускаем проверку для публичных маршрутов
		if isPublicRoute(path) {
			return c.Next()
		}

		// Получаем JWT из заголовка Authorization
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			jwtSecret := getEnv("JWT_SECRET")
			return []byte(jwtSecret), nil
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		if !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		if userId, ok := claims["sub"].(string); ok {
			c.Locals("userId", userId)
		}

		return c.Next()
	}
}
