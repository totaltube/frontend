package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jpillora/overseer"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func server(state overseer.State) {
	app := fiber.New(fiber.Config{
		CaseSensitive:         true,
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if c.Accepts("json") != "" {
				err = c.JSON(map[string]interface{}{
					"success": false,
					"value":   err.Error(),
				})
				if err == nil {
					return err
				}
			}
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		},
	})
	app.Use(recover.New())
	go func() {
		if runtime.GOOS == "windows" {
			fmt.Println(app.Listen(fmt.Sprintf(":%d", serverPort)))
		} else {
			fmt.Println(app.Listener(state.Listener))
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGABRT)
	signal.Notify(c, syscall.SIGKILL)
	select {
	case <-c:
	case <-state.GracefulShutdown:
	}
	// Здесь мы после завершения программы
	log.Println("Making some cleanup before exit...")
}
