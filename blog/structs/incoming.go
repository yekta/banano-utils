package blogStructs

type SGhostPostWebhook struct {
	Post struct {
		Current struct {
			Title         string          `json:"title"`
			Html          string          `json:"html"`
			FeatureImage  string          `json:"feature_image"`
			CreatedAt     string          `json:"created_at"`
			PublishedAt   string          `json:"published_at"`
			Excerpt       string          `json:"excerpt"`
			CustomExcerpt string          `json:"custom_excerpt"`
			Slug          string          `json:"slug"`
			Id            string          `json:"id"`
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
}

type SGhostPost struct {
	Id            string          `json:"id"`
	Title         string          `json:"title"`
	Slug          string          `json:"slug"`
	PublishedAt   string          `json:"published_at"`
	Excerpt       string          `json:"excerpt"`
	CustomExcerpt string          `json:"custom_excerpt"`
	FeatureImage  string          `json:"feature_image"`
	Tags          []SGhostPostTag `json:"tags"`
	Html          string          `json:"html"`
	Mobiledoc     string          `json:"mobiledoc"`
	Plaintext     string          `json:"plaintext"`
}

type SGhostPostsForSitemapResponse struct {
	Posts []SGhostPostForSitemap `json:"posts"`
}

type SGhostPostForSitemap struct {
	Slug      string `json:"slug"`
	UpdatedAt string `json:"updated_at"`
}
