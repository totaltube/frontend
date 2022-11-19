package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/alecthomas/kong"
	"github.com/pkg/errors"
	"github.com/posener/complete"
	"github.com/willabides/kongplete"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/internal"
)

var version = "dev"

func main() {
	if runtime.GOOS == "windows" {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
	} else {
		log.SetFlags(log.Lshortfile)
	}
	internal.Version = version
	parser := kong.Must(&CLI,
		kong.Name("totaltube"),
		kong.Description("Total tube version "+version),
		kong.UsageOnError(),
	)
	kongplete.Complete(parser,
		kongplete.WithPredictor("toml", complete.PredictFiles("*.toml")),
	)
	ctx, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)
	switch ctx.Command() {
	case "start":
		internal.InitConfig(CLI.Config)
		internal.Config.Options, err = api.Options(internal.Config.Frontend.DefaultSite)
		if err != nil {
			panic(errors.Wrap(err, "Can't get sites options"))
		}
		upgradeConfig()
		startServer()
	case "install":
		Install()
	default:
		fmt.Println("unknown command")
	}
}
