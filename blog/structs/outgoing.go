package blogStructs

type SMediumPost struct {
	Title         string   `json:"title"`
	ContentFormat string   `json:"contentFormat"`
	Content       string   `json:"content"`
	CanonicalUrl  string   `json:"canonicalUrl"`
	PublishStatus string   `json:"publishStatus"`
	Tags          []string `json:"tags"`
}

type SBlogResponse struct {
	Data  SBlogResponseData
	Error SBlogResponseError
}

type SBlogResponseData struct {
	Title string `json:"title"`
}
type SBlogResponseError struct {
	Message string `json:"message"`
}

type SBlogPostForTypesense struct {
	Id            string `json:"id"`
	Title         string `json:"title"`
	Slug          string `json:"slug"`
	PublishedAt   uint64 `json:"published_at"`
	Excerpt       string `json:"excerpt"`
	CustomExcerpt string `json:"custom_excerpt"`
	FeatureImage  string `json:"feature_image"`
	PlainText     string `json:"plaintext"`
}

type WebhookEndpoint struct {
	Name string
	Url  string
}
