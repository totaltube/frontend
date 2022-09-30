package handlers

import "github.com/gofiber/fiber/v2"

func Handle404(c *fiber.Ctx) error {
	return Generate404(c, "page not found")
}
