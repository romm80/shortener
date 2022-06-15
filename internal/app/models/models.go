// Package models describes data models
package models

// RequestURL original link for shortening
type RequestURL struct {
	URL string `json:"url"`
}

// ResponseURL shortened link
type ResponseURL struct {
	Result string `json:"result"`
}

// RequestBatch batch request for link shortening
type RequestBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// ResponseBatch result of batch shortening links
type ResponseBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// URLsID data to write to file
type URLsID struct {
	ID          string `json:"id"`
	OriginalURL string `json:"original_url"`
}

// UserURLs shortened link query result by user id
type UserURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
