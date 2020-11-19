package main

import (
	"github.com/alecthomas/kong"
	"github.com/posener/complete"
	"github.com/willabides/kongplete"
	"log"
	"os"
	"runtime"
	"sersh.com/totaltube/frontend/internal"
)

var version = "dev"

var CLI struct {
	Config string `name:"config" short:"c" type:"path" help:"location of totaltube frontend config.toml" env:"TTF_CONFIG_PATH" default:"/etc/totaltube-frontend/config.toml" predictor:"toml"`
	RebuildSass bool ``
}

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
	_, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)
	internal.InitConfig(CLI.Config)
	startServer()
}
