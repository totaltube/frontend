package types

type CountryGroup struct {
	Id        int64    `json:"id"`
	Name      string   `json:"name"`
	Countries []string `json:"countries"`
	Ignore    bool     `json:"ignore"`
}
