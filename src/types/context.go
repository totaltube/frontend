package types

type ContextKey string

const (
	ContextKeyLang               ContextKey = "lang"
	ContextKeyIsXDefault         ContextKey = "isXDefault"
	ContextKeyConfig             ContextKey = "config"
	ContextKeyPath               ContextKey = "path"
	ContextKeyHostName           ContextKey = "hostName"
	ContextKeyCustomTemplateName ContextKey = "custom_template_name"
	ContextKeyIp                 ContextKey = "ip"
)
