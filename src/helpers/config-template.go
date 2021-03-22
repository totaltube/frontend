package helpers

//language="Go Template"
var ConfigFileTemplate = `[general]
port = {{.General.Port}}
real_ip_header = "{{.General.RealIpHeader}}"
nginx = {{.General.Nginx}} # totaltube runs under nginx? In dev mode must be false. In production - true, this way script will avoid
              # double redirection if possible by using X-Accel-Redirect header.
use_ipv6_network = {{.General.UseIpV6Network}}
api_url = "{{.General.ApiUrl}}" # With trailing slash
api_secret = "{{.General.ApiSecret}}"
api_timeout = "{{.General.ApiTimeout}}"
lang_cookie = "{{.General.LangCookie}}"
development = {{.General.Development}}

[frontend]
sites_path = "{{.Frontend.SitesPath}}"
default_site = "{{.Frontend.DefaultSite}}"
secret_key = "{{.Frontend.SecretKey}}"
captcha_key = "{{.Frontend.CaptchaKey}}" # For DMCA
captcha_secret = "{{.Frontend.CaptchaSecret}}" # For DMCA
max_dmca_minute = {{.Frontend.MaxDmcaMinute}} # All other DMCA's from this IP will be captcha powered.
captcha_whitelist = [] # Whitelist of emails which doesn't need to be captcha checked.

[database]
path = "{{.Database.Path}}"`
