package handlers

import (
	"github.com/flosch/pongo2/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"log"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
	"strconv"
	"strings"
	"time"
)

var ErrNeedCaptcha = errors.New("need captcha")
var ErrCaptchaError = errors.New("captcha error")

func Dmca(c *fiber.Ctx) error {
	path := c.Locals("path").(string)
	config := c.Locals("config").(*site.Config)
	nocache, _ := strconv.ParseBool(c.Query(config.Params.Nocache, "false"))
	langId := c.Locals("lang").(string)
	customContext := generateCustomContext(c, "dmca")
	cacheKey := "dmca:" + langId
	cacheTtl := time.Minute * 15
	isOk := false
	var ip = c.IP()
	db.SessMutex.Lock(ip)
	defer db.SessMutex.Unlock(ip)
	session := db.GetSession(ip)
	defer db.SaveSession(ip, session)
	if session.LastDmca.IsZero() || session.LastDmca.Before(time.Now().Add(-time.Minute)) {
		session.DmcaAmount = 0
		session.LastDmca = time.Now()
	}
	if c.Method() == "POST" {
		session.DmcaAmount++
		params := types.DmcaParams{}
		err := c.BodyParser(&params)
		if err != nil {
			return errors.Wrap(err, "wrong parameters")
		}
		isWhitelisted := false
		for _, e := range internal.Config.Frontend.CaptchaWhiteList {
			if e == params.Email {
				isWhitelisted = true
				break
			}
		}
		if session.DmcaAmount > internal.Config.Frontend.MaxDmcaMinute && !isWhitelisted {
			if params.CaptchaResponse == "" {
				return ErrNeedCaptcha
			}
			response := helpers.Fetch("https://hcaptcha.com/siteverify").
				WithFormData(M{
					"secret":   internal.Config.Frontend.CaptchaSecret,
					"response": params.CaptchaResponse,
				}).Json()
			verifyOk := false
			if success, ok := response["success"].(bool); ok && success {
				if h, ok := response["hostname"].(string);
					ok && strings.TrimPrefix(h, "www.") ==
						strings.TrimPrefix(strings.Split(c.Hostname(), ":")[0], "www.") {
					verifyOk = true
				} else {
					log.Println("wrong hostname for hCaptcha!", strings.Split(c.Hostname(), ":")[0], response["hostname"])
				}
			}
			if !verifyOk {
				return ErrCaptchaError
			}
		}
		err = api.Dmca(params)
		if err != nil {
			return err
		}
		if c.Accepts("text/html") == "" &&
			c.Accepts("application/json") != "" {
			return c.JSON(M{"success": true})
		} else {
			isOk = true
		}
		nocache = true
	}
	renderCaptcha := session.DmcaAmount > internal.Config.Frontend.MaxDmcaMinute
	parsed, err := site.ParseTemplate("dmca", path, config, customContext, nocache, cacheKey, cacheTtl,
		func(ctx pongo2.Context) (pongo2.Context, error) {
			ctx["ok"] = isOk
			ctx["render_captcha"] = renderCaptcha
			return ctx, nil
		})
	if err != nil {
		return err
	}
	c.Set("Content-Type", "text/html")
	return c.Send(parsed)
}
