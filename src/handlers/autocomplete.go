package handlers

import (
	"github.com/gofiber/fiber/v2"
	"log"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/site"
)

func Autocomplete(c *fiber.Ctx) error {
	config := c.Locals("config").(*site.Config)
	langId := c.Locals("lang").(string)
	hostName := c.Locals("hostName").(string)
	searchQuery := c.Query(config.Params.SearchQuery)
	results, err := api.Autocomplete(hostName, searchQuery, langId)
	if err != nil {
		log.Println("Error querying autocomplete api:", err)
		return c.JSON(A{})
	}
	return c.JSON(results.Items)
}
