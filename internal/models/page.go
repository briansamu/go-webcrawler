package models

type Page struct {
	Url     string  `json:"url"`
	Title   string  `json:"title"`
	Content string  `json:"content"`
	Score   float64 `json:"score,omitempty"` // Search relevance score
}
