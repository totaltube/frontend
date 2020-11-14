package main

import (
	"github.com/flosch/pongo2/v4"
	"github.com/gofiber/fiber/v2"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
)

func newHandler(c *fiber.Ctx) error {
	return c.SendString("new")
}
func autocompleteHandler(c *fiber.Ctx) error {
	return c.SendString("autocomplete")
}
func searchHandler(c *fiber.Ctx) error {
	return c.SendString("search")
}
func categoryHandler(c *fiber.Ctx) error {
	return c.SendString("category")
}

func topCategoriesHandler(c *fiber.Ctx) error {
	path := c.Locals("path").(string)
	config := c.Locals("config").(*site.Config)
	langCookie := c.Cookies(internal.Config.General.LangCookie)
	lang := internal.DetectLanguage(langCookie, c.Get("Accept-Language"))
	customContext := pongo2.Context{
		"lang": lang.Name,
	}
	parsed, err := site.ParseTemplate("top-categories", path, config, customContext)
	if err != nil {
		return err
	}
	return c.Send(parsed)
}

func topContentHandler(c *fiber.Ctx) error {
	return c.SendString("top_content")
}

func modelHandler(c *fiber.Ctx) error {
	return c.SendString("model")
}
func channelHandler(c *fiber.Ctx) error {
	return c.SendString("channel")
}
func contentHandler(c *fiber.Ctx) error {
	return c.SendString("content")
}
func longHandler(c *fiber.Ctx) error {
	return c.SendString("long")
}
func modelsHandler(c *fiber.Ctx) error {
	return c.SendString("models")
}
func popularHandler(c *fiber.Ctx) error {
	return c.SendString("popular")
}
func outHandler(c *fiber.Ctx) error {
	return c.SendString("out")
}

func customHandler(c *fiber.Ctx) error {
	// Здесь мы выполним какой-нибудь хитрый пользовательский js, который получит нужные данные,
	// а потом загрузит шаблон этого хандлера
	return c.SendString("custom")
}
