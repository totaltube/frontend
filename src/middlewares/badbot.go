package middlewares

import (
	"net/http"

	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/types"
)

func BadBotMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if internal.Config.General.CheckForBots {
			ip := r.Context().Value(types.ContextKeyIp).(string)
			if bad, err := db.CheckIfBadBot(ip, r.UserAgent()); err == nil && bad {
				isSEBot, err := db.CheckIfSeBot(ip)
				if err == nil && !isSEBot {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
