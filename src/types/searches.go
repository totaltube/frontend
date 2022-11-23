package types

type Searches struct {
	Id           int64
	Message      string
	Translation  string
	Autocomplete bool
	LanguageId   string `json:"language_id"`
	Hash         string
	Total        int64
	Searches     int64
}

type TopSearch struct {
	Message  string `json:"message"`
	Searches int64  `json:"searches"`
}
