package handlers

import (
	"github.com/gofiber/fiber/v2"
	"sersh.com/totaltube/frontend/site"
	"strconv"
)

func Custom(c *fiber.Ctx) error {
	path := c.Locals("path").(string)
	config := c.Locals("config").(*site.Config)
	nocache, _ := strconv.ParseBool(c.Query(config.Params.Nocache, "false"))
	templateName := c.Locals("custom_template_name").(string)
	//langId := c.Locals("lang").(string)
	page, _ := strconv.ParseInt(c.Params("page", c.Query(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	customContext := generateCustomContext(c, "custom/"+templateName)
	customContext["page"] = page
	parsed, err := site.ParseCustomTemplate(templateName, path, config, customContext, nocache)
	if err != nil {
		return err
	}
	c.Set("Content-Type", "text/html")
	return c.Send(parsed)
}

