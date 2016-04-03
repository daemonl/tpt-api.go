package tptobjects

type NewsResponse struct {
	Items []NewsItem `json:"news"`
}

type NewsItem struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Summary string `json:"summary"`
}
