package types

type DmcaParams struct {
	ContentId       int64  `json:"content_id"`
	Email           string `json:"email"`
	Info            string `json:"info"`
	Lang            string `json:"lang"`
	Reason          string `json:"reason"`
	CaptchaResponse string `json:"h-captcha-response"`
	Country         string `json:"country"`
	UserAgent       string `json:"user_agent"`
	Domain          string `json:"domain"`
	Ip              string `json:"ip"`
}
