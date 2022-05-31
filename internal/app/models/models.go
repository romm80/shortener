package models

// RequestURL содержит ссылку для сокращения
type RequestURL struct {
	URL string `json:"url"`
}

// ResponseURL содержит сокращенную ссылку
type ResponseURL struct {
	Result string `json:"result"`
}

// RequestBatch содержит запррос для пакетного сокращения ссылок
type RequestBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// ResponseBatch содержит результат пакетного сокращения ссылок
type ResponseBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// URLsID содержит данные для записи в файл
type URLsID struct {
	ID          string `json:"id"`
	OriginalURL string `json:"original_url"`
}

// UserURLs содержит результат запроса сокращенных ссылок по id пользователя
type UserURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
