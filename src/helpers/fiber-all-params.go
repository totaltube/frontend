package helpers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

func FiberAllParams(c *fiber.Ctx) map[string]string {
	var res = make(map[string]string)
	for _, key := range c.Route().Params {
		res[key] = utils.ImmutableString(c.Params(key))
	}
	return res
}

func FiberAllQuery(c *fiber.Ctx) map[string]string {
	var res = make(map[string]string)
	c.Context().QueryArgs().VisitAll(func(key, value []byte) {
		k := utils.ImmutableString(string(key))
		v := utils.ImmutableString(string(value))
		res[k] = v
	})
	return res
}