package types

type DmcaParams struct {
	ContentId int64  `json:"content_id"`
	Email     string `json:"email"`
	Info      string `json:"info"`
	Lang      string `json:"lang"`
	Reason    string `json:"reason"`
}
