package api

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"sersh.com/totaltube/frontend/types"
)

func UpdateConfig(config *types.Config, configSource string) (err error) {
	if os.Getenv("DEBUG") == "true" {
		return nil
	}
	_, err = Request(config.Hostname, methodPost, uriUpdateConfig, M{"config": configSource})
	return
}

func UpdateConfigRetry(config *types.Config, configSource string) (err error) {
	for range 10 {
		if err = UpdateConfig(config, configSource); err == nil {
			return nil
		}
		if strings.Contains(err.Error(), "not found") {
			break
		}
		log.Println("failed to update config, retrying...", err)
		time.Sleep(time.Second * 3)
	}
	return errors.New("failed to update config after 10 retries")
}
