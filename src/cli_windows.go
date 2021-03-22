// +build windows

package main


//goland:noinspection ALL
var CLI struct {
	Install     struct{} `cmd help:"install totaltube-frontend"`
	Start       struct{} `cmd help:"Start totaltube-frontend"`
	Config      string   `name:"config" short:"c" type:"path" help:"location of totaltube frontend config.toml" env:"TTF_CONFIG_PATH" default:"./config.toml" predictor:"toml"`
	RebuildSass bool     ``
}
