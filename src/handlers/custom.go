package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"sersh.com/totaltube/frontend/site"
)

func Custom(c *fiber.Ctx) error {
	path := c.Locals("path").(string)
	config := c.Locals("config").(*site.Config)
	nocache, _ := strconv.ParseBool(c.Query(config.Params.Nocache, "false"))
	templateName := c.Locals("custom_template_name").(string)
	page, _ := strconv.ParseInt(c.Params("page", c.Query(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	customContext := generateCustomContext(c, "custom/"+templateName)
	customContext["page"] = page
	parsed, err := site.ParseCustomTemplate(templateName, path, config, customContext, nocache, c)
	if err != nil {
		if err1, ok := err.(site.ErrSendResponse); ok {
			if err1.Redirect != "" {
				if err1.RedirectCode > 0 {
					return c.Redirect(err1.Redirect, err1.RedirectCode)
				} else {
					return c.Redirect(err1.Redirect)
				}
			}
			if err1.JSON != nil {
				c.Set("Content-Type", "application/json")
				return c.Send(err1.JSON)
			}
			if err1.Text != nil {
				c.Set("Content-Type", "text/html")
				return c.Send(err1.Text)
			}
		}
		return err
	}
	c.Set("Content-Type", "text/html")
	return c.Send(parsed)
}

