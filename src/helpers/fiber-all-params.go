package helpers

import (
	"github.com/gofiber/fiber/v2"
)

func FiberAllParams(c *fiber.Ctx) map[string]string {
	var res = make(map[string]string)
	for _, key := range c.Route().Params {
		res[key] = c.Params(key)
	}
	return res
}

func FiberAllQuery(c *fiber.Ctx) map[string]string {
	var res = make(map[string]string)
	c.Context().QueryArgs().VisitAll(func(key, value []byte) {
		res[string(key)] = string(value)
	})
	return res
}