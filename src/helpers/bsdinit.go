package helpers

import (
	"io"
	"text/template"
)

func BSDInit(w io.Writer, path, configPath string) error {
	// language='Go template'
	t, err := template.New("init").Parse(`#!/bin/sh
#
# PROVIDE: totaltube-frontend
# REQUIRE: networking
# KEYWORD:

. /etc/rc.subr

name="totaltube-frontend"
rcvar="totaltube_enable"
totaltube_config="{{ .configPath }}"
totaltube_user="totaltube"
pidfile="{{ .path }}/totaltube-frontend.pid"
spidfile="{{ .path }}/daemon.pid"
command="/usr/local/bin/totaltube-frontend"
command_args="start -c $totaltube_config"
start_cmd="totaltube_start"
start_postcmd="totaltube_start_post"
stop_postcmd="totaltube_stop_post"
reload_postcmd="totaltube_reload_post"
sig_reload="USR2"
extra_commands="reload"
	
totaltube_start() {
	( stdbuf -oL -eL /usr/sbin/daemon -u $totaltube_user -P $spidfile $command start -c $totaltube_config ) 2>&1 | logger -t totaltube-frontend &
}
	
totaltube_stop_post() {
  if [ -e "${spidfile}" ]; then
	  kill -9 $(cat $spidfile)
  fi
}
	
totaltube_start_post() {
  sleep 1
  echo "Totaltube Frontend started. PID: $(cat $pidfile)"
}

totaltube_reload_post() {
  sleep 1
  echo "Totaltube Frontend reloaded. PID: $(cat $pidfile)"
}
	
load_rc_config $name
: ${totaltube_enable:=no}
	
run_rc_command "$1"`)
	if err != nil {
		return err
	}
	err = t.Execute(w, map[string]interface{}{"path": path, "configPath": configPath})
	return err
}

