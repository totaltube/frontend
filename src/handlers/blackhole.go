package handlers

import (
	"log"
	"net/http"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/types"
)

var Blackhole = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	hostName := r.Context().Value(types.ContextKeyHostName).(string)
	ip := r.Context().Value(types.ContextKeyIp).(string)
	userAgent := r.Header.Get("User-Agent")
	referer := r.Header.Get("Referer")
	http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	isSeBot, err := db.CheckIfSeBot(ip)
	if err != nil {
		log.Println(err)
		return
	}
	if isSeBot {
		return
	}
	err = api.BadbotRegister(hostName, ip, userAgent, referer)
	if err != nil {
		log.Println(err)
	}
})
