package helpers

import (
	"io"
	"text/template"
)

func SystemdInit(w io.Writer, path, configPath string) error {
	t, err := template.New("init").
		// language="Unit File (systemd)"
		Parse(`[Unit]
Description=Totaltube Frontend daemon
Documentation=https://totaltraffictrader.com/
After=network.target

[Service]
Type=simple
User=totaltube
Restart=always
RestartSec=5s
LimitNOFILE=infinity
LimitNPROC=infinity
LimitCORE=infinity
PIDFile={{ .path }}/totaltube-frontend.pid
WorkingDirectory={{ .path }}

ExecStart=/usr/local/bin/totaltube-frontend start -c {{ .configPath }}
ExecReload=/bin/kill -s USR2 $MAINPID
ExecStop=/bin/kill -s TERM $MAINPID

[Install]
WantedBy=multi-user.target
`)
	if err != nil {
		return err
	}
	err = t.Execute(w, map[string]interface{}{"path": path, "configPath": configPath})
	return err
}
