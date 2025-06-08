//go:build linux || darwin || freebsd
// +build linux darwin freebsd

package main

var CLI struct {
	Install     struct{} `cmd:"" help:"install totaltube-frontend"`
	Start       struct{} `cmd:"" help:"Start totaltube-frontend" hidden:""`
	Child       struct{} `cmd:"" help:"Internal command to spawn the worker" hidden:""`
	Config      string   `name:"config" short:"c" type:"path" help:"location of totaltube frontend config.toml" env:"TTF_CONFIG_PATH" default:"/var/lib/totaltube-frontend/config.toml" predictor:"toml"`
	RebuildSass bool     ``
}
