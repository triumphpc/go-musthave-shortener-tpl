package shortlink

// ShortLinks maps original and shorts
type ShortLinks map[Short]string

// Short link
type Short string

// ShortURLs shorts from mass save
type ShortURLs struct {
	Short string `json:"short_url"`
	ID    string `json:"correlation_id"`
}

// URL it's users full url
type URL struct {
	URL string `json:"url"`
}

// URLs from mass save
type URLs struct {
	ID     string `json:"correlation_id"`
	Origin string `json:"original_url"`
}
