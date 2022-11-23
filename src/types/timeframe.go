package types

type Timeframe struct {
	Id      int32  `json:"id"`
	Name    string `json:"name"`
	Seconds int64  `json:"seconds"`
}
