package models

type RequestURL struct {
	URL string `json:"url"`
}

type ResponseURL struct {
	Result string `json:"result"`
}

type RequestBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ResponseBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type URLsID struct {
	ID          string `json:"id"`
	OriginalURL string `json:"original_url"`
}

type UserURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
