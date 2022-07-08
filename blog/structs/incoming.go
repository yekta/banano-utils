package blogStructs

type SGhostPostWebhook struct {
	Post struct {
		Current struct {
			Id            string          `json:"id"`
			Title         string          `json:"title"`
			Html          string          `json:"html"`
			FeatureImage  string          `json:"feature_image"`
			CreatedAt     string          `json:"created_at"`
			PublishedAt   string          `json:"published_at"`
			Excerpt       string          `json:"excerpt"`
			CustomExcerpt string          `json:"custom_excerpt"`
			Slug          string          `json:"slug"`
			Tags          []SGhostPostTag `json:"tags"`
		} `json:"current"`
	} `json:"post"`
}

type SGhostPostTag struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type SGhostPostsResponse struct {
	Posts []SGhostPost `json:"posts"`
	Meta  SGhostMeta   `json:"meta"`
}

type SGhostPost struct {
	Id            string          `json:"id,omitempty"`
	Title         string          `json:"title,omitempty"`
	Slug          string          `json:"slug,omitempty"`
	CreatedAt     string          `json:"created_at,omitempty"`
	UpdatedAt     string          `json:"updated_at,omitempty"`
	PublishedAt   string          `json:"published_at,omitempty"`
	Excerpt       string          `json:"excerpt,omitempty"`
	CustomExcerpt string          `json:"custom_excerpt,omitempty"`
	FeatureImage  string          `json:"feature_image,omitempty"`
	Tags          []SGhostPostTag `json:"tags,omitempty"`
	Html          string          `json:"html,omitempty"`
	Plaintext     string          `json:"plaintext,omitempty"`
	Featured      bool            `json:"featured,omitempty"`
	ReadingTime   int             `json:"reading_time,omitempty"`
}

type SGhostPostsForSitemapResponse struct {
	Posts []SGhostPostForSitemap `json:"posts"`
}

type SGhostPostForSitemap struct {
	Slug      string `json:"slug"`
	UpdatedAt string `json:"updated_at"`
}

type SGhostMeta struct {
	Pagination SGhostPagination `json:"pagination"`
}

type SGhostPagination struct {
	Page  int `json:"page,omitempty"`
	Pages int `json:"pages,omitempty"`
	Limit int `json:"limit,omitempty"`
	Total int `json:"total,omitempty"`
	Next  int `json:"next,omitempty"`
	Prev  int `json:"prev,omitempty"`
}
