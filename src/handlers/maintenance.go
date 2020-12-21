package handlers

import (
	"github.com/gofiber/fiber/v2"
	"path/filepath"
	"sersh.com/totaltube/frontend/site"
)

func MaintenanceRebuildJS(c *fiber.Ctx) error {
	config := c.Locals("config").(*site.Config)
	path := c.Locals("path").(string)
	err := site.RebuildJS(filepath.Join(path, "js"), config)
	if err != nil {
		return err
	}
	return c.JSON(M{"success": true})
}

func MaintenanceRebuildSCSS(c *fiber.Ctx) error {
	config := c.Locals("config").(*site.Config)
	path := c.Locals("path").(string)
	err := site.RebuildSCSS(filepath.Join(path, "scss"), config)
	if err != nil {
		return err
	}
	return c.JSON(M{"success": true})
}