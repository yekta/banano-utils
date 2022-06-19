package blogStructs

type SMediumPost struct {
	Title         string `json:"title"`
	ContentFormat string `json:"contentFormat"`
	Content       string `json:"content"`
	CanonicalUrl  string `json:"canonicalUrl"`
	PublishStatus string `json:"publishStatus"`
	/* 	Tags          []string `json:"tags"` */
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
